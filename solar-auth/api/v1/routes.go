package v1

import (
	"github.com/gorilla/mux"
)

func Routes(r *mux.Router, h *Handler) {
	router := r.PathPrefix("/auth/v1").Subrouter()
	router.HandleFunc("/join/", h.Join).Methods("POST")
	router.HandleFunc("/new_key/", h.NewAccessKey).Methods("POST")
}
