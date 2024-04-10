package gateway

import (
	"net/http"
	"fmt"
	"strings"
	"github.com/negeek/solar-sphere/solar-galaxy/utils"
)
func HTTPGateway(w http.ResponseWriter, r *http.Request) {
	/* 
		This is basically getting the requests and rerouting it to the right microservice. Then the response is given as response too.
	*/
	var(
		requestInfo = &utils.HTTPRequestInfo{}
		config = &utils.HTTPServiceConfig{}
		err error
		service string
		request utils.HTTPRequest
		response = &utils.Response{}

	)

	requestInfo = utils.ParseHTTPRequest(r)
	config, err = utils.ReadHTTPConfig()
	if err != nil{
		utils.JsonResponse(w, false, http.StatusBadGateway , err.Error(), nil)
		return
	}

	service = strings.Split(requestInfo.OriginalURL, "/")[1]
	request.Header = requestInfo.Header
	request.Method = requestInfo.Method
	request.Body = requestInfo.Body
	switch service {
	case "auth":
		request.URL = fmt.Sprintf("%s%s", config.Services.Auth.BaseURL, requestInfo.OriginalURL)
	case "sentinel":
		request.URL = fmt.Sprintf("%s%s", config.Services.Sentinel.BaseURL, requestInfo.OriginalURL)
	default:
		utils.JsonResponse(w, false, http.StatusBadRequest , "No such service to process request", nil)
		return 	
	}
	
	response, err = utils.MakeHTTPRequest(&request)
	if err != nil {
		utils.JsonResponse(w, false, http.StatusBadGateway , err.Error(), nil)
		return 
	}

	utils.JsonResponse(w, response.Success, response.StatusCode, response.Message, response.Data)
	return
	
}