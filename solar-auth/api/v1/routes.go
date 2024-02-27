package v1

import (
	"github.com/gorilla/mux"
)

func Routes(r *mux.Router) {
	router := r.PathPrefix("/auth/v1").Subrouter()
	router.HandleFunc("/join/", Auth).Methods("POST")
	router.HandleFunc("/new_key/", NewAccessKey).Methods("POST")
}