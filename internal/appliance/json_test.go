package appliance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log"
)

func TestReadJSON(t *testing.T) {
	appliance := &Appliance{
		logger: log.NoOp(),
	}

	t.Run("ValidJSON", func(t *testing.T) {
		body := `{"key": "value"}`
		req := httptest.NewRequest("POST", "/", strings.NewReader(body))
		w := httptest.NewRecorder()

		var output map[string]string
		err := appliance.readJSON(w, req, &output)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if diff := cmp.Diff(map[string]string{"key": "value"}, output); diff != "" {
			t.Errorf("output mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("EmptyBody", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", nil)
		w := httptest.NewRecorder()

		var output map[string]string
		err := appliance.readJSON(w, req, &output)

		if err == nil {
			t.Error("expected an error, got nil")
		} else if err.Error() != "request body must not be empty" {
			t.Errorf("unexpected error message: got %q, want %q", err.Error(), "request body must not be empty")
		}
	})

	t.Run("MalformedJSON", func(t *testing.T) {
		body := `{"key": "value",}`
		req := httptest.NewRequest("POST", "/", strings.NewReader(body))
		w := httptest.NewRecorder()

		var output map[string]string
		err := appliance.readJSON(w, req, &output)

		if err == nil {
			t.Error("expected an error, got nil")
		} else if !strings.HasPrefix(err.Error(), "malformed JSON found at character") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("UnknownField", func(t *testing.T) {
		body := `{"unknown_field": "value"}`
		req := httptest.NewRequest("POST", "/", strings.NewReader(body))
		w := httptest.NewRecorder()

		var output struct{}
		err := appliance.readJSON(w, req, &output)

		if err == nil {
			t.Error("expected an error, got nil")
		} else if err.Error() != "request body contains unknown key" {
			t.Errorf("unexpected error message: got %q, want %q", err.Error(), "request body contains unknown key")
		}
	})

	t.Run("IncorrectJSONType", func(t *testing.T) {
		body := `{"key": 123}`
		req := httptest.NewRequest("POST", "/", strings.NewReader(body))
		w := httptest.NewRecorder()

		var output struct {
			Key string `json:"key"`
		}
		err := appliance.readJSON(w, req, &output)

		if err == nil {
			t.Error("expected an error, got nil")
		} else if err.Error() != `incorrect JSON type for field "key"` {
			t.Errorf("unexpected error message: got %q, want %q", err.Error(), `incorrect JSON type for field "key"`)
		}
	})

	t.Run("MultipleJSONValues", func(t *testing.T) {
		body := `{"key1": "value1"}{"key2": "value2"}`
		req := httptest.NewRequest("POST", "/", strings.NewReader(body))
		w := httptest.NewRecorder()

		var output map[string]string
		err := appliance.readJSON(w, req, &output)

		if err == nil {
			t.Error("expected an error, got nil")
		} else if err.Error() != "request body must only contain single JSON value" {
			t.Errorf("unexpected error message: got %q, want %q", err.Error(), "request body must only contain single JSON value")
		}
	})

	t.Run("LargeBody", func(t *testing.T) {
		// Create a large JSON object
		largeObject := map[string]string{}
		for i := 0; i < maxBytes/10; i++ {
			key := fmt.Sprintf("key%d", i)
			largeObject[key] = strings.Repeat("a", 10)
		}

		largeJSON, _ := json.Marshal(largeObject)
		// Ensure the JSON is larger than maxBytes
		largeJSON = append(largeJSON, []byte(`,"extra":"data"}`)...)

		req := httptest.NewRequest("POST", "/", bytes.NewReader(largeJSON))
		w := httptest.NewRecorder()

		var output map[string]string
		err := appliance.readJSON(w, req, &output)

		if err == nil {
			t.Error("expected an error, got nil")
		} else if !strings.HasPrefix(err.Error(), "request body larger than") {
			t.Errorf("unexpected error message: %v", err)
		}
	})
}

func TestWriteJSON(t *testing.T) {
	appliance := &Appliance{
		logger: log.NoOp(),
	}

	tests := []struct {
		name     string
		status   int
		data     responseData
		headers  http.Header
		expected string
	}{
		{
			name:   "Simple JSON response",
			status: http.StatusOK,
			data: responseData{
				"message": "Hello, World!",
			},
			headers:  nil,
			expected: "{\n\t\"message\": \"Hello, World!\"\n}\n",
		},
		{
			name:   "JSON response with custom headers",
			status: http.StatusCreated,
			data: responseData{
				"id":   1,
				"name": "Test",
			},
			headers: http.Header{
				"X-Custom-Header": []string{"CustomValue"},
			},
			expected: "{\n\t\"id\": 1,\n\t\"name\": \"Test\"\n}\n",
		},
		{
			name:     "Empty JSON response",
			status:   http.StatusNoContent,
			data:     responseData{},
			headers:  nil,
			expected: "{}\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			err := appliance.writeJSON(w, tt.status, tt.data, tt.headers)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(tt.status, w.Code); diff != "" {
				t.Errorf("status mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff("application/json", w.Header().Get("Content-Type")); diff != "" {
				t.Errorf("Content-Type mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.expected, w.Body.String()); diff != "" {
				t.Errorf("body mismatch (-want +got):\n%s", diff)
			}

			if tt.headers != nil {
				for key, value := range tt.headers {
					if diff := cmp.Diff(value, w.Header()[key]); diff != "" {
						t.Errorf("header %q mismatch (-want +got):\n%s", key, diff)
					}
				}
			}
		})
	}
}

func TestWriteJSONError(t *testing.T) {
	appliance := &Appliance{
		logger: log.NoOp(),
	}

	w := httptest.NewRecorder()
	data := responseData{
		"data": make(chan int),
	}

	err := appliance.writeJSON(w, http.StatusOK, data, nil)

	if err == nil {
		t.Error("expected an error, got nil")
	}

	expectedErrSubstring := "json: unsupported type: chan int"
	if diff := cmp.Diff(true, strings.Contains(err.Error(), expectedErrSubstring)); diff != "" {
		t.Errorf("error message mismatch (-want +got):\n%s", diff)
	}
}
