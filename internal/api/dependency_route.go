package api

import "net/http"

// RegisterDependencyRoutes mounts dependency endpoints onto mux.
func RegisterDependencyRoutes(mux *http.ServeMux, dm *DependencyManager) {
	mux.HandleFunc("/deps", dm.handleDependencies)
}
