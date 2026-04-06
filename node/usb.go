package node

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"os"

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
	MassStorage USBProtocol = iota
	MTP

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
		return ret, errors.Wrapf(err, "failed to verify access to USB device - serial: %s, protocol: %v, root_scope: %s", ret.Config.SerialNo, ret.Config.Protocol, ret.Config.RootScope)
	}

	err = ret.generateNewHash()
	if err != nil {
		return ret, errors.Wrap(err, "failed to generate hash for new `HttpNodeConfig`")
	}

	return ret, nil
}

func (un *USBNode) verifyAccess() error {
	// TODO: Check that:
	// 1. Device with serial number exists
	// ctx := gousb.NewContext()
	// defer ctx.Close()

	// devs, err := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
	// 	return true
	// })
	// if err != nil {
	// 	return errors.Wrapf(err, "failed to retrieve usb devices")
	// }

	// var targetDev *gousb.Device
	// for _, d := range devs {
	// 	sn, err := d.SerialNumber()
	// 	if err == nil && sn == un.Config.SerialNo {
	// 		targetDev = d
	// 		break
	// 	}
	// }
	// if targetDev == nil {
	// 	// TODO: Improve this, the loop above could catch errors but idk which error will correspond to the device
	// 	// 		 with the target serial number
	// 	return fmt.Errorf("failed to find or access device matching serial number: %s", un.Config.SerialNo)
	// }

	// 2. It can be accessed over protocol

	// 3. The scope directory can be accessed
	// 4. Recv permissions are valid
	// 5. Send permissions are valid

	return errors.New("UNIMPLEMENTED")
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
