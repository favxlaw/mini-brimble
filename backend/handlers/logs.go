package handlers

import (
	"fmt"
	"net/http"
)

func (h *Handler) StreamLogs(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	// Verify the deployment exists before opening a stream
	if _, err := h.store.GetDeployment(id); err != nil {
		http.Error(w, "deployment not found", http.StatusNotFound)
		return
	}

	// These headers turn a normal HTTP response into an SSE stream.
	// text/event-stream  → tells the browser this is SSE
	// no-cache           → prevents proxies from buffering the response
	// keep-alive         → keeps the TCP connection open
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	// Replay stored logs first so late-joining clients see the full history.
	// A user who opens the page after the build started still sees everything.
	existing, err := h.store.GetLogs(id)
	if err == nil {
		for _, log := range existing {
			fmt.Fprintf(w, "data: %s\n\n", log.Line)
		}
		flusher.Flush()
	}

	// Subscribe to live log lines from the broadcaster
	ch, cancel := h.broadcaster.Subscribe(id)
	defer cancel()

	// r.Context().Done() fires when the browser disconnects.
	// We use a select to wait for either a new log line or a disconnect.
	for {
		select {
		case line, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", line)
			flusher.Flush()

		case <-r.Context().Done():
			return
		}
	}
}
