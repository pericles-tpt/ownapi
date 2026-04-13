package node

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"os"

	gomtp "github.com/pericles-tpt/go-mtp"
	"github.com/pericles-tpt/ownapi/utility"
	"github.com/pkg/errors"
)

type USBNodeConfig struct {
	SerialNo  string      `json:"serial_no"`
	RootScope string      `json:"root_scope"`
	OutputDir string      `json:"output_dir"`
	Protocol  USBProtocol `json:"protocol"`
	RecvPerms uint8       `json:"recv_perms"`
	SendPerms uint8       `json:"send_perms"`
}

type USBNode struct {
	Hash   string        `json:"hash"`
	Config USBNodeConfig `json:"config"`
}

type USBProtocol int

const (
	MTP USBProtocol = iota
	// MassStorage

	// Permissions
	PERM_GET = 1
	PERM_ADD = 2
	PERM_REM = 5
	PERM_MOD = 7
)

func CreateUSBNode(propMap map[string]any, cfg USBNodeConfig) (USBNode, error) {
	var ret USBNode

	cfg, err := utility.OverrideTypeFromJSONMap(cfg, propMap)
	if err != nil {
		return ret, errors.Wrap(err, "failed to override `cfg` with map values")
	}
	ret.Config = cfg

	err = ret.verifyAccess()
	if err != nil {
		return ret, errors.Wrap(err, "failed to verify access to USB device")
	}

	err = ret.generateNewHash()
	if err != nil {
		return ret, errors.Wrap(err, "failed to generate hash for new `USBNodeConfig`")
	}

	return ret, nil
}

func (un *USBNode) verifyAccess() error {
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

func (un *USBNode) generateNewHash() error {
	// Remove cache file for old file
	if un.Hash != "" {
		cachedFilePath := fmt.Sprintf("%s/%s", usbResponseCacheOutputPath, un.Hash)
		err := os.Remove(cachedFilePath)
		if err != nil {
			return err
		}
	}

	copyForHash := USBNode{}
	copyForHash = *un
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

	un.Hash = fmt.Sprintf("%x", newHashBytes)
	return nil
}

func (un *USBNode) Transfer(src string, dest string) error {
	destInData := fmt.Sprintf("./_data/files/%s", dest)
	err := os.Mkdir(destInData, 0760)
	if err != nil {
		if !os.IsExist(err) {
			return errors.Wrap(err, "failed to create dest folder")
		} else {
			// TODO: Remove this later
			os.RemoveAll(destInData)
			err = os.Mkdir(destInData, 0760)
			if err != nil {
				return errors.Wrap(err, "failed to create dest folder")
			}
		}
	}

	switch un.Config.Protocol {
	case MTP:
		dev, err := gomtp.GetDeviceBySerialNumber(un.Config.SerialNo)
		if err != nil {
			return errors.Wrapf(err, "failed to retrieve device matching serial number: %s", un.Config.SerialNo)
		}
		defer dev.ReleaseDevice()

		err = dev.GetStorage()
		if err != nil {
			return errors.Wrapf(err, "failed to get storage for mtp device")
		}

		files, _, err := dev.GetFilesAndFolders(dev.Storage[0].Id, 0)
		if err != nil {
			return errors.Wrapf(err, "failed to get files and folders for first storage id")
		}

		for _, f := range files {
			// FIX [GARMIN SPECIFIC]: For some reason a my Forerunner it doesn't label files as anything but 44 (I think unknown?)
			// 						  one way to guess it's a file is if it has an extension
			probablyAFile := f.Filesize > 0
			if probablyAFile {
				fmt.Printf("Copying: %s, size: %d\n", f.Filename, f.Filesize)
				localPath := fmt.Sprintf("%s/%s", destInData, f.Filename)
				err = dev.GetFileToFile(f.ItemId, localPath)
				if err != nil {
					return errors.Wrapf(err, "failed to copy '%s' to local directory", f.Filename)
				}
			}
		}
	default:
		return fmt.Errorf("no `Transfer` method provided for protocol: %v", un.Config.Protocol)
	}
	return nil
}
