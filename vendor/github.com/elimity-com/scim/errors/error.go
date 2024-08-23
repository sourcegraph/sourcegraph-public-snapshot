package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

var (
	// ScimErrorInvalidFilter returns an 400 SCIM error with a detailed message.
	ScimErrorInvalidFilter = ScimError{
		ScimType: ScimTypeInvalidFilter,
		Detail:   "The specified filter syntax was invalid, or the specified attribute and filter comparison combination is not supported.",
		Status:   http.StatusBadRequest,
	}
	// ScimErrorTooMany returns an 400 SCIM error with a detailed message.
	ScimErrorTooMany = ScimError{
		ScimType: ScimTypeTooMany,
		Detail:   "The specified filter yields many more results than the server is willing to calculate or process.",
		Status:   http.StatusBadRequest,
	}
	// ScimErrorUniqueness returns an 409 SCIM error with a detailed message.
	ScimErrorUniqueness = ScimError{
		ScimType: ScimTypeUniqueness,
		Detail:   "One or more of the attribute values are already in use or are reserved.",
		Status:   http.StatusConflict,
	}
	// ScimErrorMutability returns an 400 SCIM error with a detailed message.
	ScimErrorMutability = ScimError{
		ScimType: ScimTypeMutability,
		Detail:   "The attempted modification is not compatible with the target attribute's mutability or current state.",
		Status:   http.StatusBadRequest,
	}
	// ScimErrorInvalidSyntax returns an 400 SCIM error with a detailed message.
	ScimErrorInvalidSyntax = ScimError{
		ScimType: ScimTypeInvalidSyntax,
		Detail:   "The request body message structure was invalid or did not conform to the request schema.",
		Status:   http.StatusBadRequest,
	}
	// ScimErrorInvalidPath returns an 400 SCIM error with a detailed message.
	ScimErrorInvalidPath = ScimError{
		ScimType: ScimTypeInvalidPath,
		Detail:   "The \"path\" attribute was invalid or malformed.",
		Status:   http.StatusBadRequest,
	}
	// ScimErrorNoTarget returns an 400 SCIM error with a detailed message.
	ScimErrorNoTarget = ScimError{
		ScimType: ScimTypeNoTarget,
		Detail:   "The specified path did not yield an attribute or attribute value that could be operated on.",
		Status:   http.StatusBadRequest,
	}
	// ScimErrorInvalidValue returns an 400 SCIM error with a detailed message.
	ScimErrorInvalidValue = ScimError{
		ScimType: ScimTypeInvalidValue,
		Detail:   "A required value was missing, or the value specified was not compatible with the operation or attribute type, or resource schema.",
		Status:   http.StatusBadRequest,
	}
	// ScimErrorInvalidVersion returns an 400 SCIM error with a detailed message.
	ScimErrorInvalidVersion = ScimError{
		ScimType: ScimTypeInvalidVersion,
		Detail:   "The specified SCIM protocol version is not supported.",
		Status:   http.StatusBadRequest,
	}
	// ScimErrorSensitive returns an 403 SCIM error with a detailed message.
	ScimErrorSensitive = ScimError{
		ScimType: ScimTypeSensitive,
		Detail:   "The specified request cannot be completed, due to the passing of sensitive information in a request URI.",
		Status:   http.StatusForbidden,
	}
	// ScimErrorInternal returns an 500 SCIM error without a message.
	ScimErrorInternal = ScimError{
		Status: http.StatusInternalServerError,
	}
)

var (
	applicableToAll = []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete}
	applicability   = map[int][]string{
		// The client is directed to repeat the same HTTP request at the location identified. The client SHOULD NOT use the
		// location provided in the response as a permanent reference to the resource and SHOULD continue to use the
		// original request URI [RFC7231].
		http.StatusTemporaryRedirect: applicableToAll,
		// The client is directed to repeat the same HTTP request at the location identified. The client SHOULD use the
		// location provided in the response as the permanent reference to the resource [RFC7538].
		http.StatusPermanentRedirect: applicableToAll,
		// Request is unparsable, syntactically incorrect, or violates schema.
		http.StatusBadRequest: applicableToAll,
		// Authorization failure. The authorization header is invalid or missing.
		http.StatusUnauthorized: applicableToAll,
		// Operation is not permitted based on the supplied authorization.
		http.StatusForbidden: applicableToAll,
		// Specified resource (e.g., User) or endpoint does not exist.
		http.StatusNotFound: applicableToAll,
		// The specified version number does not match the resource's latest version number, or a service provider
		// refused to create a new, duplicate resource.
		http.StatusConflict: {http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete},
		// Failed to update. Resource has changed on the server.
		http.StatusPreconditionFailed: {http.MethodPut, http.MethodPatch, http.MethodDelete},
		// {"maxOperations": 1000, "maxPayloadSize": 1048576}
		http.StatusRequestEntityTooLarge: {http.MethodPost},
		// An internal error. Implementers SHOULD provide descriptive debugging advice.
		http.StatusInternalServerError: applicableToAll,
		// Service provider does not support the request operation, e.g., PATCH.
		http.StatusNotImplemented: applicableToAll,
	}
)

