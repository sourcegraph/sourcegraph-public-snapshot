package core

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
)

const (
	// contentType specifies the JSON Content-Type header value.
	contentType       = "application/json"
	contentTypeHeader = "Content-Type"
)

// HTTPClient is an interface for a subset of the *http.Client.
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

// EncodeURL encodes the given arguments into the URL, escaping
// values as needed.
func EncodeURL(urlFormat string, args ...interface{}) string {
	escapedArgs := make([]interface{}, 0, len(args))
	for _, arg := range args {
		escapedArgs = append(escapedArgs, url.PathEscape(fmt.Sprintf("%v", arg)))
	}
	return fmt.Sprintf(urlFormat, escapedArgs...)
}

// MergeHeaders merges the given headers together, where the right
// takes precedence over the left.
func MergeHeaders(left, right http.Header) http.Header {
	for key, values := range right {
		if len(values) > 1 {
			left[key] = values
			continue
		}
		if value := right.Get(key); value != "" {
			left.Set(key, value)
		}
	}
	return left
}

// WriteMultipartJSON writes the given value as a JSON part.
// This is used to serialize non-primitive multipart properties
// (i.e. lists, objects, etc).
func WriteMultipartJSON(writer *multipart.Writer, field string, value interface{}) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return writer.WriteField(field, string(bytes))
}

// APIError is a lightweight wrapper around the standard error
// interface that preserves the status code from the RPC, if any.
type APIError struct {
	err error

	StatusCode int `json:"-"`
}

// NewAPIError constructs a new API error.
func NewAPIError(statusCode int, err error) *APIError {
	return &APIError{
		err:        err,
		StatusCode: statusCode,
	}
}

// Unwrap returns the underlying error. This also makes the error compatible
// with errors.As and errors.Is.
func (a *APIError) Unwrap() error {
	if a == nil {
		return nil
	}
	return a.err
}

// Error returns the API error's message.
func (a *APIError) Error() string {
	if a == nil || (a.err == nil && a.StatusCode == 0) {
		return ""
	}
	if a.err == nil {
		return fmt.Sprintf("%d", a.StatusCode)
	}
	if a.StatusCode == 0 {
		return a.err.Error()
	}
	return fmt.Sprintf("%d: %s", a.StatusCode, a.err.Error())
}

// ErrorDecoder decodes *http.Response errors and returns a
// typed API error (e.g. *APIError).
type ErrorDecoder func(statusCode int, body io.Reader) error

// Caller calls APIs and deserializes their response, if any.
type Caller struct {
	client  HTTPClient
	retrier *Retrier
}

// CallerParams represents the parameters used to constrcut a new *Caller.
type CallerParams struct {
	Client      HTTPClient
	MaxAttempts uint
}

// NewCaller returns a new *Caller backed by the given parameters.
func NewCaller(params *CallerParams) *Caller {
	var httpClient HTTPClient = http.DefaultClient
	if params.Client != nil {
		httpClient = params.Client
	}
	var retryOptions []RetryOption
	if params.MaxAttempts > 0 {
		retryOptions = append(retryOptions, WithMaxAttempts(params.MaxAttempts))
	}
	return &Caller{
		client:  httpClient,
		retrier: NewRetrier(retryOptions...),
	}
}

// CallParams represents the parameters used to issue an API call.
type CallParams struct {
	URL                string
	Method             string
	MaxAttempts        uint
	Headers            http.Header
	Client             HTTPClient
	Request            interface{}
	Response           interface{}
	ResponseIsOptional bool
	ErrorDecoder       ErrorDecoder
}

// Call issues an API call according to the given call parameters.
func (c *Caller) Call(ctx context.Context, params *CallParams) error {
	req, err := newRequest(ctx, params.URL, params.Method, params.Headers, params.Request)
	if err != nil {
		return err
	}

	// If the call has been cancelled, don't issue the request.
	if err := ctx.Err(); err != nil {
		return err
	}

	client := c.client
	if params.Client != nil {
		// Use the HTTP client scoped to the request.
		client = params.Client
	}

	var retryOptions []RetryOption
	if params.MaxAttempts > 0 {
		retryOptions = append(retryOptions, WithMaxAttempts(params.MaxAttempts))
	}

	resp, err := c.retrier.Run(
		client.Do,
		req,
		params.ErrorDecoder,
		retryOptions...,
	)
	if err != nil {
		return err
	}

	// Close the response body after we're done.
	defer resp.Body.Close()

	// Check if the call was cancelled before we return the error
	// associated with the call and/or unmarshal the response data.
	if err := ctx.Err(); err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return decodeError(resp, params.ErrorDecoder)
	}

	// Mutate the response parameter in-place.
	if params.Response != nil {
		if writer, ok := params.Response.(io.Writer); ok {
			_, err = io.Copy(writer, resp.Body)
		} else {
			err = json.NewDecoder(resp.Body).Decode(params.Response)
		}
		if err != nil {
			if err == io.EOF {
				if params.ResponseIsOptional {
					// The response is optional, so we should ignore the
					// io.EOF error
					return nil
				}
				return fmt.Errorf("expected a %T response, but the server responded with nothing", params.Response)
			}
			return err
		}
	}

	return nil
}

// newRequest returns a new *http.Request with all of the fields
// required to issue the call.
func newRequest(
	ctx context.Context,
	url string,
	method string,
	endpointHeaders http.Header,
	request interface{},
) (*http.Request, error) {
	requestBody, err := newRequestBody(request)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, url, requestBody)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	req.Header.Set(contentTypeHeader, contentType)
	for name, values := range endpointHeaders {
		req.Header[name] = values
	}
	return req, nil
}

// newRequestBody returns a new io.Reader that represents the HTTP request body.
func newRequestBody(request interface{}) (io.Reader, error) {
	var requestBody io.Reader
	if request != nil {
		if body, ok := request.(io.Reader); ok {
			requestBody = body
		} else {
			requestBytes, err := json.Marshal(request)
			if err != nil {
				return nil, err
			}
			requestBody = bytes.NewReader(requestBytes)
		}
	}
	return requestBody, nil
}

// decodeError decodes the error from the given HTTP response. Note that
// it's the caller's responsibility to close the response body.
func decodeError(response *http.Response, errorDecoder ErrorDecoder) error {
	if errorDecoder != nil {
		// This endpoint has custom errors, so we'll
		// attempt to unmarshal the error into a structured
		// type based on the status code.
		return errorDecoder(response.StatusCode, response.Body)
	}
	// This endpoint doesn't have any custom error
	// types, so we just read the body as-is, and
	// put it into a normal error.
	bytes, err := io.ReadAll(response.Body)
	if err != nil && err != io.EOF {
		return err
	}
	if err == io.EOF {
		// The error didn't have a response body,
		// so all we can do is return an error
		// with the status code.
		return NewAPIError(response.StatusCode, nil)
	}
	return NewAPIError(response.StatusCode, errors.New(string(bytes)))
}
