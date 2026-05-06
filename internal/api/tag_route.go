package api

import "net/http"

// RegisterTagRoutes mounts tag endpoints onto the provided ServeMux.
func RegisterTagRoutes(mux *http.ServeMux, tm *TagManager) {
	mux.HandleFunc("/tags", tm.handleTags)
}
