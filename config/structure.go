package config

type ConfigStartup struct {
	AppName          string      `json:"appName"`
	BackendURL       string      `json:"backendURL"`
	LocalBackendURL  string      `json:"localBackendURL"`
	FrontendURL      string      `json:"frontendURL"`
	LocalFrontendURL string      `json:"localFrontendURL"`
	CorsOptions      CorsOptions `json:"corsOptions"`

	RuntimeConfigReloadMS int64 `json:"runtimeConfigReloadMS"`

	PrefixSeparator string `json:"prefixSeparator"`
	Prefixes        struct {
		Secret string `json:"secret"`
	} `json:"prefixes"`
	DataRootDir string `json:"dataRootDir"`
}

type ConfigRuntime struct {
	WebsocketSleepUS int64 `json:"webSocketSleepUS"`
	Log              struct {
		FileSizeLimit int64 `json:"fileSizeLimit"`
	} `json:"log"`
	InitProps map[string]map[string]any `json:"initProps"`
}

type CorsOptions struct {
	AllowedMethods []string `json:"allowedMethods"`
	AllowedOrigins []string `json:"allowedOrigins"`
}
