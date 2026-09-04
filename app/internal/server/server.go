// Package server exposes the guestbook JSON REST API together with the static
// web page.
package server

import (
	"encoding/json"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"
)

// Server bundles the guestbook API handlers and the embedded web assets.
type Server struct {
	store   Store
	webRoot fs.FS
}

// New creates a Server serving entries from store and static web assets from
// webRoot.
func New(store Store, webRoot fs.FS) *Server {
	return &Server{
		store:   store,
		webRoot: webRoot,
	}
}

// Handler builds the HTTP handler with all routes registered.
func (t *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/entries", t.handleCreateEntry)
	mux.HandleFunc("GET /api/entries", t.handleListEntries)
	mux.Handle("GET /", http.FileServerFS(t.webRoot))

	return mux
}

type createEntryRequest struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

func (t *Server) handleCreateEntry(w http.ResponseWriter, r *http.Request) {
	var request createEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "request body must be valid JSON")
		return
	}

	name := strings.TrimSpace(request.Name)
	message := strings.TrimSpace(request.Message)
	if name == "" || message == "" {
		writeError(w, http.StatusBadRequest, "name and message must not be empty")
		return
	}

	entry, err := t.store.Create(name, message)
	if err != nil {
		slog.Error("failed creating guestbook entry", "error", err)
		writeError(w, http.StatusInternalServerError, "failed creating entry")
		return
	}

	writeJSON(w, http.StatusCreated, entry)
}

func (t *Server) handleListEntries(w http.ResponseWriter, r *http.Request) {
	entries, err := t.store.List(r.URL.Query().Get("last_id"))
	if err != nil {
		slog.Error("failed listing guestbook entries", "error", err)
		writeError(w, http.StatusInternalServerError, "failed listing entries")
		return
	}

	writeJSON(w, http.StatusOK, entries)
}
