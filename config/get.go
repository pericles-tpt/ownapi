package config

import (
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
	return cfg.AppName
}

func GetLocalFrontendURL() string {
	return cfg.LocalFrontendURL
}
func GetFrontendURL() string {
	return cfg.FrontendURL
}
func GetStaticFrontendURL() string {
	if GetIsDev() {
		return cfg.LocalFrontendURL
	}
	return cfg.FrontendURL

}
func GetLocalBackendURL() string {
	return cfg.LocalBackendURL
}
func GetBackendURL() string {
	return cfg.BackendURL
}
func GetURLPort(url string) string {
	urlParts := strings.Split(url, ":")
	return urlParts[len(urlParts)-1]
}

func GetCorsOptions() CorsOptions {
	return cfg.CorsOptions
}
