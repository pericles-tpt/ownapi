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
	cfgS *ConfigStatic  = nil
	cfgD *ConfigDynamic = nil

	lastConfigModtime = time.Time{}
)

func LoadConfigs(staticConfigPath, dynamicConfigPath string) error {
	file, err := os.OpenFile(staticConfigPath, os.O_RDONLY, 0200)
	if err != nil {
		return rterror.PrependErrorWithRuntimeInfo(err, "failed to open static config from '%s'", staticConfigPath)
	}

	jd := json.NewDecoder(file)
	err = jd.Decode(&cfgS)
	if err != nil {
		return rterror.PrependErrorWithRuntimeInfo(err, "failed to decode file into `ConfigStatic` structure")
	}

	file, err = os.OpenFile(dynamicConfigPath, os.O_RDONLY, 0200)
	if err != nil {
		return rterror.PrependErrorWithRuntimeInfo(err, "failed to open dynamic config from '%s'", dynamicConfigPath)
	}

	jd = json.NewDecoder(file)
	err = jd.Decode(&cfgD)
	if err != nil {
		return rterror.PrependErrorWithRuntimeInfo(err, "failed to decode file into `ConfigDynamic` structure")
	}

	return err
}

func AutoReloadDynamicConfig(dynamicConfigPath string) {
	for {
		utility.SleepLinux(time.Duration((*cfgS).DynamicConfigReloadMS) * time.Millisecond)

		st, err := os.Stat(dynamicConfigPath)
		if err != nil {
			fmt.Printf("[CONFIG] Failed to access dynamic config at path '%s', please ensure it exists and has the correct permissions\n", dynamicConfigPath)
			continue
		}

		if st.ModTime().After(lastConfigModtime) {
			lastConfigModtime = st.ModTime()
		} else {
			continue
		}

		file, err := os.OpenFile(dynamicConfigPath, os.O_RDONLY, 0200)
		if err != nil {
			fmt.Printf("[CONFIG] Failed to load open dynamic config at path '%s', please check file permissions\n", dynamicConfigPath)
			continue
		}

		var tmpCfg *ConfigDynamic
		jd := json.NewDecoder(file)
		err = jd.Decode(&tmpCfg)
		if err != nil {
			fmt.Printf("[CONFIG] Failed to decode config at path '%s', check that your config has the right structure\n", dynamicConfigPath)
			continue
		}

		cfgD = tmpCfg
		fmt.Printf("[CONFIG] Successfully reloaded config from %s!\n", dynamicConfigPath)
	}
}
