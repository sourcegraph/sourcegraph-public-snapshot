package ghinstallation

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

const (
	// acceptHeader is the GitHub Integrations Preview Accept header.
	acceptHeader = "application/vnd.github.machine-man-preview+json"
	apiBaseURL   = "https://api.github.com"
)

// Transport provides a http.RoundTripper by wrapping an existing
// http.RoundTripper and provides GitHub Integration authentication as an
// installation.
//
// Client can also be overwritten, and is useful to change to one which
// provides retry logic if you do experience retryable errors.
//
// See https://developer.github.com/apps/building-integrations/setting-up-and-registering-github-apps/about-authentication-options-for-github-apps/
type Transport struct {
	BaseURL        string            // baseURL is the scheme and host for GitHub API, defaults to https://api.github.com
	Client         Client            // Client to use to refresh tokens, defaults to http.Client with provided transport
	tr             http.RoundTripper // tr is the underlying roundtripper being wrapped
	key            *rsa.PrivateKey   // key is the GitHub Integration's private key
	integrationID  int               // integrationID is the GitHub Integration's Installation ID
	installationID int               // installationID is the GitHub Integration's Installation ID

	mu    *sync.Mutex  // mu protects token
	token *accessToken // token is the installation's access token
}

// accessToken is an installation access token response from GitHub
type accessToken struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

var _ http.RoundTripper = &Transport{}

// NewKeyFromFile returns an Transport using a private key from file.
func NewKeyFromFile(tr http.RoundTripper, integrationID, installationID int, privateKeyFile string) (*Transport, error) {
	privateKey, err := ioutil.ReadFile(privateKeyFile)
	if err != nil {
		return nil, fmt.Errorf("could not read private key: %s", err)
	}
	return New(tr, integrationID, installationID, privateKey)
}

// Client is a HTTP client which sends a http.Request and returns a http.Response
// or an error.
type Client interface {
	Do(*http.Request) (*http.Response, error)
}

// New returns an Transport using private key. The key is parsed
// and if any errors occur the transport is nil and error is non-nil.
//
// The provided tr http.RoundTripper should be shared between multiple
// installations to ensure reuse of underlying TCP connections.
//
// The returned Transport's RoundTrip method is safe to be used concurrently.
func New(tr http.RoundTripper, integrationID, installationID int, privateKey []byte) (*Transport, error) {
	t := &Transport{
		tr:             tr,
		integrationID:  integrationID,
		installationID: installationID,
		BaseURL:        apiBaseURL,
		Client:         &http.Client{Transport: tr},
		mu:             &sync.Mutex{},
	}
	var err error
	t.key, err = jwt.ParseRSAPrivateKeyFromPEM(privateKey)
	if err != nil {
		return nil, fmt.Errorf("could not parse private key: %s", err)
	}
	return t, nil
}

// RoundTrip implements http.RoundTripper interface.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.mu.Lock()
	if t.token == nil || t.token.ExpiresAt.Add(-time.Minute).Before(time.Now()) {
		// Token is not set or expired/nearly expired, so refresh
		if err := t.refreshToken(); err != nil {
			t.mu.Unlock()
			return nil, fmt.Errorf("could not refresh installation id %v's token: %s", t.installationID, err)
		}
	}
	t.mu.Unlock()

	req.Header.Set("Authorization", "token "+t.token.Token)
	req.Header.Set("Accept", acceptHeader)
	resp, err := t.tr.RoundTrip(req)
	return resp, err
}

func (t *Transport) refreshToken() error {
	// TODO these claims could probably be reused between installations before expiry
	claims := &jwt.StandardClaims{
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(time.Minute).Unix(),
		Issuer:    strconv.Itoa(t.integrationID),
	}
	bearer := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	ss, err := bearer.SignedString(t.key)
	if err != nil {
		return fmt.Errorf("could not sign jwt: %s", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/installations/%v/access_tokens", t.BaseURL, t.installationID), nil)
	if err != nil {
		return fmt.Errorf("could not create request: %s", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", ss))
	req.Header.Set("Accept", acceptHeader)

	resp, err := t.Client.Do(req)
	if err != nil {
		return fmt.Errorf("could not get access_tokens from GitHub API for installation ID %v: %v", t.installationID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("received non 2xx response status %q when fetching %v", resp.Status, req.URL)
	}

	if err := json.NewDecoder(resp.Body).Decode(&t.token); err != nil {
		return err
	}

	return nil
}
