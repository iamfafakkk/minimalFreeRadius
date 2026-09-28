package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/iamfafakkk/minimalFreeRadius/freeradius-api/internal/radiox"
)

// RadiusLogHandler serves the live FreeRADIUS log (recent snapshot + SSE).
type RadiusLogHandler struct {
	hub  *radiox.Hub
	tail *radiox.Tailer
	path string
}

func NewRadiusLogHandler(hub *radiox.Hub, tail *radiox.Tailer, path string) *RadiusLogHandler {
	return &RadiusLogHandler{hub: hub, tail: tail, path: path}
}

// Recent returns the last N buffered lines as JSON (initial page load).
func (h *RadiusLogHandler) Recent(w http.ResponseWriter, r *http.Request) {
	lines := h.tail.Recent()
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success": true, "message": "Recent RADIUS log lines retrieved successfully",
		"data": lines, "count": len(lines), "path": h.path,
	})
}

// Stream pushes new log lines as Server-Sent Events. Uses the session cookie
// for auth (EventSource cannot set headers), so the route sits behind the same
// Authenticate middleware as the rest of /api.
func (h *RadiusLogHandler) Stream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		WriteErr(w, http.StatusInternalServerError, "Streaming unsupported")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	// Backlog first so a (re)connecting client catches up without gaps.
	for _, l := range h.tail.Recent() {
		writeSSE(w, l)
	}
	flusher.Flush()

	ch := h.hub.Subscribe()
	defer h.hub.Unsubscribe(ch)

	ping := time.NewTicker(20 * time.Second)
	defer ping.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case l, open := <-ch:
			if !open {
				return
			}
			writeSSE(w, l)
			flusher.Flush()
		case <-ping.C:
			fmt.Fprint(w, ": ping\n\n")
			flusher.Flush()
		}
	}
}

func writeSSE(w http.ResponseWriter, l radiox.LogLine) {
	b, _ := json.Marshal(l)
	fmt.Fprintf(w, "event: log\ndata: %s\n\n", b)
}
