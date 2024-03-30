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