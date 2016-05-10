package client

import "encoding/json"

// SimpleError wraps simple error messages that come from
// the Usage API, such as:
// {"error":"json: cannot unmarshal number into Go value of type string"}
type SimpleError struct {
	Message string `json:"error"`
}

func (se SimpleError) Error() string {
	return se.Message
}

// ValidationErrors wraps more complex validation errors
// that the Usage API generates. These most usually come
// as the result of a 422 error.
type ValidationErrors struct {
	Errors map[string][]string `json:"errors"`
}

func (ve ValidationErrors) Error() string {
	b, _ := json.Marshal(ve)
	return string(b)
}
