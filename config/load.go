package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/pericles-tpt/ownapi/utility"
	"github.com/pericles-tpt/rterror"
)

var (
	cfgS *ConfigStartup = nil
	cfgR *ConfigRuntime = nil

	lastConfigModtime = time.Time{}
)

func LoadConfigs(startupConfigPath, runtimeConfigPath string) error {
	file, err := os.OpenFile(startupConfigPath, os.O_RDONLY, 0200)
	if err != nil {
		return rterror.PrependErrorWithRuntimeInfo(err, "failed to open startup config from '%s'", startupConfigPath)
	}

	jd := json.NewDecoder(file)
	err = jd.Decode(&cfgS)
	if err != nil {
		return rterror.PrependErrorWithRuntimeInfo(err, "failed to decode file into `ConfigStartup` structure")
	}

	file, err = os.OpenFile(runtimeConfigPath, os.O_RDONLY, 0200)
	if err != nil {
		return rterror.PrependErrorWithRuntimeInfo(err, "failed to open runtime config from '%s'", runtimeConfigPath)
	}

	jd = json.NewDecoder(file)
	err = jd.Decode(&cfgR)
	if err != nil {
		return rterror.PrependErrorWithRuntimeInfo(err, "failed to decode file into `ConfigRuntime` structure")
	}

	return err
}

func AutoReloadRuntimeConfig(runtimeConfigPath string) {
	for {
		utility.SleepLinux(time.Duration((*cfgS).RuntimeConfigReloadMS) * time.Millisecond)

		st, err := os.Stat(runtimeConfigPath)
		if err != nil {
			fmt.Printf("[CONFIG] Failed to access runtime config at path '%s', please ensure it exists and has the correct permissions\n", runtimeConfigPath)
			continue
		}

		if st.ModTime().Equal(lastConfigModtime) {
			continue
		}
		lastConfigModtime = st.ModTime()

		file, err := os.OpenFile(runtimeConfigPath, os.O_RDONLY, 0200)
		if err != nil {
			fmt.Printf("[CONFIG] Failed to load open runtime config at path '%s', please check file permissions\n", runtimeConfigPath)
			continue
		}

		var tmpCfg *ConfigRuntime
		jd := json.NewDecoder(file)
		err = jd.Decode(&tmpCfg)
		if err != nil {
			fmt.Printf("[CONFIG] Failed to decode config at path '%s', check that your config has the right structure\n", runtimeConfigPath)
			continue
		}

		cfgR = tmpCfg
		fmt.Printf("[CONFIG] Successfully reloaded config from %s!\n", runtimeConfigPath)
	}
}
