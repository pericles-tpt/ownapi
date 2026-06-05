package config

type ConfigStatic struct {
	AppName          string      `json:"appName"`
	BackendURL       string      `json:"backendURL"`
	LocalBackendURL  string      `json:"localBackendURL"`
	FrontendURL      string      `json:"frontendURL"`
	LocalFrontendURL string      `json:"localFrontendURL"`
	CorsOptions      CorsOptions `json:"corsOptions"`

	DynamicConfigReloadMS int64 `json:"dynamicConfigReloadMS"`

	PrefixSeparator string `json:"prefixSeparator"`
	Prefixes        struct {
		Secret string `json:"secret"`
	} `json:"prefixes"`
	DataRootDir string `json:"dataRootDir"`
}

type ConfigDynamic struct {
	WebsocketSleepUS int64 `json:"webSocketSleepUS"`
	Log              struct {
		FileSizeLimit int64 `json:"fileSizeLimit"`
	} `json:"log"`
}

type CorsOptions struct {
	AllowedMethods []string `json:"allowedMethods"`
	AllowedOrigins []string `json:"allowedOrigins"`
}
