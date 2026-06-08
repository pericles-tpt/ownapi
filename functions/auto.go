package functions

import (
	"fmt"
	"strings"
	"time"

	"github.com/pericles-tpt/ownapi/utility"
	"github.com/pkg/errors"
)

const AUTO_RECOMPILE_INTERVAL = time.Millisecond * 250

var funcsModified = map[string]time.Time{}

func Recompile() string {
	_, fileBasenames, _, isNew, err := utility.WalkMaxDepth1(customFunctionsPath, &funcsModified, func(s string) bool { return strings.HasSuffix(s, ".go") && s != "main.go" }, func(s string) bool { return true })
	if err != nil {
		return errors.Wrapf(err, "failed to recompile, unable to get files from: %s", customFunctionsPath).Error()
	}

	var (
		created  = make([]string, 0, len(fileBasenames))
		modified = make([]string, 0, len(fileBasenames))
	)
	for i, bn := range fileBasenames {
		if (*isNew)[i] {
			created = append(created, bn)
		} else {
			modified = append(modified, bn)
		}
	}

	recompile := len(fileBasenames) > 0
	if len(modified) > 0 {
		fmt.Printf("files modified: %v, recompiling...\n", modified)
	}
	if len(created) > 0 {
		fmt.Printf("files added: %v, recompiling...\n", created)
	}
	if recompile {
		befCompile := time.Now()
		success := reload(false)
		if success {
			took := time.Since(befCompile)
			return fmt.Sprintf("finished compile in %v", took)
		}
	}
	return ""
}
