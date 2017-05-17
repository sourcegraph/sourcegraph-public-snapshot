package clearbit

import (
	"encoding/json"
	"fmt"
)

// apiError represents a Clearbit API Error response
// https://clearbit.com/docs#errors
type apiError struct {
	Errors []ErrorDetail `json:"error"`
}

// ErrorDetail represents an individual item in an apiError.
type ErrorDetail struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// ErrorDetail represents an individual item in an apiError.
func (e apiError) Error() string {
	if len(e.Errors) > 0 {
		err := e.Errors[0]
		return fmt.Sprintf("clearbit: %s %v", err.Type, err.Message)
	}
	return ""
}

// UnmarshalJSON is used to be able to read dynamic json
//
// This is because sometimes our errors are not arrays of ErrorDetail but a
// single ErrorDetail
func (e *apiError) UnmarshalJSON(b []byte) (err error) {
	errorDetail, errors := ErrorDetail{}, []ErrorDetail{}
	if err = json.Unmarshal(b, &errors); err == nil {
		e.Errors = errors
		return
	}

	if err = json.Unmarshal(b, &errorDetail); err == nil {
		errors = append(errors, errorDetail)
		e.Errors = errors
		return
	}

	fmt.Println(err)
	return err
}

// Empty returns true if empty. Otherwise, at least 1 error message/code is
// present and false is returned.
func (e *apiError) Empty() bool {
	if len(e.Errors) == 0 {
		return true
	}
	return false
}

// relevantError returns any non-nil http-related error (creating the request,
// getting the response, decoding) if any. If the decoded apiError is non-zero
// the apiError is returned. Otherwise, no errors occurred, returns nil.
func relevantError(httpError error, ae apiError) error {
	if httpError != nil {
		return httpError
	}
	if ae.Empty() {
		return nil
	}
	return ae
}
