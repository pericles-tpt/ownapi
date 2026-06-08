package runtime

import (
	"fmt"
	"time"

	"github.com/pericles-tpt/ownapi/config"
	"github.com/pericles-tpt/ownapi/functions"
	"github.com/pericles-tpt/ownapi/utility"
)

func AutoReload() {
	var (
		start         = time.Now()
		counter int64 = 1

		msg string
	)
	for {

		msg = functions.Recompile()
		if len(msg) > 0 {
			fmt.Println("[PLUGIN]: ", msg)
		}

		msg = config.ReloadRuntimeConfig()
		if len(msg) > 0 {
			fmt.Println("[CONFIG]: ", msg)
		}

		utility.SleepLinuxUntilIteration(start, counter, time.Duration(config.GetRuntimeConfigReloadMS())*time.Millisecond)
		counter++
	}
}
