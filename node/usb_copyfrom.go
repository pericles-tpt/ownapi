package node

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	gomtp "github.com/pericles-tpt/go-mtp"
	"github.com/pericles-tpt/ownapi/utility"
	"github.com/pkg/errors"
)

var usbResponseCacheOutputPath string

type USBCopyFromNodeConfig struct {
	BaseNodeProps

	SerialNo string      `json:"serial_no"`
	Protocol USBProtocol `json:"protocol"`
	Perms    uint8       `json:"perms"`
	SrcPath  string      `json:"src_path"`
}

type USBCopyFromNode struct {
	Config USBCopyFromNodeConfig `json:"config"`
}

type USBProtocol int

const (
	MTP USBProtocol = iota
	// MassStorage

	// Permissions
	PERM_ADD = 1
	PERM_REM = 2 // TODO:
	PERM_MOD = 4
)

func CreateUSBCopyFromNode(propMap map[string]any, cfg USBCopyFromNodeConfig, reload bool) (USBCopyFromNode, error) {
	var ret USBCopyFromNode

	cfg, err := utility.OverrideTypeFromJSONMap(cfg, propMap)
	if err != nil {
		return ret, errors.Wrap(err, "failed to override `cfg` with map values")
	}
	ret.Config = cfg

	if !reload {
		err = ret.verifyAccess()
		if err != nil {
			return ret, errors.Wrap(err, "failed to verify access to USB device")
		}
	}

	err = ret.regenerateHash()
	if err != nil {
		return ret, errors.Wrap(err, "failed to generate hash for new `USBCopyFromNodeConfig`")
	}

	return ret, nil
}

func (un *USBCopyFromNode) verifyAccess() error {
	switch un.Config.Protocol {
	case MTP:
		dev, err := gomtp.GetDeviceBySerialNumber(un.Config.SerialNo)
		if err != nil {
			return errors.Wrapf(err, "failed to access device over MTP with serial no: %s", un.Config.SerialNo)
		}
		dev.ReleaseDevice()
	default:
		return fmt.Errorf("failed to verify access for USB protocol: %v", un.Config.Protocol)
	}
	return nil
}

