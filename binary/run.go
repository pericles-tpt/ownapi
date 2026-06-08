package binary

import (
	"os/exec"

	"github.com/pkg/errors"

	"fmt"
	"os"

	"github.com/pericles-tpt/ownapi/utility"
)

func Run(name string, args []string) error {
	idx, exists := utility.Contains(name, names)
	if !exists {
		return fmt.Errorf("failed to find binary '%s'", name)
	}
	path := paths[idx]

	bs, err := os.ReadFile(path)
	if err != nil {
		return errors.Wrapf(err, "failed to read binary at path %s", path)
	}
	gotHash := utility.HashSHA256(bs)

	valid := ValidateBinary(root, name, paths[idx], gotHash)
	if !valid {
		return fmt.Errorf("failed to validate binary, moved to quarrantine")
	}

	cmd := exec.Command(path, args...)
	var out []byte
	out, err = cmd.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "failed to run command, out: %s", out)
	}
	return nil
}
