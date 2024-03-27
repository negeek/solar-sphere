package v1

import (
	"github.com/gorilla/mux"
)

func Routes(r *mux.Router) {
	router := r.PathPrefix("/sentinel/v1").Subrouter()
	router.HandleFunc("/device/", Create_DeviceID).Methods("POST")
}