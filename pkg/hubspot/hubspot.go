package hubspot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"

	"golang.org/x/net/context/ctxhttp"
)

// Client is a HubSpot API client
type Client struct {
	portalID string
	hapiKey  string
}

// New returns a new HubSpot client using the given Portal ID.
func New(portalID string, hapiKey string) *Client {
	return &Client{
		portalID: portalID,
		hapiKey:  hapiKey,
	}
}

// Send a POST request with JSON data to HubSpot APIs that accept JSON
// (e.g. the Contacts, Lists, etc APIs)
func (c *Client) postJSON(methodName string, baseURL *url.URL, reqPayload, respPayload interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	data, err := json.Marshal(reqPayload)
	if err != nil {
		return wrapError(methodName, err)
	}

	resp, err := ctxhttp.Post(ctx, nil, baseURL.String(), "application/json", bytes.NewReader(data))
	if err != nil {
		return wrapError(methodName, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(resp.Body)
		return wrapError(methodName, fmt.Errorf("Code %v: %s", resp.StatusCode, buf.String()))
	}

	return json.NewDecoder(resp.Body).Decode(respPayload)
}

// Send a GET request to HubSpot APIs that accept JSON in a querystring
// (e.g. the Events API)
func (c *Client) get(methodName string, baseURL *url.URL, suffix string, params map[string]string) error {
	q := make(url.Values, len(params))
	for k, v := range params {
		q.Set(k, v)
	}

	baseURL.Path = path.Join(baseURL.Path, suffix)
	baseURL.RawQuery = q.Encode()
	req, err := http.NewRequest("GET", baseURL.String(), nil)
	if err != nil {
		return wrapError(methodName, err)
	}
	req.Header.Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	resp, err := ctxhttp.Do(ctx, nil, req)
	if err != nil {
		return wrapError(methodName, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(resp.Body)
		return wrapError(methodName, fmt.Errorf("Code %v: %s", resp.StatusCode, buf.String()))
	}
	return nil
}

func wrapError(methodName string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("hubspot.%s: %v", methodName, err)
}
