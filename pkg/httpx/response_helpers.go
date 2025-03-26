package httpx

import (
	"encoding/json"
	"go-app-marketplace/pkg/reqresp"
	"net/http"
)

func WriteSuccess(w http.ResponseWriter, status int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := reqresp.StandardResponse{
		Success: true,
		Message: message,
		Data:    data,
	}

	json.NewEncoder(w).Encode(resp)
}

func WriteError(w http.ResponseWriter, status int, message string, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := reqresp.StandardResponse{
		Success: false,
		Message: message,
		Error: &reqresp.ErrorInfo{
			Code:    status,
			Details: details,
		},
	}

	json.NewEncoder(w).Encode(resp)
}