func checkApplicability(err ScimError, method string) bool {
	methods, ok := applicability[err.Status]
	if !ok {
		return false
	}

	for _, m := range methods {
		if m == method {
			return true
		}
	}
	return false
}

// ScimError is a SCIM error response to indicate operation success or failure.
type ScimError struct {
	// scimType is a SCIM detail error keyword.
	ScimType ScimType
	// detail is a detailed human-readable message.
	Detail string
	// status is the HTTP status code expressed as a JSON string. REQUIRED.
	Status int
}

// CheckScimError checks whether the error's status code is defined by SCIM for the given HTTP method.
func CheckScimError(err error, method string) ScimError {
	scimErr, ok := err.(ScimError)
	if !ok {
		return ScimError{
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}
	if !checkApplicability(scimErr, method) {
		return ScimError{
			Detail: fmt.Sprintf("The HTTP status code %d is not applicable to the %s-operation.", scimErr.Status, method),
			Status: http.StatusInternalServerError,
		}
	}
	return scimErr
}

// ScimErrorBadParams returns an 400 SCIM error with a detailed message based on the invalid parameters.
func ScimErrorBadParams(invalidParams []string) ScimError {
	var suffix string

	if len(invalidParams) > 1 {
		suffix = "s"
	}

	return ScimErrorBadRequest(fmt.Sprintf(
		"Bad Request. Invalid parameter%s provided in request: %s.",
		suffix,
		strings.Join(invalidParams, ", "),
	))
}

// ScimErrorBadRequest returns an 400 SCIM error with the given message.
func ScimErrorBadRequest(msg string) ScimError {
	return ScimError{
		Detail: msg,
		Status: http.StatusBadRequest,
	}
}

// ScimErrorResourceNotFound returns an 404 SCIM error with a detailed message based on the id.
func ScimErrorResourceNotFound(id string) ScimError {
	return ScimError{
		Detail: fmt.Sprintf("Resource %s not found.", id),
		Status: http.StatusNotFound,
	}
}

func (e ScimError) Error() string {
	errorMessage := fmt.Sprint(e.Status)
	if e.ScimType != "" {
		errorMessage += fmt.Sprintf(" (%s)", e.ScimType)
	}
	if e.Detail != "" {
		return fmt.Sprintf("%s - %s", errorMessage, e.Detail)
	}
	return fmt.Sprintf("%s - No detailed human-readable message", errorMessage)
}

// MarshalJSON converts the error struct to its corresponding json representation.
func (e ScimError) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Schemas  []string `json:"schemas"`
		ScimType ScimType `json:"scimType,omitempty"`
		Detail   string   `json:"detail,omitempty"`
		Status   string   `json:"status"`
	}{
		Schemas:  []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
		ScimType: e.ScimType,
		Detail:   e.Detail,
		Status:   strconv.Itoa(e.Status),
	})
}

// UnmarshalJSON converts the error json data to its corresponding struct representation.
func (e *ScimError) UnmarshalJSON(data []byte) error {
	var tmpScimError struct {
		ScimType ScimType
		Detail   string
		Status   string
	}

	err := json.Unmarshal(data, &tmpScimError)
	if err != nil {
		return err
	}

	status, err := strconv.Atoi(tmpScimError.Status)
	if err != nil {
		return err
	}

	*e = ScimError{
		ScimType: tmpScimError.ScimType,
		Detail:   tmpScimError.Detail,
		Status:   status,
	}

	return nil
}

// ScimType is A SCIM detail error keyword.
// Source: RFC7644.3.12 Table9.
type ScimType string

const (
	// ScimTypeInvalidFilter indicates that the specified filter syntax was invalid or the specified attribute and
	// filter comparison combination is not supported.
	ScimTypeInvalidFilter ScimType = "invalidFilter"
	// ScimTypeTooMany indicates that the specified filter yields many more results than the server is willing to
	// calculate or process.
	ScimTypeTooMany ScimType = "tooMany"
	// ScimTypeUniqueness indicates that one or more of the attribute values are already in use or are reserved.
	ScimTypeUniqueness ScimType = "uniqueness"
	// ScimTypeMutability indicates that the attempted modification is not compatible with the target attribute's
	// mutability or current state.
	ScimTypeMutability ScimType = "mutability"
	// ScimTypeInvalidSyntax indicates that the request body message structure was invalid or did not conform to the
	// request schema.
	ScimTypeInvalidSyntax ScimType = "invalidSyntax"
	// ScimTypeInvalidPath indicates that the "path" attribute was invalid or malformed.
	ScimTypeInvalidPath ScimType = "invalidPath"
	// ScimTypeNoTarget indicates that the specified "path" did not yield an attribute or attribute value that could be
	// operated on.
	ScimTypeNoTarget ScimType = "noTarget"
	// ScimTypeInvalidValue indicates that a required value was missing or the value specified was not compatible with
	// the operation, attribute type or resource schema.
	ScimTypeInvalidValue ScimType = "invalidValue"
	// ScimTypeInvalidVersion indicates that the specified SCIM protocol version is not supported.
	ScimTypeInvalidVersion ScimType = "invalidVers"
	// ScimTypeSensitive indicates that the specified request cannot be completed, due to the passing of sensitive information in a request URI.
	ScimTypeSensitive ScimType = "sensitive"
)
