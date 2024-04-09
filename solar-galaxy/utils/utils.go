package utils

import(
	"io/ioutil"
	"net/http"
	"encoding/json"
	"gopkg.in/yaml.v2"
)

func JsonResponse(w http.ResponseWriter, success bool, statusCode int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(Response{
		StatusCode: statusCode
		Success:  success,
		Message: message,
		Data:    data,
	})
}

func ReadConfig() (*ServiceConfig, error) {
	data, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		return nil, err
	}

	var config ServiceConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func ParseHTTPRequest(r *http.Request) *HTTPRequestInfo {
	return &HTTPRequestInfo{
		Header:      r.Header,
		IPAddress:   r.RemoteAddr,
		UserAgent:   r.Header.Get("User-Agent"),
		Host:        r.Host,
		Method:      r.Method,
		OriginalURL: r.URL.RequestURI(),
		Body:        r.Body,
	}
}


func MakeRequest(r *Request)(Response, error){
	var (
		body *bytes.Buffer
		resp = &http.Response{}
		respBody []bytes
		err  error
		respData Response
	)
	data, err := ioutil.ReadAll(requestInfo.Body)
	if err != nil{
		return nil, error
	}

	if data != nil {
		body = bytes.NewBuffer(*data)
		req, err = http.NewRequest(r.Method, r.URL, body)
		if err != nil {
			return nil, err
		}
	} else {
		req, err = http.NewRequest(r.Method, r.URL, nil)
		if err != nil {
			return nil, err
		}
	}
	// Add headers
    for headerKey, headerValues := range r.Header{
		for headerValue:= range headerValues{
			req.Header.Add(headerKey, headerValue)
		}
    }

	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		return nil, err
	}
	
	respBody, err = ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	
	err = json.Unmarshal(respBody, &respData)
	if err != nil {
		return nil, err
	}
	return &respData, nil
}