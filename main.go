package main

import (
	"context"
	"log"
	"net/http"

	"github.com/pericles-tpt/ownapi/binary"
	"github.com/pericles-tpt/ownapi/config"
	"github.com/pericles-tpt/ownapi/db"
	"github.com/pericles-tpt/ownapi/functions"
	"github.com/pericles-tpt/ownapi/handlers"

	// TODO: Wrap the go logger instead of doing my own thing
	log2 "github.com/pericles-tpt/ownapi/log"
	"github.com/pericles-tpt/ownapi/pipelines"
	"github.com/pericles-tpt/ownapi/runtime"
	"github.com/pericles-tpt/ownapi/secrets"

	"github.com/pericles-tpt/ownapi/node"
	"github.com/pericles-tpt/ownapi/setup"
	"github.com/pericles-tpt/ownapi/views"

	"github.com/rs/cors"
)

func main() {
	var err error

	err = config.LoadConfigs("_config/startup.json", "_config/runtime.json")
	if err != nil {
		panic(err)
	}
	go config.ReloadRuntimeConfig()

	err = log2.Setup()
	if err != nil {
		panic(err)
	}

	secretsPath := "./secrets.txt"
	pw, secretNames, err := secrets.Init(secretsPath)
	if err != nil {
		panic(err)
	}

	err = functions.Init()
	if err != nil {
		panic(err)
	}

	err = binary.Init(pw)
	if err != nil {
		panic(err)
	}

	// TODO: Validate pipelines are non-enpty and don't contain empty stages, validate other stuff?
	pipelinesBytes, myPipelines, err := pipelines.Load("./_config/pipelines.json")
	if err != nil {
		panic(err)
	}

	err = secrets.PromptForMissingSecretsWipePW(secretNames, pipelinesBytes, secretsPath, pw)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	// TODO: Do something with first `queryer` argument
	_, err = db.InitDB(ctx)
	if err != nil {
		panic(err)
	}

	go runtime.AutoReload()
	go pipelines.ScheduleAutoTriggeredPipelines(myPipelines)

	err = setup.MakeDirectories()
	if err != nil {
		panic(err)
	}

	err = node.Init()
	if err != nil {
		panic(err)
	}

	frontendMux := http.NewServeMux()
	views.RegisterViews(frontendMux)
	go startServer(config.GetLocalFrontendURL(), frontendMux)

	backendMux := http.NewServeMux()
	handlers.RegisterHandlers(backendMux)
	startServer(config.GetLocalBackendURL(), backendMux)
}

func MiddlewareChain(ctx map[string]interface{}, w http.ResponseWriter, r *http.Request, middleware ...func(ctx map[string]interface{}, w http.ResponseWriter, r *http.Request)) {
	for _, mw := range middleware {
		mw(ctx, w, r)
	}
}

func TimingMiddleware(ctx map[string]interface{}, w http.ResponseWriter, r *http.Request) {

}

func startServer(url string, mux *http.ServeMux) {
	// Register Handlers

	// Setup CORS
	corsOptions := config.GetCorsOptions()
	c := cors.New(cors.Options{
		AllowedOrigins:   corsOptions.AllowedOrigins,
		AllowedMethods:   corsOptions.AllowedMethods,
		AllowCredentials: true, // TODO: Should be false when NOT running as dev
	})
	handler := c.Handler(mux)

	// Start (backend) server
	log.Printf("Listening on port %s...\n", config.GetURLPort(url))
	log.Fatal(http.ListenAndServe(url, handler))
}
