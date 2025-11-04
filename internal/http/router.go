package http

import (
	"net/http"

	"github.com/gorilla/mux"
)

func NewRouter(h *Handler) *mux.Router {
	r := mux.NewRouter()

	// API routes
	r.HandleFunc("/api/status", h.GetStatus).Methods("GET")
	r.HandleFunc("/api/units", h.GetUnits).Methods("GET")
	r.HandleFunc("/api/unit/{serialNumber}", h.GetUnitDetails).Methods("GET")

	// Web interface
	r.HandleFunc("/", h.Index).Methods("GET")
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	return r
}
