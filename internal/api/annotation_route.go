package api

import "net/http"

// RegisterAnnotationRoutes attaches annotation endpoints to the given mux.
func RegisterAnnotationRoutes(mux *http.ServeMux, m *AnnotationManager) {
	mux.HandleFunc("/annotations", m.handleAnnotations)
}
