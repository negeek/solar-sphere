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
		config = &utils.ServiceConfig{}
		err error
		serviceName string
		request utils.Request
		response utils.Response

	)

	requestInfo = utils.ParseHTTPRequest(r)
	config, err = utils.ReadConfig()
	if err != nil{
		fmt.Println(err.Error())
	}

	service = strings.Split(requestInfo.OriginalURL, "/")[1]
	request.Header = requestInfo.Header
	request.Method = requestInfo.Method
	request.Body = requestInfo.Body
	switch service {
	case "auth":
		request.URL = fmt.Sprintf("%s:%v%s", config.Services.Auth.BaseURL, config.Services.Auth.Port, requestInfo.OriginalURL)

	}
	response, err = utils.MakeRequest(&request)
	if err != nil {
		utils.JsonResponse(w, false, http.StatusBadGateway , err.Error(), nil)
		return 
	}

	utils.JsonResponse(w, response.Success, response.StatusCode, response.Message, response.Data)
	return
}