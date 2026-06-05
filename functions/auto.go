package functions

import (
	"fmt"
	"os"
	"strings"
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
		dirents, err := os.ReadDir(customFunctionsPath)
		if err != nil {
			fmt.Printf("ERROR: Failed to recompile, unable to read path: %s\n", customFunctionsPath)
		}

		// TODO: Handle delete case, a bit more complicated since deleted files could contain
		// 		 functions in-use by a pipeline
		var (
			created  = make([]string, 0, len(dirents))
			modified = make([]string, 0, len(dirents))
		)
		for _, de := range dirents {
			var (
				name             = de.Name()
				prevLastModified time.Time
				exists           bool
			)
			if de.Type().IsRegular() && strings.HasSuffix(name, ".go") && name != "main.go" {
				info, err := de.Info()
				if err != nil {
					fmt.Printf("WARN: Failed to read file: %s\n", name)
					continue
				}
				currLastModified := info.ModTime()

				if prevLastModified, exists = funcsModified[name]; !exists {
					created = append(created, name)
				} else if prevLastModified != currLastModified {
					modified = append(modified, name)
				}
				funcsModified[name] = currLastModified
			}
		}

		var recompile bool
		if len(modified) > 0 {
			fmt.Printf("Files modified: %v, recompiling...\n", modified)
			recompile = true
		}
		if len(created) > 0 {
			fmt.Printf("Files added: %v, recompiling...\n", created)
			recompile = true
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
