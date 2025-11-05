package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSONResponse(t *testing.T) {
	t.Run("success response with data", func(t *testing.T) {
		data := map[string]interface{}{
			"id":    "123",
			"name":  "Test User",
			"email": "test@example.com",
		}

		w := httptest.NewRecorder()
		JSONResponse(w, http.StatusOK, data)

		if w.Code != http.StatusOK {
			t.Errorf("JSONResponse() status = %v, want %v", w.Code, http.StatusOK)
		}

		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("JSONResponse() Content-Type = %v, want %v", contentType, "application/json")
		}

		// Проверяем тело ответа
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["id"] != "123" {
			t.Errorf("Response id = %v, want %v", response["id"], "123")
		}
	})

	t.Run("success response without data", func(t *testing.T) {
		w := httptest.NewRecorder()
		JSONResponse(w, http.StatusNoContent, nil)

		if w.Code != http.StatusNoContent {
			t.Errorf("JSONResponse() status = %v, want %v", w.Code, http.StatusNoContent)
		}

		if w.Body.Len() != 0 {
			t.Errorf("JSONResponse() body should be empty for nil data, got %v", w.Body.String())
		}
	})

	t.Run("error status code", func(t *testing.T) {
		w := httptest.NewRecorder()
		JSONResponse(w, http.StatusNotFound, map[string]string{"message": "not found"})

		if w.Code != http.StatusNotFound {
			t.Errorf("JSONResponse() status = %v, want %v", w.Code, http.StatusNotFound)
		}
	})
}

func TestErrorResponse(t *testing.T) {
	t.Run("error response", func(t *testing.T) {
		w := httptest.NewRecorder()
		ErrorResponse(w, http.StatusBadRequest, "Invalid input")

		if w.Code != http.StatusBadRequest {
			t.Errorf("ErrorResponse() status = %v, want %v", w.Code, http.StatusBadRequest)
		}

		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("ErrorResponse() Content-Type = %v, want %v", contentType, "application/json")
		}

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["error"] != "Invalid input" {
			t.Errorf("ErrorResponse() error = %v, want %v", response["error"], "Invalid input")
		}
	})

	t.Run("internal server error", func(t *testing.T) {
		w := httptest.NewRecorder()
		ErrorResponse(w, http.StatusInternalServerError, "Server error")

		if w.Code != http.StatusInternalServerError {
			t.Errorf("ErrorResponse() status = %v, want %v", w.Code, http.StatusInternalServerError)
		}

		var response map[string]string
		json.Unmarshal(w.Body.Bytes(), &response)

		if response["error"] != "Server error" {
			t.Errorf("ErrorResponse() error = %v, want %v", response["error"], "Server error")
		}
	})
}

func TestSuccessResponse(t *testing.T) {
	t.Run("success with data", func(t *testing.T) {
		data := map[string]string{
			"status":  "success",
			"message": "Operation completed",
		}

		w := httptest.NewRecorder()
		SuccessResponse(w, data)

		if w.Code != http.StatusOK {
			t.Errorf("SuccessResponse() status = %v, want %v", w.Code, http.StatusOK)
		}

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["status"] != "success" {
			t.Errorf("SuccessResponse() status = %v, want %v", response["status"], "success")
		}
	})

	t.Run("success with nil data", func(t *testing.T) {
		w := httptest.NewRecorder()
		SuccessResponse(w, nil)

		if w.Code != http.StatusOK {
			t.Errorf("SuccessResponse() status = %v, want %v", w.Code, http.StatusOK)
		}

		if w.Body.Len() != 0 {
			t.Errorf("SuccessResponse() body should be empty for nil data")
		}
	})

	t.Run("success with complex data", func(t *testing.T) {
		type User struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Email string `json:"email"`
		}

		user := User{
			ID:    "123",
			Name:  "John Doe",
			Email: "john@example.com",
		}

		w := httptest.NewRecorder()
		SuccessResponse(w, user)

		var response User
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.ID != user.ID {
			t.Errorf("Response ID = %v, want %v", response.ID, user.ID)
		}
		if response.Name != user.Name {
			t.Errorf("Response Name = %v, want %v", response.Name, user.Name)
		}
	})
}

func TestResponseEncoding(t *testing.T) {
	// Проверяем, что специальные символы правильно кодируются
	t.Run("special characters", func(t *testing.T) {
		data := map[string]string{
			"message": "Hello \"world\" & <html> entities",
			"unicode": "Привет мир! 🎉",
		}

		w := httptest.NewRecorder()
		JSONResponse(w, http.StatusOK, data)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response with special chars: %v", err)
		}

		if response["message"] != data["message"] {
			t.Errorf("Special chars not preserved: got %v, want %v", response["message"], data["message"])
		}
	})

}
