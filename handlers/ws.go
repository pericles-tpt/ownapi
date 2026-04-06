package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pericles-tpt/ownapi2/pipelines"
	"github.com/pericles-tpt/ownapi2/utility"
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

	lastStatusBytes := []byte{}
	for {
		// Read message from browser
		// msgType, msg, err := conn.ReadMessage()
		// if err != nil {
		// 	return
		// }
		maybeNewStatus := pipelines.GetPipelinesStatuses()
		// if err != nil {
		// 	fmt.Println("Closing WS because of failure to retrieve pipeline statuses: ", err)
		// 	return
		// }

		maybeNewStatusBytes, err := json.Marshal(maybeNewStatus)
		statusChanged := err == nil && !utility.BytesEqual(lastStatusBytes, maybeNewStatusBytes)

		// // Print the message to the console
		// fmt.Printf("%s sent: %s\n", conn.RemoteAddr(), string(msg))
		// Write message back to browser
		if statusChanged {
			fmt.Println("STATUS CHANGED")
			if err = conn.WriteMessage(websocket.TextMessage, maybeNewStatusBytes); err != nil {
				return
			}
		}

		lastStatusBytes = maybeNewStatusBytes

		time.Sleep(5 * time.Microsecond)

		// Pipline: NOT RUNNING, RUNNING, ERROR
		// Stage: NOT RUNNING, RUNNING, ERROR
		// Node: NOT RUNNING, RUNNING, ERROR
	}
}
