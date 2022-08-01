package oauthutil

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Token represents the credentials used to authorize
// the requests to access protected resources on the OAuth 2.0
// provider's backend.
//
// This type is a mirror of oauth2.Token and exists to break
// an otherwise-circular dependency. Other internal packages
// should convert this Token into an oauth2.Token before use.
type Token struct {
	// AccessToken is the token that authorizes and authenticates
	// the requests.
	AccessToken string

	// TokenType is the type of token.
	// The Type method returns either this or "Bearer", the default.
	TokenType string

	// RefreshToken is a token that's used by the application
	// (as opposed to the user) to refresh the access token
	// if it expires.
	RefreshToken string

	// Expiry is the optional expiration time of the access token.
	//
	// If zero, TokenSource implementations will reuse the same
	// token forever and RefreshToken or equivalent
	// mechanisms for that TokenSource will not be used.
	Expiry time.Time

	// Raw optionally contains extra metadata from the server
	// when updating a token.
	Raw interface{}
}

// tokenJSON is the struct representing the HTTP response from OAuth2
// providers returning a token or error in JSON form.
type tokenJSON struct {
	AccessToken      string         `json:"access_token"`
	TokenType        string         `json:"token_type"`
	RefreshToken     string         `json:"refresh_token"`
	ExpiresIn        expirationTime `json:"expires_in"` // at least PayPal returns string, while most return number
	Error            string         `json:"error"`
	ErrorDescription string         `json:"error_description"`
	ErrorURI         string         `json:"error_uri"`
}

func (e *tokenJSON) expiry() (t time.Time) {
	if v := e.ExpiresIn; v != 0 {
		return time.Now().Add(time.Duration(v) * time.Second)
	}
	return
}

type expirationTime int32

func (e *expirationTime) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || string(b) == "null" {
		return nil
	}
	var n json.Number
	err := json.Unmarshal(b, &n)
	if err != nil {
		return err
	}
	i, err := n.Int64()
	if err != nil {
		return err
	}
	if i > math.MaxInt32 {
		i = math.MaxInt32
	}
	*e = expirationTime(i)
	return nil
}

// RegisterBrokenAuthHeaderProvider previously did something. It is now a no-op.
//
// Deprecated: this function no longer does anything. Caller code that
// wants to avoid potential extra HTTP requests made during
// auto-probing of the provider's auth style should set
// Endpoint.AuthStyle.
func RegisterBrokenAuthHeaderProvider(tokenURL string) {}

// AuthStyle is a copy of the golang.org/x/oauth2 package's AuthStyle type.
type AuthStyle int

const (
	AuthStyleInParams AuthStyle = 1
	AuthStyleInHeader AuthStyle = 2
)

// newTokenRequest returns a new *http.Request to retrieve a new token
// from tokenURL using the provided clientID, clientSecret, and POST
// body parameters.
//
// inParams is whether the clientID & clientSecret should be encoded
// as the POST body. An 'inParams' value of true means to send it in
// the POST body (along with any values in v); false means to send it
// in the Authorization header.
func newTokenRequest(oauthCtx OauthContext, authStyle AuthStyle) (*http.Request, error) {
	v := url.Values{}
	if authStyle == AuthStyleInParams {
		v.Set("client_id", oauthCtx.ClientID)
		v.Set("client_secret", oauthCtx.ClientSecret)
	}

	fmt.Println(".... NEW TOKEN REQUEST  - V", v)

	req, err := http.NewRequest("POST", oauthCtx.Endpoint.TokenURL, strings.NewReader(v.Encode()))

	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if authStyle == AuthStyleInHeader {
		req.SetBasicAuth(url.QueryEscape(oauthCtx.ClientID), url.QueryEscape(oauthCtx.ClientSecret))
	}
	return req, nil
}

func RetrieveToken(ctx context.Context, doer httpcli.Doer, oauthCtx OauthContext, authStyle AuthStyle) (*Token, error) {

	fmt.Println("... auth style", authStyle)
	fmt.Println(".... retrieve token oauthctx", oauthCtx)
	fmt.Println(".... retrieve token ctx", ctx)

	req, err := newTokenRequest(oauthCtx, authStyle)
	if err != nil {
		return nil, err
	}

	token, err := doTokenRoundTrip(ctx, doer, req)
	if err != nil {
		return nil, errors.Wrap(err, "do token round trip")
	}

	// Don't overwrite `RefreshToken` with an empty value
	// if this was a token refreshing request.
	if token != nil && token.RefreshToken == "" {
		token.RefreshToken = oauthCtx.RefreshToken
	}

	return token, err
}

func doTokenRoundTrip(ctx context.Context, doer httpcli.Doer, req *http.Request) (*Token, error) {
	fmt.Println("... 0 = do token round trip request", req)
	r, err := doer.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "do request")
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	r.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("oauth2: cannot fetch token: %v", err)
	}

	fmt.Println("... 1 do token round trip...")
	if code := r.StatusCode; code < 200 || code > 299 {
		fmt.Println("1 A do token round trip -- r.Status code", r.StatusCode)
		fmt.Println("1 A do token round trip -- body", string(body))

		return nil, &RetrieveError{
			Response: r,
			Body:     body,
		}
	}

	fmt.Println("... 2 do token round trip...")

	var token *Token
	content, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	switch content {
	case "application/x-www-form-urlencoded", "text/plain":
		vals, err := url.ParseQuery(string(body))
		if err != nil {
			return nil, err
		}
		if tokenError := vals.Get("error"); tokenError != "" {
			return nil, &TokenError{
				Err:              tokenError,
				ErrorDescription: vals.Get("error_description"),
				ErrorURI:         vals.Get("error_uri"),
			}
		}
		token = &Token{
			AccessToken:  vals.Get("access_token"),
			TokenType:    vals.Get("token_type"),
			RefreshToken: vals.Get("refresh_token"),
			Raw:          vals,
		}
		e := vals.Get("expires_in")
		expires, _ := strconv.Atoi(e)
		if expires != 0 {
			token.Expiry = time.Now().Add(time.Duration(expires) * time.Second)
		}
	default:
		var tj tokenJSON
		if err = json.Unmarshal(body, &tj); err != nil {
			return nil, err
		}
		if tj.Error != "" {
			return nil, &TokenError{
				Err:              tj.Error,
				ErrorDescription: tj.ErrorDescription,
				ErrorURI:         tj.ErrorURI,
			}
		}
		token = &Token{
			AccessToken:  tj.AccessToken,
			TokenType:    tj.TokenType,
			RefreshToken: tj.RefreshToken,
			Expiry:       tj.expiry(),
			Raw:          make(map[string]interface{}),
		}
		json.Unmarshal(body, &token.Raw) // no error checks for optional fields
	}
	if token.AccessToken == "" {
		return nil, errors.New("oauth2: server response missing access_token")
	}
	return token, nil
}

type RetrieveError struct {
	Response *http.Response
	Body     []byte
}

func (r *RetrieveError) Error() string {
	return fmt.Sprintf("oauth2: cannot fetch token: %v\nResponse: %s", r.Response.Status, r.Body)
}

type TokenError struct {
	Err              string
	ErrorDescription string
	ErrorURI         string
}

func (t *TokenError) Error() string {
	return fmt.Sprintf("oauth2: error in token fetch repsonse: %s\nerror_description: %s\nerror_uri: %s", t.Err, t.ErrorDescription, t.ErrorURI)
}
