package api

import "net/http"

// RegisterLabelRoutes attaches the label endpoints to the given mux.
func RegisterLabelRoutes(mux *http.ServeMux, lm *LabelManager) {
	mux.HandleFunc("/labels", handleLabels(lm))
}
