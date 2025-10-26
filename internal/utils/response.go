package utils

import (
	"encoding/json"
	"net/http"
)

// JSONResponse отправляет JSON ответ
func JSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// ErrorResponse отправляет ошибку в JSON формате
func ErrorResponse(w http.ResponseWriter, status int, message string) {
	JSONResponse(w, status, map[string]string{
		"error": message,
	})
}

// SuccessResponse отправляет успешный JSON ответ
func SuccessResponse(w http.ResponseWriter, data interface{}) {
	JSONResponse(w, http.StatusOK, data)
}
