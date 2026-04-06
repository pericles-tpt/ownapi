package node

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"os"

	"github.com/pericles-tpt/ownapi2/utility"
	"github.com/pkg/errors"
)

type JSONNodeConfig struct {
	InputKey     string             `json:"input_key"`
	ExtractNodes []utility.JSONProp `json:"extract_nodes"`
}

type JSONNode struct {
	Hash   string         `json:"hash"`
	Config JSONNodeConfig `json:"config"`
}

func CreateJSONNode(propMap map[string]any, cfg JSONNodeConfig) (JSONNode, error) {
	var ret JSONNode

	cfg, err := utility.OverrideTypeFromJSONMap(cfg, propMap)
	if err != nil {
		return ret, errors.Wrap(err, "failed to override `cfg` with map values")
	}
	ret = JSONNode{
		Config: cfg,
	}

	err = ret.generateNewHash()
	if err != nil {
		return ret, errors.Wrap(err, "failed to generate hash for new `httpBaseNode`")
	}

	return ret, nil
}

func (jn *JSONNode) triggerNoCache(propMap map[string]any) (map[string]any, error) {
	var outputMap = map[string]any{}
	if jn.Config.ExtractNodes == nil {
		return outputMap, errors.New("unable to extract data from JSON response, `ExtractNodes` not defined")
	}

	input, err := jn.maybeConsumeInput(propMap)
	if err != nil {
		return outputMap, err
	}

	var anyJson any
	err = json.Unmarshal(input, &anyJson)
	if err != nil {
		return outputMap, errors.Wrap(err, "failed to unmarshal response body as expected JSON")
	}

	vals, err := utility.TraverseJSONExtractValues(anyJson, jn.Config.ExtractNodes)
	if err != nil {
		return outputMap, errors.Wrap(err, "failed to extract JSON values")
	}

	for k, v := range vals {
		outputMap[fmt.Sprintf("output:%s", k)] = v
	}

	return outputMap, nil
}

func (jn *JSONNode) Trigger(propMap map[string]any) (map[string]any, error) {
	var outputMap = map[string]any{}

	if jn.Config.ExtractNodes == nil {
		return outputMap, errors.New("unable to extract data from JSON response, `ExtractNodes` not defined")
	}

	input, err := jn.maybeConsumeInput(propMap)
	if err != nil {
		return outputMap, err
	}

	var anyJson any
	err = json.Unmarshal(input, &anyJson)
	if err != nil {
		return outputMap, errors.Wrap(err, "failed to unmarshal response body as expected JSON")
	}

	vals, err := utility.TraverseJSONExtractValues(anyJson, jn.Config.ExtractNodes)
	if err != nil {
		return outputMap, errors.Wrap(err, "failed to extract JSON values")
	}

	for k, v := range vals {
		outputMap[fmt.Sprintf("output:%s", k)] = v
	}

	data, err := json.Marshal(vals)
	if err != nil {
		return outputMap, errors.Wrap(err, "failed to marshal extracted JSON values to bytes")
	}
	path := fmt.Sprintf("%s/%s", jsonResponseCacheOutputPath, jn.Hash)
	err = os.WriteFile(path, data, 0660)
	if err != nil {
		return outputMap, errors.Wrap(err, "failed to write output of JSON HTTP request to file")
	}

	return outputMap, nil
}

func (jn *JSONNode) generateNewHash() error {
	// Remove cache file for old file
	if jn.Hash != "" {
		cachedFilePath := fmt.Sprintf("%s/%s", jsonResponseCacheOutputPath, jn.Hash)
		err := os.Remove(cachedFilePath)
		if err != nil {
			return err
		}
	}

	copyForHash := JSONNode{}
	copyForHash = *jn
	copyForHash.Hash = ""

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

	jn.Hash = fmt.Sprintf("%x", newHashBytes)
	return nil
}

func (hn *JSONNode) readCachedResponseData() *[]byte {
	cachedFilePath := fmt.Sprintf("%s/%s", jsonResponseCacheOutputPath, hn.Hash)
	data, err := os.ReadFile(cachedFilePath)
	if err != nil {
		return nil
	}
	return &data
}

func (hn *JSONNode) writeCachedResponseData(data []byte) {
	cachedFilePath := fmt.Sprintf("%s/%s", jsonResponseCacheOutputPath, hn.Hash)
	os.Remove(cachedFilePath)

	err := os.WriteFile(cachedFilePath, data, 0660)
	if err != nil {
		fmt.Println("Failed to write file: ", err)
	}
}

func (jn *JSONNode) maybeConsumeInput(propMap map[string]any) ([]byte, error) {
	var (
		maybeInput any
		input      []byte
		ok         bool
	)
	if maybeInput, ok = propMap[jn.Config.InputKey]; !ok {
		return input, fmt.Errorf("failed to find input at key '%s' in propMap", jn.Config.InputKey)
	}
	if input, ok = maybeInput.([]byte); !ok {
		return input, fmt.Errorf("input at key '%s' in propMap is not a []byte", jn.Config.InputKey)
	}
	return input, nil
}
