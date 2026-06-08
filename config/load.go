package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/pericles-tpt/rterror"
	"github.com/pkg/errors"
)

var (
	startupConfigPath, runtimeConfigPath string

	cfgS *ConfigStartup = nil
	cfgR *ConfigRuntime = nil

	lastConfigModtime = time.Time{}
)

func LoadConfigs(suConfigPath, rtConfigPath string) error {
	startupConfigPath = suConfigPath
	runtimeConfigPath = rtConfigPath

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

func ReloadRuntimeConfig() string {
	st, err := os.Stat(runtimeConfigPath)
	if err != nil {
		return errors.Wrapf(err, "failed to access runtime config at path '%s', please ensure it exists and has the correct permissions", runtimeConfigPath).Error()
	}

	if st.ModTime().Equal(lastConfigModtime) {
		return ""
	}
	lastConfigModtime = st.ModTime()

	file, err := os.OpenFile(runtimeConfigPath, os.O_RDONLY, 0200)
	if err != nil {
		return errors.Wrapf(err, "failed to load open runtime config at path '%s', please check file permissions", runtimeConfigPath).Error()
	}

	var tmpCfg *ConfigRuntime
	jd := json.NewDecoder(file)
	err = jd.Decode(&tmpCfg)
	if err != nil {
		return errors.Wrapf(err, "failed to decode config at path '%s', check that your config has the right structure", runtimeConfigPath).Error()
	}

	cfgR = tmpCfg
	return fmt.Sprintf("successfully reloaded config from %s!\n", runtimeConfigPath)
}
