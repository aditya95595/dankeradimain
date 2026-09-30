package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"sync"
)

type sseClient struct { ch chan []byte }

type EventHub struct {
	mu sync.Mutex
	clients map[*sseClient]bool
}

var eventHub = &EventHub{clients: make(map[*sseClient]bool)}

func (h *EventHub) subscribe() *sseClient {
	client := &sseClient{ch: make(chan []byte, 64)}
	h.mu.Lock()
	h.clients[client] = true
	h.mu.Unlock()
	return client
}

func (h *EventHub) unsubscribe(client *sseClient) {
	h.mu.Lock()
	delete(h.clients, client)
	h.mu.Unlock()
	close(client.ch)
}

func (h *EventHub) broadcast(eventName string, args ...interface{}) {
	var data interface{}
	if len(args) == 1 {
		kind := reflect.TypeOf(args[0]).Kind()
		switch kind {
		case reflect.Struct, reflect.Ptr, reflect.Map:
			data = args[0]
		default:
			data = args
		}
	} else {
		data = args
	}

	if eventName == "instanceUpdate" {
		raw, err := json.Marshal(data)
		if err == nil {
			var view map[string]interface{}
			if json.Unmarshal(raw, &view) == nil {
				if account, ok := view["accountCfg"].(map[string]interface{}); ok {
					token, _ := account["token"].(string)
					account["token"] = ""
					if token != "" {
						view["instanceId"] = instanceID(token)
					}
				}
				data = view
			}
		}
	}

	payload, err := json.Marshal(map[string]interface{}{"name": eventName, "data": data})
	if err != nil {
		slog.Error("Failed to marshal SSE event", "error", err)
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	for client := range h.clients {
		select {
		case client.ch <- payload:
		default:
		}
	}
}

func sseHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	client := eventHub.subscribe()
	defer eventHub.unsubscribe(client)

	_, _ = fmt.Fprintf(w, ": connected\n\n")
	flusher.Flush()

	for {
		select {
		case msg, ok := <-client.ch:
			if !ok {
				return
			}
			_, _ = fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
