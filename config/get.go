package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

var isDev *bool

// TODO: This technically isn't in config, but this is the most suitable place
func GetIsDev() bool {
	if isDev == nil {
		b := (os.Getenv("IS_DEV") == "true")
		isDev = &b
	}
	return *isDev
}

func GetAppName() string {
	return cfgS.AppName
}

func GetLocalFrontendURL() string {
	return cfgS.LocalFrontendURL
}
func GetFrontendURL() string {
	return cfgS.FrontendURL
}
func GetStaticFrontendURL() string {
	if GetIsDev() {
		return cfgS.LocalFrontendURL
	}
	return cfgS.FrontendURL

}
func GetLocalBackendURL() string {
	return cfgS.LocalBackendURL
}
func GetBackendURL() string {
	return cfgS.BackendURL
}
func GetURLPort(url string) string {
	urlParts := strings.Split(url, ":")
	return urlParts[len(urlParts)-1]
}

func GetCorsOptions() CorsOptions {
	return cfgS.CorsOptions
}

func GetRuntimeConfigReloadMS() int64 {
	return cfgS.RuntimeConfigReloadMS
}

func GetPrefixesSeparator() string {
	return cfgS.PrefixSeparator
}
func GetSecretsPrefix() string {
	return cfgS.Prefixes.Secret
}
func GetDataDir(path string) string {
	return fmt.Sprintf("%s/%s", strings.TrimSuffix(cfgS.DataRootDir, "/"), strings.TrimPrefix(path, "/"))
}

// RUNTIME
func GetWebsocketSleepUS() int64 {
	return cfgR.WebsocketSleepUS
}
func GetLogFilesizeLimit() int64 {
	return cfgR.Log.FileSizeLimit
}
func GetInitPropsForPipeline(pipelineName string) map[string]any {
	var (
		ret = map[string]any{}

		globals  = cfgR.InitProps[""]
		pipeline = cfgR.InitProps[pipelineName]
	)

	for k, v := range globals {
		pre := getPrefix(k)
		ret[fmt.Sprintf("%s:%s", pre, k)] = v
	}
	for k, v := range pipeline {
		pre := getPrefix(k)
		ret[fmt.Sprintf("%s:%s", pre, k)] = v
	}
	return ret
}

func getPrefix(key string) string {
	parts := strings.Split(key, ":")
	if len(parts) == 1 {
		return "input"
	}
	_, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return "input"
	}
	return "output"
}
