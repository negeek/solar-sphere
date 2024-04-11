package gateway

import (
	"io/ioutil"
	"encoding/json"
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
		resp = &http.Response{}
		respBody []byte
		respData = &utils.Response{}

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

	// Get the right URL of service
	switch service {
	case "auth":
		request.URL = fmt.Sprintf("%s%s", config.Services.Auth.BaseURL, requestInfo.OriginalURL)
	case "sentinel":
		request.URL = fmt.Sprintf("%s%s", config.Services.Sentinel.BaseURL, requestInfo.OriginalURL)
	default:
		utils.JsonResponse(w, false, http.StatusBadRequest , "No such service to process request", nil)
		return 	
	}

	resp, err = utils.MakeHTTPRequest(&request)
	if err != nil {
		utils.JsonResponse(w, false, http.StatusBadGateway , err.Error(), nil)
	 	return 
	}

	if resp.Close{
		defer resp.Body.Close()
	}

	// Let's treat content-types differently
	contentType := resp.Header.Get("Content-Type")
    switch contentType {
	case "application/json":
		respBody, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			utils.JsonResponse(w, false, http.StatusBadGateway , err.Error(), nil)
	 		return 
		}

		err = json.Unmarshal(respBody, &respData) 
		if err != nil {
			utils.JsonResponse(w, false, http.StatusBadGateway , err.Error(), nil)
	 		return 
		}

		utils.JsonResponse(w, respData.Success, respData.StatusCode, respData.Message, respData.Data)
	 	return

	case "text/csv":
		respBody, err = ioutil.ReadAll(resp.Body)
        if err != nil {
            utils.JsonResponse(w, false, http.StatusBadGateway , err.Error(), nil)
	 		return
        }

		// Re-use the service headers
		for headerKey, headerValues := range resp.Header{
			for _, headerValue:= range headerValues{
				w.Header().Set(headerKey, headerValue)
			}
		}
		_, err = w.Write(respBody)
		if err != nil {
			utils.JsonResponse(w, false, http.StatusBadGateway , err.Error(), nil)
			return
		}

	default:
		utils.JsonResponse(w, false, http.StatusBadGateway ,"Unrecognised Content-Type", nil)
	 	return 
	}	
	
}