package main

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/pericles-tpt/ownapi/config"
	"github.com/pericles-tpt/ownapi/handlers"
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
	err := setup.MakeDirectories()
	if err != nil {
		panic(err)
	}

	err = node.Init()
	if err != nil {
		panic(err)
	}

	pipelinesBytes, _, err := pipelines.Load("./_config/pipelines.json")
	if err != nil {
		panic(err)
	}

	dev, err := node.CreateUSBNode(map[string]any{}, node.USBNodeConfig{})
	if err != nil {
		panic(err)
	}
	fmt.Println("dev: ", dev)

	err = secrets.Init("./secrets.txt", pipelinesBytes)
	if err != nil {
		panic(err)
	}

	// hardcoded test of pipelines
	// hardcodedTest(pipelinesMap)

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

func hardcodedTest(pipelines map[string]pipelines.Pipeline) {
	propMap := map[string]any{}

	bef := time.Now()

	for name, pl := range pipelines {
		go func() {
			var err error
			fmt.Println("Running pipeline:", name)
			propMap, err = runPipeline(pl.Nodes, propMap)
			if err != nil {
				panic(err)
			}
			fmt.Println("Pipeline took: ", time.Since(bef))
		}()
	}
	// time.Sleep(time.Second * 5)
}

func runPipeline(pipeline [][]node.BaseNode, propMap map[string]any) (map[string]any, error) {
	var wg sync.WaitGroup

	for sn, stage := range pipeline {
		var err error
		bef := time.Now()

		wg.Add(len(stage))

		// fmt.Printf("propMap types at START of stage: %d\n", sn)
		// for k, v := range propMap {
		// 	to := reflect.TypeOf(v).String()
		// 	if strings.HasPrefix(to, "[]") || strings.HasPrefix(to, "map[") {
		// 		fmt.Printf("k: %s, tv: %s\n", k, reflect.TypeOf(v))
		// 	} else {
		// 		fmt.Printf("k: %s, tv: %v\n", k, v)
		// 	}
		// }

		stageErrCh := make(chan error, len(stage))
		outputMaps := make([]map[string]any, len(stage))
		errs := make([]error, 0, len(stage))

		for i, step := range stage {
			go func(s node.BaseNode) {
				defer wg.Done()

				// Execute the step with context
				var err error
				if outputMaps[i], err = s.Trigger(propMap); err != nil {
					// Send error to stage-specific error channel
					stageErrCh <- err
				}
			}(step)
		}

		go func() {
			wg.Wait()
			close(stageErrCh)
		}()

		for err := range stageErrCh {
			errs = append(errs, err)
		}

		for _, om := range outputMaps {
			for k, v := range om {
				propMap[k] = v
			}
		}

		propMap, err = node.UpdateKeys(propMap, sn)
		if err != nil {
			errs = append(errs, err)
		}

		if len(errs) > 0 {
			fmt.Printf("Error(s) occurred at pipeline stage %d: %v", sn, errs)
			break
		}

		fmt.Printf("propMap types at END of stage: %d\n", sn)
		for k, v := range propMap {
			to := reflect.TypeOf(v).String()
			if strings.HasPrefix(to, "[]") || strings.HasPrefix(to, "map[") {
				fmt.Printf("k: %s, tv: %s\n", k, reflect.TypeOf(v))
			} else {
				fmt.Printf("k: %s, tv: %v\n", k, v)
			}
		}

		fmt.Printf("Time taken for stage %d is: %v\n", sn, time.Since(bef))
	}
	return propMap, nil
}

// Codegen -> Binaries
// IMPORT:
// METHOD:
// OUTPUT:
