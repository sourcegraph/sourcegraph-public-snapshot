// Package papertrail provides an API client for Papertrail
// (https://papertrailapp.com), a hosted log management service.
package papertrail

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/google/go-querystring/query"
)

const (
	libraryVersion = "0.0.1"
	userAgent      = "go-papertrail/" + libraryVersion
)

// A Client communicates with the Papertrail HTTP API.
type Client struct {
	// BaseURL is the base URL for all HTTP requests; by default,
	// "https://papertrailapp.com/api/v1/".
	BaseURL *url.URL

	// UserAgent is the HTTP User-Agent to send with all requests.
	UserAgent string

	httpClient *http.Client // HTTP client to use when contacting API
}

// NewClient creates a new client for communicating with the Papertrail HTTP
// API.
//
// Authentication is required to access the Papertrail API. To create a client
// whose requests are authenticated, use TokenTransport. For example:
//
//  t := &TokenTransport{Token: "my-token"} // obtain token from https://papertrailapp.com/user/edit
//  c := NewClient(t.Client())
//  // ...
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{
		BaseURL:    &url.URL{Scheme: "https", Host: "papertrailapp.com", Path: "/api/v1/"},
		UserAgent:  userAgent,
		httpClient: httpClient,
	}
}

// NewRequest creates an API request. A relative URL can be provided in urlStr,
// in which case it is resolved relative to the BaseURL of the Client. Relative
// URL paths should always be specified without a preceding slash. If opt is
// specified, its encoding (using go-querystring) is used as the request URL's
// querystring. If body is specified, the value pointed to by body is JSON
// encoded and included as the request body.
func (c *Client) NewRequest(method, urlPath string, opt interface{}, body interface{}) (*http.Request, error) {
	u := c.BaseURL.ResolveReference(&url.URL{Path: urlPath})

	if opt != nil {
		qs, err := query.Values(opt)
		if err != nil {
			return nil, err
		}
		u.RawQuery = qs.Encode()
	}

	buf := new(bytes.Buffer)
	if body != nil {
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", c.UserAgent)
	return req, nil
}

// Do sends an API request and returns the API response. The API response is
// decoded and stored in the value pointed to by v, or returned as an error if
// an API error has occurred.
func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = CheckResponse(resp)
	if err != nil {
		// even though there was an error, we still return the response
		// in case the caller wants to inspect it further
		return resp, err
	}

	if v != nil {
		if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
			return nil, fmt.Errorf("reading response from %s %s: %s", req.Method, req.URL.RequestURI(), err)
		}
	}
	return resp, nil
}

// CheckResponse checks the API response for errors, and returns them if
// present. A response is considered an error if it has a status code outside
// the 200 range. API error responses are expected to have either no response
// body, or a JSON response body that maps to ErrorResponse. Any other
// response body will be silently ignored.
func CheckResponse(r *http.Response) error {
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}
	errorResponse := &ErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && data != nil {
		json.Unmarshal(data, errorResponse)
	}
	return errorResponse
}

// An ErrorResponse reports errors caused by an API request.
type ErrorResponse struct {
	Response *http.Response `json:",omitempty"` // HTTP response that caused this error
	Message  string         // error message
}

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d %v",
		r.Response.Request.Method, r.Response.Request.URL,
		r.Response.StatusCode, r.Message)
}

func (r *ErrorResponse) HTTPStatusCode() int { return r.Response.StatusCode }
