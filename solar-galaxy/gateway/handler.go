package gateway

import (
	"net/http"
)
func ReRoute(w http.ResponseWriter, r *http.Request) {
	/* 
		This is basically getting the requests and rerouting it to the right microservice. Then the response is given as response too.
	*/
}