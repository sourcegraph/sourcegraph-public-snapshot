package hubspot

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/google/go-querystring/query"
	"go.uber.org/atomic"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Client is a HubSpot API client
type Client struct {
	portalID    string
	accessToken string

	lastPing       atomic.Time
	lastPingResult atomic.Error
}

// New returns a new HubSpot client using the given Portal ID.
func New(portalID, accessToken string) *Client {
	return &Client{
		portalID:    portalID,
		accessToken: accessToken,
	}
}

// Send a POST request with form data to HubSpot APIs that accept
// application/x-www-form-urlencoded data (e.g. the Forms API)
func (c *Client) postForm(methodName string, baseURL *url.URL, suffix string, body any) error {
	var data url.Values
	switch body := body.(type) {
	case map[string]string:
		data = make(url.Values, len(body))
		for k, v := range body {
			data.Set(k, v)
		}
	default:
		var err error
		data, err = query.Values(body)
		if err != nil {
			return wrapError(methodName, err)
		}
	}

	baseURL.Path = path.Join(baseURL.Path, suffix)
	req, err := http.NewRequest("POST", baseURL.String(), strings.NewReader(data.Encode()))
	if err != nil {
		return wrapError(methodName, err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	setAccessTokenAuthorizationHeader(req, c.accessToken)

	resp, err := httpcli.ExternalDoer.Do(req)
	if err != nil {
		return wrapError(methodName, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusFound {
		buf, err := io.ReadAll(resp.Body)
		if err != nil {
			return wrapError(methodName, err)
		}
		return wrapError(methodName, errors.Errorf("Code %v: %s", resp.StatusCode, string(buf)))
	}

	return nil
}

// Send a POST request with JSON data to HubSpot APIs that accept JSON
// (e.g. the Contacts, Lists, etc APIs)
func (c *Client) postJSON(methodName string, baseURL *url.URL, reqPayload, respPayload any) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	data, err := json.Marshal(reqPayload)
	if err != nil {
		return wrapError(methodName, err)
	}

	req, err := http.NewRequest("POST", baseURL.String(), bytes.NewBuffer(data))
	if err != nil {
		return wrapError(methodName, err)
	}
	req.Header.Set("Content-Type", "application/json")
	setAccessTokenAuthorizationHeader(req, c.accessToken)

	resp, err := httpcli.ExternalDoer.Do(req.WithContext(ctx))
	if err != nil {
		return wrapError(methodName, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(resp.Body)
		return wrapError(methodName, errors.Errorf("Code %v: %s", resp.StatusCode, buf.String()))
	}

	return json.NewDecoder(resp.Body).Decode(respPayload)
}

// Send a GET request to HubSpot APIs that accept JSON in a querystring
// (e.g. the Events API)
func (c *Client) get(ctx context.Context, methodName string, baseURL *url.URL, suffix string, params map[string]string) error {
	q := make(url.Values, len(params))
	for k, v := range params {
		q.Set(k, v)
	}

	baseURL.Path = path.Join(baseURL.Path, suffix)
	baseURL.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL.String(), nil)
	if err != nil {
		return wrapError(methodName, err)
	}
	req.Header.Set("Content-Type", "application/json")
	setAccessTokenAuthorizationHeader(req, c.accessToken)

	ctx, cancel := context.WithTimeout(req.Context(), time.Minute)
	defer cancel()

	resp, err := httpcli.ExternalDoer.Do(req.WithContext(ctx))
	if err != nil {
		return wrapError(methodName, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(resp.Body)
		return wrapError(methodName, errors.Errorf("Code %v: %s", resp.StatusCode, buf.String()))
	}
	return nil
}

// Ping does a naive API call to HubSpot to check if the API key is valid. The
// value of the `ttl` is used determine whether the previous ping result may be
// reused. This is to avoid wasting large volume of quotes because every ping
// consumes one rate limit quote.
func (c *Client) Ping(ctx context.Context, ttl time.Duration) error {
	if time.Since(c.lastPing.Load()) > ttl {
		c.lastPingResult.Store(
			c.get(
				ctx,
				"Ping",
				&url.URL{
					Scheme: "https",
					Host:   "api.hubapi.com",
					Path:   "/account-info/v3/details",
				},
				"",
				nil,
			),
		)
	}

	c.lastPing.Store(time.Now())
	return c.lastPingResult.Load()
}

func setAccessTokenAuthorizationHeader(req *http.Request, accessToken string) {
	if accessToken != "" {
		// As documented at:
		// https://developers.hubspot.com/docs/api/migrate-an-api-key-integration-to-a-private-app#update-the-authorization-method-of-your-integration-s-api-requests.
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
}

func wrapError(methodName string, err error) error {
	if err == nil {
		return nil
	}
	return errors.Errorf("hubspot.%s: %v", methodName, err)
}
