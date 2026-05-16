package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/pericles-tpt/ownapi/pipelines"
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
	pl, _, err := pipelines.GetPipelineByName(plName)
	if err != nil {
		http.Error(w, "failed to find pipeline matching name", http.StatusBadRequest)
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
	_, idx, err := pipelines.GetPipelineByName(plName)
	if err != nil {
		http.Error(w, "failed to find pipeline matching name", http.StatusBadRequest)
		return
	}

	go pipelines.Run(nil, &idx, false)
}
