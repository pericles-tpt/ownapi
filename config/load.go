package config

import (
	"encoding/json"
	"os"

	"github.com/pericles-tpt/rterror"
)

var (
	cfg *Config = nil
)

func LoadConfig(fileName string) (*Config, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY, 0200)
	if err != nil {
		return nil, rterror.PrependErrorWithRuntimeInfo(err, "failed to open file named '%s'", fileName)
	}

	jd := json.NewDecoder(file)
	err = jd.Decode(&cfg)
	if err != nil {
		return nil, rterror.PrependErrorWithRuntimeInfo(err, "failed to decode file into `Config` structure")
	}

	return cfg, err
}
