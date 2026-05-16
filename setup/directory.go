package setup

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
)

var (
	// e.g. _data/input/http
	dataRoot = "_data"
	stages   = []string{"output"}
	nodeMap  = map[string][]string{
		"http": {},
		"json": {},
		"usb":  {"copy"},
	}
)

func MakeDirectories() error {
	// Make 'files' directory
	path := fmt.Sprintf("%s/files", dataRoot)
	err := os.Mkdir(path, 0760)
	if err != nil && !os.IsExist(err) {
		return errors.Wrap(err, "failed to create data directory for files")
	}

	for _, stage := range stages {
		for rootNode, leaves := range nodeMap {
			if len(leaves) == 0 {
				path := fmt.Sprintf("%s/%s/%s", dataRoot, stage, rootNode)
				err := os.MkdirAll(path, 0760)
				if err != nil && !os.IsExist(err) {
					return errors.Wrapf(err, "failed to create directory: %s", path)
				}
			}

			for _, leaf := range leaves {
				path := fmt.Sprintf("%s/%s/%s/%s", dataRoot, stage, rootNode, leaf)
				err := os.MkdirAll(path, 0760)
				if err != nil && !os.IsExist(err) {
					return errors.Wrapf(err, "failed to create directory: %s", path)
				}
			}
		}
	}
	return nil
}

func GetDirectoryPath(targetLeafNodePath [2]string) (string, bool, error) {
	var (
		hasChildren bool
		currIdx     int
	)

	dirParts := make([]string, 0, 4)
	dirParts = append(dirParts, []string{dataRoot, "output"}...)
	for maybeNodeName, maybeLeaves := range nodeMap {
		hasChildren = len(maybeLeaves) > 0

		if maybeNodeName == targetLeafNodePath[currIdx] {
			if hasChildren {
				var maybeFoundIdx = -1
				for i, c := range maybeLeaves {
					if c == targetLeafNodePath[1] {
						maybeFoundIdx = i
					}
				}
				if maybeFoundIdx < 0 {
					return "", hasChildren, fmt.Errorf("failed to find matching child for targetLeafNodePath: %v", targetLeafNodePath)
				}
			}
			dirParts = append(dirParts, targetLeafNodePath[:]...)
			return strings.Join(dirParts, "/"), hasChildren, nil
		}

		for _, maybeLeafName := range maybeLeaves {
			if maybeLeafName == targetLeafNodePath[currIdx] {
				dirParts = append(dirParts, maybeLeafName)
				return strings.Join(dirParts, "/"), hasChildren, nil
			}
		}
	}
	return "", hasChildren, fmt.Errorf("no path found matching - targetLeafNodePath: %v", targetLeafNodePath)
}
