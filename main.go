package main

import (
	"context"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/pericles-tpt/ownapi/config"
	"github.com/pericles-tpt/ownapi/db"
	"github.com/pericles-tpt/ownapi/handlers"

	// TODO: Wrap the go logger instead of doing my own thing
	log2 "github.com/pericles-tpt/ownapi/log"
	"github.com/pericles-tpt/ownapi/node"
	"github.com/pericles-tpt/ownapi/pipelines"
	"github.com/pericles-tpt/ownapi/secrets"
	"github.com/pericles-tpt/ownapi/setup"
	"github.com/pericles-tpt/ownapi/views"

	"github.com/rs/cors"
)

// e.g. Garmin MTP Device -> Directory -> DB -> Dashboard (OUTPUT)
//		NSW Transport API -> DB1
//		BOM Weather API -> DB2
//		DB1 + DB2 -> Image (OUTPOT)

type InputType int

const (
	Directory InputType = iota
	API
)

/*
Functionality
	INPUTS
		- USB
			- MTP
			- MassStorage
			- Others?
		- Network
			- Input: Url, Params, Body, Method, etc
				InBetween: Traverse JSON, read properties and return a match
			- Output: Params for function as []any
	OUTPUTS
		- Update DB
		- Email
		- SMS
		- API
			- Pushover (notifications)
			- Strava POST activities
	SERVER
		- API (uses data in DB to respond)
		- Dashboard (as above)

*/

type Input struct {
	Type InputType `json:"type"`
}

// type Store struct {
// 	e
// }

type OutputType int

const (
	LocalAPI OutputType = iota
	Dashboard
	ExternalAPI
	Email
	SMS
)

// Input -> Store *-> Transform
//	-> ExternalAPI
// 	-> Email
// 	-> SMS
//	-> LocalAPI
//	-> Dashboard

type Output struct {
	Type OutputType `json:"type"`
}

func main() {
	ctx := context.Background()
	// TODO: Do something with first `queryer` argument
	_, err := db.InitDB(ctx)
	if err != nil {
		panic(err)
	}

	err = setup.MakeDirectories()
	if err != nil {
		panic(err)
	}

	err = node.Init()
	if err != nil {
		panic(err)
	}

	err = log2.Setup()
	if err != nil {
		panic(err)
	}

	// TODO: Validate pipelines are non-enpty and don't contain empty stages, validate other stuff?
	pipelinesBytes, myPipelines, err := pipelines.Load("./_config/pipelines.json")
	if err != nil {
		panic(err)
	}

	go pipelines.ScheduleAutoTriggeredPipelines(myPipelines)

	err = secrets.Init("./secrets.txt", pipelinesBytes)
	if err != nil {
		panic(err)
	}

	// Need to serve web pages
	err = godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	// Setup
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		panic(err)
	}

	frontendMux := http.NewServeMux()
	views.RegisterViews(frontendMux)
	go startServer(cfg.LocalFrontendURL, *cfg, frontendMux)

	backendMux := http.NewServeMux()
	handlers.RegisterHandlers(backendMux)
	startServer(cfg.LocalBackendURL, *cfg, backendMux)
}

func MiddlewareChain(ctx map[string]interface{}, w http.ResponseWriter, r *http.Request, middleware ...func(ctx map[string]interface{}, w http.ResponseWriter, r *http.Request)) {
	for _, mw := range middleware {
		mw(ctx, w, r)
	}
}

func TimingMiddleware(ctx map[string]interface{}, w http.ResponseWriter, r *http.Request) {

}

func startServer(url string, cfg config.Config, mux *http.ServeMux) {
	// Register Handlers

	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   cfg.CorsOptions.AllowedOrigins,
		AllowedMethods:   cfg.CorsOptions.AllowedMethods,
		AllowCredentials: true, // TODO: Should be false when NOT running as dev
	})
	handler := c.Handler(mux)

	// Start (backend) server
	log.Printf("Listening on port %s...\n", config.GetURLPort(url))
	log.Fatal(http.ListenAndServe(url, handler))
}

// Codegen -> Binaries
// IMPORT:
// METHOD:
// OUTPUT:
