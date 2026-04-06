package node

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pericles-tpt/ownapi/setup"
	"github.com/pkg/errors"
)

type BaseNode interface {
	Trigger(propMap map[string]any) (map[string]any, error)
	triggerNoCache(propMap map[string]any) (map[string]any, error)
	generateNewHash() error
	readCachedResponseData() *[]byte
	writeCachedResponseData(data []byte)
}

var _ BaseNode = (*HTTPNode)(nil)
var _ BaseNode = (*JSONNode)(nil)

type NodeType int

const (
	Http NodeType = iota
	Json
)

// var _ BaseNode = (*JSONNode)(nil)

func Init() error {
	var err error
	httpResponseCacheOutputPath, _, err = setup.GetDirectoryPath("http")
	if err != nil {
		return errors.Wrap(err, "failed to init `httpResponseCacheOutputPath`")
	}
	jsonResponseCacheOutputPath, _, err = setup.GetDirectoryPath("json")
	if err != nil {
		return errors.Wrap(err, "failed to init `jsonResponseCacheOutputPath`")
	}
	usbResponseCacheOutputPath, _, err = setup.GetDirectoryPath("usb")
	if err != nil {
		return errors.Wrap(err, "failed to init `usbResponseCacheOutputPath`")
	}
	return nil
}

// Utilities
// KEY RULES (applied at the END of a pipeline stage, in this order):
//  1. Keys can have 2-3 parts, e.g. input:foo OR input:n:foo
//  2. 2-part keys starting with 'input' are removed from the map
//  3. Any keys starting with 'output' become 'input'
//  4. 3-part keys, the 2nd part MUST be an integer
//  5. 3-part keys become 2-part keys when `pipelineStageCount + 1` matches their integer part
func UpdateKeys(propMap map[string]any, pipelineStageCount int) (map[string]any, error) {
	newPropMap := make(map[string]any, len(propMap))

	for k, v := range propMap {
		keyParts := strings.Split(k, ":")
		isOutput := keyParts[0] == "output"

		switch len(keyParts) {
		case 2:
			if isOutput {
				newKey := fmt.Sprintf("input:%s", keyParts[1])
				newPropMap[newKey] = v
			}
			// INPUT -> DROP
		case 3:
			useAtStage, err := strconv.ParseInt(keyParts[1], 10, 64)
			if err != nil {
				return propMap, fmt.Errorf("invalid key with 3 parts found, but 2nd part is NOT an integer - %s", k)
			}
			useAtNextPipelineStage := int(useAtStage) == (pipelineStageCount + 1)
			maybeNewKey := fmt.Sprintf("input:%d:%s", useAtStage, keyParts[2])
			if useAtNextPipelineStage {
				maybeNewKey = fmt.Sprintf("input:%s", keyParts[2])
			}
			newPropMap[maybeNewKey] = v
		default:
			return propMap, fmt.Errorf("invalid output key with invalid number of parts - %s", k)
		}
	}
	return newPropMap, nil
}
