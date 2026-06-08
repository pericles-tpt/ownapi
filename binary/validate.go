package binary

import (
	"fmt"

	"github.com/pericles-tpt/ownapi/secrets"
	"github.com/pericles-tpt/ownapi/utility"
)

func ValidateBinary(root, name, path string, gotHash []byte) bool {
	expHash, err := secrets.GetKeyFromLKKS(fmt.Sprintf("binary:%s", name))
	if err != nil {
		fmt.Printf("failed to retrieve hash for expected binary from LKKS: %s, err: %v\n", name, err)
		tryMoveToQuarrantine(root, name, path, &gotHash, nil)
		return false
	}

	match := utility.BytesEqual(expHash, gotHash)
	if !match {
		tryMoveToQuarrantine(root, name, path, &gotHash, nil)
		return false
	}

	return true
}
