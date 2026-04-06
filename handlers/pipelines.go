package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/pericles-tpt/ownapi2/pipelines"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")

		allowedOrigins := []string{"http://localhost:"}
		for _, allowed := range allowedOrigins {
			if strings.HasPrefix(origin, allowed) {
				return true
			}
		}
		return false
	},
}

func ListPipelines(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	// TODO: Auth

	names := pipelines.GetPipelineNames()
	je := json.NewEncoder(w)
	je.Encode(names)
}

func GetPipelineContent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	// TODO: Auth

	plName := r.PathValue("name")
	pl, exists, err := pipelines.GetPipleine(plName)
	if err != nil {
		if !exists {
			http.Error(w, "failed to find pipeline matching name", http.StatusBadRequest)
			return
		}
		http.Error(w, "failed to get pipeline contents", http.StatusBadRequest)
		return
	}

	je := json.NewEncoder(w)
	err = je.Encode(pl)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func RunPipeline(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.NotFound(w, r)
		return
	}

	// TODO: Auth
	// TODO: This currently runs synchronously, ideally should be async

	plName := r.PathValue("name")
	exists := pipelines.PipelineExists(plName)
	if !exists {
		http.Error(w, "failed to find pipeline matching name", http.StatusBadRequest)
		return
	}

	go pipelines.Run(plName)
}
