package node

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pericles-tpt/ownapi/functions"
	"github.com/pericles-tpt/ownapi/utility"
	"github.com/pkg/errors"
)

type CustomNodeConfig struct {
	BaseNodeProps

	Name string `json:"name"`

	InputKeys  []string `json:"input_keys"`
	OutputKeys []string `json:"output_keys"`
}

type CustomNode struct {
	Config CustomNodeConfig `json:"config"`
}

func CreateCustomNode(propMap map[string]any, cfg CustomNodeConfig) (CustomNode, error) {
	var ret CustomNode

	err := utility.ErrOnAnyMatch([][]string{cfg.InputKeys, cfg.OutputKeys}, []string{"empty keys provided for InputKeys at indices", "empty keys provided for OutputKeys at indices"}, "")
	if err != nil {
		return ret, err
	}

	// TODO: Review if this is necessary, doesn't seem that useful in this case...
	cfg, err = utility.OverrideTypeFromJSONMap(cfg, propMap)
	if err != nil {
		return ret, errors.Wrap(err, "failed to override `cfg` with map values")
	}
	ret.Config = cfg

	cf, err := functions.GetFunc(cfg.Name)
	if err != nil {
		return ret, errors.Wrapf(err, "failed to retrieve func with name '%s'", cfg.Name)
	}

	if len(cfg.InputKeys) != len(cf.SigParams.Names) {
		return ret, fmt.Errorf("invalid num of input keys doesn't match num of function parameters, %d != %d", len(cfg.InputKeys), len(cfg.InputKeys))
	}

	numReturnTypesMinusError := len(cf.SigReturnTypes) - 1
	if len(cfg.OutputKeys) != numReturnTypesMinusError {
		return ret, fmt.Errorf("invalid num of outut keys doesn't match num of return types (excluding error), %d != %d", len(cfg.OutputKeys), numReturnTypesMinusError)
	}

	err = ret.regenerateHash()
	if err != nil {
		return ret, errors.Wrap(err, "failed to generate hash for new `HttpNodeConfig`")
	}

	return ret, nil
}

func (cn *CustomNode) Trigger(propMap map[string]any, useCache bool) (map[string]any, error) {
	var outputMap = map[string]any{}
	newCfg, err := utility.OverrideTypeFromJSONMap(cn.Config, propMap)
	if err != nil {
		return outputMap, errors.Wrap(err, "failed to override `cfg` with map values")
	}
	defer func(cn *CustomNode, oldCfg CustomNodeConfig) {
		cn.Config = oldCfg
	}(cn, cn.Config)
	cn.Config = newCfg

	f, err := functions.GetFunc(cn.Config.Name)
	if err != nil {
		return outputMap, errors.Wrapf(err, "failed to retrieve func '%s'", cn.Config.Name)
	}

	params := make([]any, 0, len(cn.Config.InputKeys))
	for _, key := range cn.Config.InputKeys {
		if !strings.HasPrefix(key, "input:") {
			key = fmt.Sprintf("input:%s", key)
		}
		params = append(params, propMap[key])
	}
	res, err := f.F(params)
	if err != nil {
		return outputMap, errors.Wrapf(err, "failed to execute func '%s'", cn.Config.Name)
	}

	for i, val := range res {
		key := cn.Config.OutputKeys[i]
		if !strings.HasPrefix(key, "output:") {
			key = fmt.Sprintf("output:%s", key)
		}
		outputMap[key] = val
	}

	return outputMap, nil
}

func (cn *CustomNode) regenerateHash() error {
	copyForHash := CustomNode{}
	copyForHash = *cn
	copyForHash.Config.Hash = ""

	nodeBytes, err := json.Marshal(copyForHash)
	if err != nil {
		return errors.Wrap(err, "failed to marshal node to bytes")
	}

	hash := sha1.New()
	_, err = hash.Write(nodeBytes)
	if err != nil {
		return errors.Wrap(err, "failed to write bytes to hash")
	}
	newHashBytes := hash.Sum(nil)

	cn.Config.Hash = fmt.Sprintf("%x", newHashBytes)
	return nil
}

func (cn *CustomNode) readCachedResponseData() *[]byte {
	return nil
}
func (cn *CustomNode) writeCachedResponseData(data []byte) {
}
func (cn *CustomNode) Changed(propsMap map[string]any) bool {
	return true
}
func (cn *CustomNode) revert(changed *bool, propsMap map[string]any) {
}
func (cn *CustomNode) GetTrigger() *Trigger {
	return cn.Config.NodeTrigger
}
