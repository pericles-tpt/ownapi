package binary

import (
	"github.com/pericles-tpt/ownapi/utility"
)

func Exists(binaryName string) bool {
	_, exists := utility.Contains(binaryName, names)
	return exists
}
