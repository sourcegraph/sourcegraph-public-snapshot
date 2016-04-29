package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"
)

// URL is the default URL for the host of the Usage API.
// This variable can be set globally or on a per Client
// instance.
var URL = "https://usage.influxdata.com"

// Client handles all of the heavy lifting of talking
// to the Usage API for you.
type Client struct {
	URL   string // Defaults to `client.URL`
	Token string // OPTIONAL: The token of the customer making the request
}

// New returns a configured `Client`. The `token`
// is optional, but if you have it, you should pass
// it in.
func New(token string) *Client {
	return &Client{
		URL:   URL,
		Token: token,
	}
}

// Saveable needs to be implemented for types that
// want to be able to be saved to the Usage API.
type Saveable interface {
	// Path returns specific path to where this type should
	// be saved, that is everything in the path __after__ "/api/v1".
	Path() string
}

// Save does all of the heavy lifting of saving a Saveable
// Type to the Usage API. This will take care of things
// like building the full path, setting the `token` on the
// request if one is available, etc... It will also check
// the status code of the response and handle non-successful
// responses by generating a proper `error` for them.
func (c *Client) Save(s Saveable) (*http.Response, error) {
	u := fmt.Sprintf("%s/api/v1%s", c.URL, s.Path())

	b, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", u, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return nil, err
	}
	if c.Token != "" {
		req.Header.Set("X-Authorization", c.Token)
	}

	cl := http.Client{
		Timeout: time.Minute,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 0,
			Proxy:               http.ProxyFromEnvironment,
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}
	res, err := cl.Do(req)
	if err != nil {
		return res, err
	}

	code := res.StatusCode
	switch code {
	case 401, 404, 500:
		se := SimpleError{}
		err = json.NewDecoder(res.Body).Decode(&se)
		if err != nil {
			return res, err
		}
		return res, se
	case 422:
		ve := ValidationErrors{}
		err = json.NewDecoder(res.Body).Decode(&ve)
		if err != nil {
			return res, err
		}
		return res, ve
	}

	return res, err
}
