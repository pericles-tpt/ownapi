package node

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/pericles-tpt/ownapi/binary"
	"github.com/pericles-tpt/ownapi/utility"
	"github.com/pkg/errors"
)

type BinaryNodeConfig struct {
	BaseNodeProps

	Name string `bson:"name" json:"name"`

	BinaryName string   `bson:"binary_name" json:"binary_name"`
	Params     []string `bson:"params" json:"params"`

	// TODO: Outputs
}

type BinaryNode struct {
	Config BinaryNodeConfig `bson:"config" json:"config"`
}

func CreateBinaryNode(propMap map[string]any, cfg BinaryNodeConfig) (BinaryNode, error) {
	var ret BinaryNode

	if !binary.Exists(cfg.BinaryName) {
		return ret, fmt.Errorf("failed to find verified binary matching name: %s", cfg.BinaryName)
	}

	if len(cfg.Params) == 0 {
		return ret, errors.New("no params provided")
	}

	// TODO: Review if this is necessary, doesn't seem that useful in this case...
	cfg, err := utility.OverrideTypeFromJSONMap(cfg, propMap)
	if err != nil {
		return ret, errors.Wrap(err, "failed to override `cfg` with map values")
	}
	ret.Config = cfg

	err = ret.regenerateHash()
	if err != nil {
		return ret, errors.Wrap(err, "failed to generate hash for new `HttpNodeConfig`")
	}

	return ret, nil
}

func (cn *BinaryNode) Trigger(propMap map[string]any, useCache bool) (map[string]any, error) {
	var outputMap = map[string]any{}
	newCfg, err := utility.OverrideTypeFromJSONMap(cn.Config, propMap)
	if err != nil {
		return outputMap, errors.Wrap(err, "failed to override `cfg` with map values")
	}
	defer func(cn *BinaryNode, oldCfg BinaryNodeConfig) {
		cn.Config = oldCfg
	}(cn, cn.Config)
	cn.Config = newCfg

	fmt.Println("new cfg: ", spew.Sdump(newCfg))

	err = binary.Run(cn.Config.BinaryName, cn.Config.Params)
	if err != nil {
		return propMap, errors.Wrap(err, "failed to run")
	}

	return outputMap, nil
}

func (cn *BinaryNode) regenerateHash() error {
	copyForHash := BinaryNode{}
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

func (cn *BinaryNode) readCachedResponseData() *[]byte {
	return nil
}
func (cn *BinaryNode) writeCachedResponseData(data []byte) {
}
func (cn *BinaryNode) Changed(propsMap map[string]any) bool {
	return true
}
func (cn *BinaryNode) revert(changed *bool, propsMap map[string]any) {
}
func (cn *BinaryNode) GetTrigger() *Trigger {
	return cn.Config.NodeTrigger
}
