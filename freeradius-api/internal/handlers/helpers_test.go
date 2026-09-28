package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestPathParamDecodes(t *testing.T) {
	// chi exposes the raw segment; PathParam must percent-decode it so usernames
	// like "olt-fajar.jb@dsnet" (sent as ...%40...) resolve correctly.
	h := chi.NewRouter()
	var got string
	h.Get("/users/{username}", func(w http.ResponseWriter, req *http.Request) {
		got = PathParam(req, "username")
	})
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/users/olt-fajar.jb%40dsnet", nil))
	if got != "olt-fajar.jb@dsnet" {
		t.Fatalf("PathParam = %q; want %q", got, "olt-fajar.jb@dsnet")
	}
}
