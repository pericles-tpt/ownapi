package functions

import (
	"fmt"
	"time"

	"github.com/pericles-tpt/ownapi/utility"
)

const AUTO_RECOMPILE_INTERVAL = time.Millisecond * 250

var funcsModified = map[string]time.Time{}

func AutoRecompile() {
	var (
		start         = time.Now()
		counter int64 = 1
	)
	for {
		_, fileBasenames, _, isNew, err := getFilesToCompile(customFunctionsPath, &funcsModified)
		if err != nil {
			fmt.Printf("ERROR: Failed to recompile, unable to get files from: %s, err: %v\n", customFunctionsPath, err)
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
			fmt.Printf("Files modified: %v, recompiling...\n", modified)
		}
		if len(created) > 0 {
			fmt.Printf("Files added: %v, recompiling...\n", created)
		}
		if recompile {
			befCompile := time.Now()
			success := reload(false)
			if success {
				took := time.Since(befCompile)
				fmt.Printf("SUCCESS: compiled in %v\n", took)
			}
		}

		utility.SleepLinuxUntilIteration(start, counter, AUTO_RECOMPILE_INTERVAL)
		counter++
	}
}
