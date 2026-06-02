package handlers

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pericles-tpt/ownapi/pipelines"
	"github.com/pericles-tpt/ownapi/utility"
)

type wsPipelinesUpdate struct {
	Pipelines        []pipelines.Pipeline         `json:"pipelines"`
	PipelineStatuses []pipelines.PipelineProgress `json:"pipelineStatuses"`
}

var (
	APPROX_EIGTH_OF_AUTO_RUN_LOOP_FREQUENCY = (2500 * time.Microsecond) // 2.5ms
	WS_SLEEP_NS                             = math.Max(float64(APPROX_EIGTH_OF_AUTO_RUN_LOOP_FREQUENCY), float64(utility.MIN_SLEEP_ACCURACY))
)

func OpenClientWS(w http.ResponseWriter, r *http.Request) {
	// TODO: Communication is one-way so far, Server Sent Events (SSE) is better for this: https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events
	conn, err := upgrader.Upgrade(w, r, nil) // error ignored for sake of simplicity
	if err != nil {
		fmt.Println(r.RemoteAddr)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("Failed to open ws for client: ", err)
		return
	}

	var (
		start                          = time.Now()
		counter                  int64 = 1
		lastPipelinesUpdateBytes       = []byte{}
	)
	for {
		// Read message from browser
		// msgType, msg, err := conn.ReadMessage()
		// if err != nil {
		// 	return
		// }
		maybeNewPipelinesUpdate := wsPipelinesUpdate{
			Pipelines:        pipelines.GetPipelines(),
			PipelineStatuses: pipelines.GetPipelinesStatuses(),
		}
		// if err != nil {
		// 	fmt.Println("Closing WS because of failure to retrieve pipeline statuses: ", err)
		// 	return
		// }

		maybeNewPipelinesUpdateBytes, err := json.Marshal(maybeNewPipelinesUpdate)
		statusChanged := err == nil && !utility.BytesEqual(lastPipelinesUpdateBytes, maybeNewPipelinesUpdateBytes)

		// // Print the message to the console
		// fmt.Printf("%s sent: %s\n", conn.RemoteAddr(), string(msg))
		// Write message back to browser
		if statusChanged {
			if err = conn.WriteMessage(websocket.TextMessage, maybeNewPipelinesUpdateBytes); err != nil {
				return
			}
		}

		lastPipelinesUpdateBytes = maybeNewPipelinesUpdateBytes

		utility.SleepLinuxUntilIteration(start, counter, time.Duration(WS_SLEEP_NS))
		counter++

		// Pipline: NOT RUNNING, RUNNING, ERROR
		// Stage: NOT RUNNING, RUNNING, ERROR
		// Node: NOT RUNNING, RUNNING, ERROR
	}
}
