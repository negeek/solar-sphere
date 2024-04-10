package utils

import(
	"io"
	"io/ioutil"
	"net/http"
	"encoding/json"
	"gopkg.in/yaml.v2"
)

func JsonResponse(w http.ResponseWriter, success bool, statusCode int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(Response{
		StatusCode: statusCode,
		Success:  success,
		Message: message,
		Data:    data,
	})
}

func ReadHTTPConfig() (*HTTPServiceConfig, error) {
	data, err := ioutil.ReadFile("http_config.yaml")
	if err != nil {
		return nil, err
	}

	var config HTTPServiceConfig
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


func MakeHTTPRequest(r *HTTPRequest)(*Response, error){
	var (
		body io.Reader = r.Body
		resp = &http.Response{}
		respBody []byte
		err  error
		respData Response
		req *http.Request
	)
	
	if body != nil {
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
		for _, headerValue:= range headerValues{
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