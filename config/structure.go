package config

type Config struct {
	AppName          string      `json:"appName"`
	BackendURL       string      `json:"backendURL"`
	LocalBackendURL  string      `json:"localBackendURL"`
	FrontendURL      string      `json:"frontendURL"`
	LocalFrontendURL string      `json:"localFrontendURL"`
	CorsOptions      CorsOptions `json:"corsOptions"`
}

type CorsOptions struct {
	AllowedMethods []string `json:"allowedMethods"`
	AllowedOrigins []string `json:"allowedOrigins"`
}
