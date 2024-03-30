package v1

import (
	"github.com/gorilla/mux"
	v1middlewares "github.com/negeek/solar-sphere/solar-sentinel/middlewares/v1"
)

func Routes(r *mux.Router) {
	router := r.PathPrefix("/sentinel/v1").Subrouter()
	router.Use(v1middlewares.AuthenticationMiddleware)
	router.HandleFunc("/device/", CreateDeviceID).Methods("POST")
	router.HandleFunc("/download/{device_id}", DownloadSolarIrrData).Methods("GET")
}