package gateway

import (
	"net/http"
	"github.com/negeek/solar-sphere/solar-galaxy/utils"
)
func Gateway(w http.ResponseWriter, r *http.Request) {
	/* 
		This is basically getting the requests and rerouting it to the right microservice. Then the response is given as response too.
	*/
}