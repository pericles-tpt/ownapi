package config

import (
	"fmt"
	"os"
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