func (un *USBCopyFromNode) regenerateHash() error {
	// Remove cache file for old file
	if un.Config.Hash != "" {
		cachedFilePath := fmt.Sprintf("%s/%s", usbResponseCacheOutputPath, un.Config.Hash)
		err := os.Remove(cachedFilePath)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	copyForHash := USBCopyFromNode{}
	copyForHash = *un
	copyForHash.Config.Hash = ""
	copyForHash.Config.NodeTrigger = nil
	// TODO: Should this be excluded from the hash?
	copyForHash.Config.Perms = 0

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

	un.Config.Hash = fmt.Sprintf("%x", newHashBytes)
	return nil
}

func (un *USBCopyFromNode) Trigger(propMap map[string]any) (map[string]any, error) {
	destInData := fmt.Sprintf("./_data/files/%s", un.Config.Hash)
	_, err := os.Stat(destInData)
	if err != nil {
		if !os.IsNotExist(err) {
			return propMap, errors.Wrap(err, "failed to create dest folder")
		} else {
			err = os.Mkdir(destInData, 0760)
			if err != nil {
				return propMap, errors.Wrap(err, "failed to create dest folder")
			}
		}
	}

	switch un.Config.Protocol {
	case MTP:
		dev, err := gomtp.GetDeviceBySerialNumber(un.Config.SerialNo)
		if err != nil {
			return propMap, errors.Wrapf(err, "failed to retrieve device matching serial number: %s", un.Config.SerialNo)
		}
		defer dev.ReleaseDevice()

		err = dev.GetStorage()
		if err != nil {
			return propMap, errors.Wrapf(err, "failed to get storage for mtp device")
		}

		// TODO: Review this, could be more than one for other MTP devices, just using this for my Garmin
		files, _, err := dev.GetFilesAndFolders(dev.Storage[0].Id, 0)
		if err != nil {
			return propMap, errors.Wrapf(err, "failed to get files and folders for first storage id")
		}

		var maybeParentChildIdMap = map[uint32]struct {
			Basename string
			ModTime  time.Time
			Children []uint32
			ZeroSize bool
		}{}
		var nodeQ = []struct {
			Id       uint32
			ModTime  time.Time
			Path     []string
			ZeroSize bool
		}{}
		for _, f := range files {
			hasParent := f.ParentId > 0
			if hasParent {
				var (
					maybeParentNode struct {
						Basename string
						ModTime  time.Time
						Children []uint32
						ZeroSize bool
					}
					ok bool
				)
				if maybeParentNode, ok = maybeParentChildIdMap[f.ParentId]; ok {
					maybeParentNode.Children = append(maybeParentNode.Children, f.ItemId)
				}
				maybeParentChildIdMap[f.ParentId] = maybeParentNode
			} else {
				nodeQ = append(nodeQ, struct {
					Id       uint32
					ModTime  time.Time
					Path     []string
					ZeroSize bool
				}{
					Id:       f.ItemId,
					ModTime:  f.ModificationDate,
					Path:     []string{destInData, f.Filename},
					ZeroSize: f.Filesize == 0,
				})
			}

			var (
				thisNode = struct {
					Basename string
					ModTime  time.Time
					Children []uint32
					ZeroSize bool
				}{
					Basename: f.Filename,
					ModTime:  f.ModificationDate,
					ZeroSize: f.Filesize == 0,
				}
			)
			// A child might initialise the parent WITHOUT a basename, this accounts for that
			if maybeThisNode, ok := maybeParentChildIdMap[f.ItemId]; ok {
				thisNode.Children = maybeThisNode.Children
			}
			maybeParentChildIdMap[f.ItemId] = thisNode
		}

		copiedFilePaths := make([]string, 0)
		for i := 0; i < len(nodeQ); i++ {
			n := nodeQ[i]
			basenameAndChildren := maybeParentChildIdMap[n.Id]

			localPath := strings.Join(n.Path, "/")

			fs, err := os.Stat(localPath)
			if err != nil && !os.IsNotExist(err) {
				return propMap, errors.Wrap(err, "unexpected stat error")
			}
			exists := err == nil

			var copied bool

			// ASSUMPTION: Empty file entries are folders, they could be empty files but that doesn't work with GetFileToFile
			isFolder := len(basenameAndChildren.Children) > 0 || basenameAndChildren.ZeroSize
			if isFolder {
				if !exists {
					err = os.Mkdir(localPath, 0760)
					if err != nil {
						return propMap, errors.Wrapf(err, "failed to create directory: %s", localPath)
					}
				}
				// TODO: How to handle queue where items could be [blah/, ha/, something/, foo/bar/]. Can't keep appending names because it'll nest them incorrectly...

				basenames := make([][]string, 0, len(basenameAndChildren.Children))

				// Appending like this could blow out memory?
				for _, c := range basenameAndChildren.Children {
					basenameAndChildren := maybeParentChildIdMap[c]

					childPath := make([]string, len(n.Path))
					copy(childPath, n.Path)
					childPath = append(childPath, basenameAndChildren.Basename)

					nodeQ = append(nodeQ, struct {
						Id       uint32
						ModTime  time.Time
						Path     []string
						ZeroSize bool
					}{
						Id:       c,
						ModTime:  basenameAndChildren.ModTime,
						Path:     childPath,
						ZeroSize: basenameAndChildren.ZeroSize,
					})

					basenames = append(basenames, childPath)
				}
			} else if !exists {
				err = dev.GetFileToFile(n.Id, localPath)
				if err != nil {
					basename := strings.Split(localPath, "/")[len(localPath)-1]
					return propMap, errors.Wrapf(err, "failed to copy '%s' to local directory", basename)
				}
				copied = true
			} else { // exists
				localModTime := time.Time{}
				if fs != nil {
					localModTime = fs.ModTime()
				}
				deviceFileNewer := n.ModTime.After(localModTime)
				if un.Config.Perms&PERM_MOD > 0 && deviceFileNewer {
					// Rename to tmp file in case something goes wrong
					tmpName := fmt.Sprintf("%s.tmp", localPath)
					err = os.Rename(localPath, tmpName)
					if err != nil {
						return propMap, errors.Wrapf(err, "failed to rename original file: '%s' to tmp file before MODIFY", localPath)
					}
					// Copy file to localpath
					err = dev.GetFileToFile(n.Id, localPath)
					if err != nil {
						err = os.Rename(tmpName, localPath)
						if err != nil {
							return propMap, errors.Wrapf(err, "failed to revert tmp file: '%s' to original file after failed MODIFY", tmpName)
						}
						basename := strings.Split(localPath, "/")[len(localPath)-1]
						return propMap, errors.Wrapf(err, "failed to copy '%s' to local directory", basename)
					}
					copied = true
					// Remove tmp file after successful copy
					err = os.Remove(tmpName)
					if err != nil {
						fmt.Println("WARN: Failed to remove tmp file: ", tmpName)
					}
				}
			}

			if copied {
				copiedFilePaths = append(copiedFilePaths, localPath)
			}
		}

		propMap["output:copied_files"] = copiedFilePaths
		return propMap, nil
	default:
		return propMap, fmt.Errorf("no `Transfer` method provided for protocol: %v", un.Config.Protocol)
	}
	return propMap, nil
}

func (un *USBCopyFromNode) Changed(propMap map[string]any) bool {
	var changed bool
	defer un.revert(&changed, propMap)

	if cf, ok := propMap["output:copied_files"]; ok {
		if files, ok := cf.([]string); ok && len(files) > 0 {
			changed = true
		}
	}
	return changed
}
func (un *USBCopyFromNode) revert(changed *bool, propMap map[string]any) {
	if *changed {
		return
	}

	defer delete(propMap, "output:copied_files")
	if cf, ok := propMap["output:copied_files"]; ok {
		if files, ok := cf.([]string); ok {
			for _, f := range files {
				os.Remove(f)
			}
		}
	}
}

// TODO: No-ops, review
func (un *USBCopyFromNode) readCachedResponseData() *[]byte {
	return nil
}
func (un *USBCopyFromNode) writeCachedResponseData(data []byte) {
}
func (un *USBCopyFromNode) triggerNoCache(propMap map[string]any) (map[string]any, error) {
	return propMap, nil
}
func (un *USBCopyFromNode) GetTrigger() *Trigger {
	return un.Config.NodeTrigger
}
