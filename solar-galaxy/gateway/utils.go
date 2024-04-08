package gateway

import(
	"net/http"
	"econding/json"
)

func JsonResponse(w http.ResponseWriter, success bool, statusCode int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(Response{
		Success:  success,
		Message: message,
		Data:    data,
	})
}

func ReadConfig() (*ServiceConfig, error) {
	data, err := ioutil.ReadFile("conf.yaml")
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