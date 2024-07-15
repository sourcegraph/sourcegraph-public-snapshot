package appliance

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckAuthorization(t *testing.T) {
	// Create a mock Appliance
	mockAppliance := &Appliance{
		adminPasswordBcrypt: []byte("$2y$10$o2gHR6vUX7XPQj8tjUfi/e0zel.kpgvdTdSUkQthO9hTYooDUuoay"), // bcrypt hash for "password123"
	}

	// Create a mock next handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name           string
		password       string
		expectedStatus int
	}{
		{
			name:           "Valid password",
			password:       "password123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid password",
			password:       "wrongpassword",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Empty password",
			password:       "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new request
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			// Set the authorization header
			req.Header.Set(authHeaderName, tt.password)

			// Create a ResponseRecorder to record the response
			rr := httptest.NewRecorder()

			// Call the checkAuthorization middleware
			handler := mockAppliance.checkAuthorization(nextHandler)
			handler.ServeHTTP(rr, req)

			// Check the status code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}
		})
	}
}
