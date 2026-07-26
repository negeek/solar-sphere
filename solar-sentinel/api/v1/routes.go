package v1

import (
	"net/http"

	"github.com/gorilla/mux"
)

func Routes(r *mux.Router, h *Handler, authMiddleware func(http.Handler) http.Handler) {
	router := r.PathPrefix("/sentinel/v1").Subrouter()
	router.Use(authMiddleware)
	router.HandleFunc("/device/", h.CreateDevice).Methods("POST")
	router.HandleFunc("/download/{device_id}", h.DownloadSolarIrrData).Methods("GET")
}
