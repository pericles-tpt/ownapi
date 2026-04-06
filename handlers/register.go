package handlers

import (
	"fmt"
	"net/http"
	"slices"
	"time"
)

// SOURCE: https://www.alexedwards.net/blog/organize-your-go-middleware-without-dependencies
type chain []func(http.Handler) http.Handler

func (c chain) thenFunc(h http.HandlerFunc) http.Handler {
	return c.then(h)
}
func (c chain) then(h http.Handler) http.Handler {
	for _, mw := range slices.Backward(c) {
		h = mw(h)
	}
	return h
}

var baseChain = chain{Timing}

func RegisterHandlers(m *http.ServeMux) {
	m.Handle("/pipelines/", http.StripPrefix("/pipelines", RoutePipelines()))
	m.Handle("GET /events", baseChain.thenFunc(OpenClientWS))
	// start := time.Now()
	// fmt.Printf("[be]: %s request took: %v\n", r.URL.Path, time.Since(start))
}

func RoutePipelines() *http.ServeMux {
	m := http.NewServeMux()
	m.Handle("GET /list", baseChain.thenFunc(ListPipelines))
	m.Handle("GET /content/{name}", baseChain.thenFunc(GetPipelineContent))
	m.Handle("PUT /run/{name}", baseChain.thenFunc(RunPipeline))
	return m
}

func Timing(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		h.ServeHTTP(w, r)

		fmt.Printf("[fe]: %s request took: %v\n", r.URL.Path, time.Since(start))
	})
}
