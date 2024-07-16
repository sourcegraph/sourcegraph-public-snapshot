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

	tests := []struct {
		name                  string
		password              string
		expectedStatus        int
		shouldCallNextHandler bool
	}{
		{
			name:                  "Valid password",
			password:              "password123",
			expectedStatus:        http.StatusOK,
			shouldCallNextHandler: true,
		},
		{
			name:                  "Invalid password",
			password:              "wrongpassword",
			expectedStatus:        http.StatusUnauthorized,
			shouldCallNextHandler: false,
		},
		{
			name:                  "Empty password",
			password:              "",
			expectedStatus:        http.StatusUnauthorized,
			shouldCallNextHandler: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextHandlerCalled := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextHandlerCalled = true
				if !tt.shouldCallNextHandler {
					t.Error("Next handler should not be called after a 403")
				}
			})

			req, _ := http.NewRequest("GET", "/", nil)
			req.Header.Set(authHeaderName, tt.password)
			rr := httptest.NewRecorder()

			handler := mockAppliance.checkAuthorization(nextHandler)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			if tt.expectedStatus == http.StatusUnauthorized && nextHandlerCalled {
				t.Error("Next handler was called after a 403 response")
			}
		})
	}
}
