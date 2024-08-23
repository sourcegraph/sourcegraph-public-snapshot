// Copyright 2010 Gary Burd
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

// Package oauth is consumer interface for OAuth 1.0, OAuth 1.0a and RFC 5849.
//
// Redirection-based Authorization
//
// This section outlines how to use the oauth package in redirection-based
// authorization (http://tools.ietf.org/html/rfc5849#section-2).
//
// Step 1: Create a Client using credentials and URIs provided by the server.
// The Client can be initialized once at application startup and stored in a
// package-level variable.
//
// Step 2: Request temporary credentials using the Client
// RequestTemporaryCredentials method. The callbackURL parameter is the URL of
// the callback handler in step 4. Save the returned credential secret so that
// it can be later found using credential token as a key. The secret can be
// stored in a database keyed by the token. Another option is to store the
// token and secret in session storage or a cookie.
//
// Step 3: Redirect the user to URL returned from AuthorizationURL method. The
// AuthorizationURL method uses the temporary credentials from step 2 and other
// parameters as specified by the server.
//
// Step 4: The server redirects back to the callback URL specified in step 2
// with the temporary token and a verifier. Use the temporary token to find the
// temporary secret saved in step 2. Using the temporary token, temporary
// secret and verifier, request token credentials using the client RequestToken
// method. Save the returned credentials for later use in the application.
//
// Signing Requests
//
// The Client type has two low-level methods for signing requests, SignForm and
// SetAuthorizationHeader.
//
// The SignForm method adds an OAuth signature to a form. The application makes
// an authenticated request by encoding the modified form to the query string
// or request body.
//
// The SetAuthorizationHeader method adds an OAuth signature to a request
// header. The SetAuthorizationHeader method is the only way to correctly sign
// a request if the application sets the URL Opaque field when making a
// request.
//
// The Get, Put, Post and Delete methods sign and invoke a request using the
// supplied net/http Client. These methods are easy to use, but not as flexible
// as constructing a request using one of the low-level methods.
//
// Context With HTTP Client
//
// A context-enabled method can include a custom HTTP client in the
// context and execute an HTTP request using the included HTTP client.
//
//     hc := &http.Client{Timeout: 2 * time.Second}
//     ctx := context.WithValue(context.Background(), oauth.HTTPClient, hc)
//     c := oauth.Client{ /* Any settings */ }
//     resp, err := c.GetContext(ctx, &oauth.Credentials{}, rawurl, nil)
package oauth // import "github.com/gomodule/oauth1/oauth"

import (
	"bytes"
	"context"
	"crypto"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

// noscape[b] is true if b should not be escaped per section 3.6 of the RFC.
var noEscape = [256]bool{
	'A': true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true,
	'a': true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true,
	'0': true, true, true, true, true, true, true, true, true, true,
	'-': true,
	'.': true,
	'_': true,
	'~': true,
}

// encode encodes string per section 3.6 of the RFC. If double is true, then
// the encoding is applied twice.
func encode(s string, double bool) []byte {
	// Compute size of result.
	m := 3
	if double {
		m = 5
	}
	n := 0
	for i := 0; i < len(s); i++ {
		if noEscape[s[i]] {
			n++
		} else {
			n += m
		}
	}

	p := make([]byte, n)

	// Encode it.
	j := 0
	for i := 0; i < len(s); i++ {
		b := s[i]
		if noEscape[b] {
			p[j] = b
			j++
		} else if double {
			p[j] = '%'
			p[j+1] = '2'
			p[j+2] = '5'
			p[j+3] = "0123456789ABCDEF"[b>>4]
			p[j+4] = "0123456789ABCDEF"[b&15]
			j += 5
		} else {
			p[j] = '%'
			p[j+1] = "0123456789ABCDEF"[b>>4]
			p[j+2] = "0123456789ABCDEF"[b&15]
			j += 3
		}
	}
	return p
}

type keyValue struct{ key, value []byte }

type byKeyValue []keyValue

func (p byKeyValue) Len() int      { return len(p) }
func (p byKeyValue) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p byKeyValue) Less(i, j int) bool {
	sgn := bytes.Compare(p[i].key, p[j].key)
	if sgn == 0 {
		sgn = bytes.Compare(p[i].value, p[j].value)
	}
	return sgn < 0
}

func (p byKeyValue) appendValues(values url.Values) byKeyValue {
	for k, vs := range values {
		k := encode(k, true)
		for _, v := range vs {
			v := encode(v, true)
			p = append(p, keyValue{k, v})
		}
	}
	return p
}

// writeBaseString writes method, url, and params to w using the OAuth signature
// base string computation described in section 3.4.1 of the RFC.
func writeBaseString(w io.Writer, method string, u *url.URL, form url.Values, oauthParams map[string]string) {
	// Method
	w.Write(encode(strings.ToUpper(method), false))
	w.Write([]byte{'&'})

	// URL
	scheme := strings.ToLower(u.Scheme)
	host := strings.ToLower(u.Host)

	uNoQuery := *u
	uNoQuery.RawQuery = ""
	path := uNoQuery.RequestURI()

	switch {
	case scheme == "http" && strings.HasSuffix(host, ":80"):
		host = host[:len(host)-len(":80")]
	case scheme == "https" && strings.HasSuffix(host, ":443"):
		host = host[:len(host)-len(":443")]
	}

	w.Write(encode(scheme, false))
	w.Write(encode("://", false))
	w.Write(encode(host, false))
	w.Write(encode(path, false))
	w.Write([]byte{'&'})

	// Create sorted slice of encoded parameters. Parameter keys and values are
	// double encoded in a single step. This is safe because double encoding
	// does not change the sort order.
	queryParams := u.Query()
	p := make(byKeyValue, 0, len(form)+len(queryParams)+len(oauthParams))
	p = p.appendValues(form)
	p = p.appendValues(queryParams)
	for k, v := range oauthParams {
		p = append(p, keyValue{encode(k, true), encode(v, true)})
	}
	sort.Sort(p)

	// Write the parameters.
	encodedAmp := encode("&", false)
	encodedEqual := encode("=", false)
	sep := false
	for _, kv := range p {
		if sep {
			w.Write(encodedAmp)
		} else {
			sep = true
		}
		w.Write(kv.key)
		w.Write(encodedEqual)
		w.Write(kv.value)
	}
}

var nonceCounter uint64

func init() {
	if err := binary.Read(rand.Reader, binary.BigEndian, &nonceCounter); err != nil {
		// fallback to time if rand reader is broken
		nonceCounter = uint64(time.Now().UnixNano())
	}
}

// nonce returns a unique string.
func nonce() string {
	return strconv.FormatUint(atomic.AddUint64(&nonceCounter, 1), 16)
}

// SignatureMethod identifies a signature method.
type SignatureMethod int

func (sm SignatureMethod) String() string {
	switch sm {
	case RSASHA1:
		return "RSA-SHA1"
	case RSASHA256:
		return "RSA-SHA256"
	case HMACSHA1:
		return "HMAC-SHA1"
	case HMACSHA256:
		return "HMAC-SHA256"
	case PLAINTEXT:
		return "PLAINTEXT"
	default:
		return "unknown"
	}
}

const (
	HMACSHA1   SignatureMethod = iota // HMAC-SHA1
	RSASHA1                           // RSA-SHA1
	PLAINTEXT                         // Plain text
	HMACSHA256                        // HMAC-256
	RSASHA256                         // RSA-SHA256
)

// Credentials represents client, temporary and token credentials.
type Credentials struct {
	Token  string // Also known as consumer key or access token.
	Secret string // Also known as consumer secret or access token secret.
}

// Client represents an OAuth client.
type Client struct {
	// Credentials specifies the client key and secret.
	// Also known as the consumer key and secret
	Credentials Credentials

	// TemporaryCredentialRequestURI is the endpoint used by the client to
	// obtain a set of temporary credentials. Also known as the request token
	// URL.
	TemporaryCredentialRequestURI string

	// ResourceOwnerAuthorizationURI is the endpoint to which the resource
	// owner is redirected to grant authorization. Also known as authorization
	// URL.
	ResourceOwnerAuthorizationURI string

	// TokenRequestURI is the endpoint used by the client to request a set of
	// token credentials using a set of temporary credentials. Also known as
	// access token URL.
	TokenRequestURI string

	// RenewCredentialRequestURI is the endpoint the client uses to
	// request a new set of token credentials using the old set of credentials.
	RenewCredentialRequestURI string

	// TemporaryCredentialsMethod is the HTTP method used by the client to
	// obtain a set of temporary credentials. If this field is the empty
	// string, then POST is used.
	TemporaryCredentialsMethod string

	// TokenCredentailsMethod is the HTTP method used by the client to request
	// a set of token credentials. If this field is the empty string, then POST
	// is used.
	TokenCredentailsMethod string

	// Header specifies optional extra headers for requests.
	Header http.Header

	// SignatureMethod specifies the method for signing a request.
	SignatureMethod SignatureMethod

	// PrivateKey is the private key to use for RSA-SHA* signatures. This field
	// must be set for RSA-SHA* signatures and ignored for other signature
	// methods.
	PrivateKey *rsa.PrivateKey
}

type request struct {
	credentials   *Credentials
	method        string
	u             *url.URL
	form          url.Values
	verifier      string
	sessionHandle string
	callbackURL   string
}

var testHook = func(map[string]string) {}

// oauthParams returns the OAuth request parameters for the given credentials,
// method, URL and application params. See
// http://tools.ietf.org/html/rfc5849#section-3.4 for more information about
// signatures.
func (c *Client) oauthParams(r *request) (map[string]string, error) {
	oauthParams := map[string]string{
		"oauth_consumer_key":     c.Credentials.Token,
		"oauth_signature_method": c.SignatureMethod.String(),
		"oauth_version":          "1.0",
	}

	if c.SignatureMethod != PLAINTEXT {
		oauthParams["oauth_timestamp"] = strconv.FormatInt(time.Now().Unix(), 10)
		oauthParams["oauth_nonce"] = nonce()
	}

	if r.credentials != nil {
		oauthParams["oauth_token"] = r.credentials.Token
	}

	if r.verifier != "" {
		oauthParams["oauth_verifier"] = r.verifier
	}

	if r.sessionHandle != "" {
		oauthParams["oauth_session_handle"] = r.sessionHandle
	}

	if r.callbackURL != "" {
		oauthParams["oauth_callback"] = r.callbackURL
	}

	testHook(oauthParams)

	var (
		signature string
		err       error
	)
	switch c.SignatureMethod {
	case HMACSHA1:
		signature = c.hmacSignature(r, sha1.New, oauthParams)
	case HMACSHA256:
		signature = c.hmacSignature(r, sha256.New, oauthParams)
	case RSASHA1:
		signature, err = c.rsaSignature(r, crypto.SHA1, oauthParams)
	case RSASHA256:
		signature, err = c.rsaSignature(r, crypto.SHA256, oauthParams)
	case PLAINTEXT:
		signature = c.plainTextSignature(r)
	default:
		err = errors.New("oauth: unknown signature method")
	}
	if err != nil {
		return nil, err
	}

	oauthParams["oauth_signature"] = signature
	return oauthParams, nil
}

func (c *Client) plainTextSignature(r *request) string {
	signature := encode(c.Credentials.Secret, false)
	signature = append(signature, '&')
	if r.credentials != nil {
		signature = append(signature, encode(r.credentials.Secret, false)...)
	}
	return string(signature)
}

func (c *Client) hmacSignature(r *request, h func() hash.Hash, oauthParams map[string]string) string {
	key := encode(c.Credentials.Secret, false)
	key = append(key, '&')
	if r.credentials != nil {
		key = append(key, encode(r.credentials.Secret, false)...)
	}
	hm := hmac.New(h, key)
	writeBaseString(hm, r.method, r.u, r.form, oauthParams)
	return base64.StdEncoding.EncodeToString(hm.Sum(key[:0]))
}

func (c *Client) rsaSignature(r *request, h crypto.Hash, oauthParams map[string]string) (string, error) {
	if c.PrivateKey == nil {
		return "", errors.New("oauth: private key not set")
	}
	w := h.New()
	writeBaseString(w, r.method, r.u, r.form, oauthParams)
	rawSignature, err := rsa.SignPKCS1v15(rand.Reader, c.PrivateKey, h, w.Sum(nil))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(rawSignature), nil
}

// SignForm adds an OAuth signature to form. The urlStr argument must not
// include a query string.
//
// See http://tools.ietf.org/html/rfc5849#section-3.5.2 for
// information about transmitting OAuth parameters in a request body and
// http://tools.ietf.org/html/rfc5849#section-3.5.2 for information about
// transmitting OAuth parameters in a query string.
func (c *Client) SignForm(credentials *Credentials, method, urlStr string, form url.Values) error {
	u, err := url.Parse(urlStr)
	switch {
	case err != nil:
		return err
	case u.RawQuery != "":
		return errors.New("oauth: urlStr argument to SignForm must not include a query string")
	}
	p, err := c.oauthParams(&request{credentials: credentials, method: method, u: u, form: form})
	if err != nil {
		return err
	}
	for k, v := range p {
		form.Set(k, v)
	}
	return nil
}

// SignParam is deprecated. Use SignForm instead.
func (c *Client) SignParam(credentials *Credentials, method, urlStr string, params url.Values) {
	u, _ := url.Parse(urlStr)
	u.RawQuery = ""
	p, _ := c.oauthParams(&request{credentials: credentials, method: method, u: u, form: params})
	for k, v := range p {
		params.Set(k, v)
	}
}

var oauthKeys = []string{
	"oauth_consumer_key",
	"oauth_nonce",
	"oauth_signature",
	"oauth_signature_method",
	"oauth_timestamp",
	"oauth_token",
	"oauth_version",
	"oauth_callback",
	"oauth_verifier",
	"oauth_session_handle",
}

func (c *Client) authorizationHeader(r *request) (string, error) {
	p, err := c.oauthParams(r)
	if err != nil {
		return "", err
	}
	var h []byte
	// Append parameters in a fixed order to support testing.
	for _, k := range oauthKeys {
		if v, ok := p[k]; ok {
			if h == nil {
				h = []byte(`OAuth `)
			} else {
				h = append(h, ", "...)
			}
			h = append(h, k...)
			h = append(h, `="`...)
			h = append(h, encode(v, false)...)
			h = append(h, '"')
		}
	}
	return string(h), nil
}

// AuthorizationHeader returns the HTTP authorization header value for given
// method, URL and parameters.
//
// AuthorizationHeader is deprecated. Use SetAuthorizationHeader instead.
func (c *Client) AuthorizationHeader(credentials *Credentials, method string, u *url.URL, params url.Values) string {
	// Signing a request can return an error. This method is deprecated because
	// this method does not return an error.
	v, _ := c.authorizationHeader(&request{credentials: credentials, method: method, u: u, form: params})
	return v
}

// SetAuthorizationHeader adds an OAuth signature to a request header.
//
// See http://tools.ietf.org/html/rfc5849#section-3.5.1 for information about
// transmitting OAuth parameters in an HTTP request header.
func (c *Client) SetAuthorizationHeader(header http.Header, credentials *Credentials, method string, u *url.URL, form url.Values) error {
	v, err := c.authorizationHeader(&request{credentials: credentials, method: method, u: u, form: form})
	if err != nil {
		return err
	}
	header.Set("Authorization", v)
	return nil
}

func (c *Client) do(ctx context.Context, urlStr string, r *request) (*http.Response, error) {
	var body io.Reader
	if r.method != http.MethodGet {
		body = strings.NewReader(r.form.Encode())
	}
	req, err := http.NewRequest(r.method, urlStr, body)
	if err != nil {
		return nil, err
	}
	if req.URL.RawQuery != "" {
		return nil, errors.New("oauth: url must not contain a query string")
	}
	for k, v := range c.Header {
		req.Header[k] = v
	}
	r.u = req.URL
	auth, err := c.authorizationHeader(r)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", auth)
	if r.method == http.MethodGet {
		req.URL.RawQuery = r.form.Encode()
	} else {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	req = req.WithContext(ctx)
	client := contextClient(ctx)
	return client.Do(req)
}

// Get issues a GET to the specified URL with form added as a query string.
func (c *Client) Get(client *http.Client, credentials *Credentials, urlStr string, form url.Values) (*http.Response, error) {
	ctx := context.WithValue(context.Background(), HTTPClient, client)
	return c.GetContext(ctx, credentials, urlStr, form)
}

// GetContext uses Context to perform Get.
func (c *Client) GetContext(ctx context.Context, credentials *Credentials, urlStr string, form url.Values) (*http.Response, error) {
	return c.do(ctx, urlStr, &request{method: http.MethodGet, credentials: credentials, form: form})
}

// Post issues a POST with the specified form.
func (c *Client) Post(client *http.Client, credentials *Credentials, urlStr string, form url.Values) (*http.Response, error) {
	ctx := context.WithValue(context.Background(), HTTPClient, client)
	return c.PostContext(ctx, credentials, urlStr, form)
}

// PostContext uses Context to perform Post.
func (c *Client) PostContext(ctx context.Context, credentials *Credentials, urlStr string, form url.Values) (*http.Response, error) {
	return c.do(ctx, urlStr, &request{method: http.MethodPost, credentials: credentials, form: form})
}

// Delete issues a DELETE with the specified form.
func (c *Client) Delete(client *http.Client, credentials *Credentials, urlStr string, form url.Values) (*http.Response, error) {
	ctx := context.WithValue(context.Background(), HTTPClient, client)
	return c.DeleteContext(ctx, credentials, urlStr, form)
}

// DeleteContext uses Context to perform Delete.
func (c *Client) DeleteContext(ctx context.Context, credentials *Credentials, urlStr string, form url.Values) (*http.Response, error) {
	return c.do(ctx, urlStr, &request{method: http.MethodDelete, credentials: credentials, form: form})
}

// Put issues a PUT with the specified form.
func (c *Client) Put(client *http.Client, credentials *Credentials, urlStr string, form url.Values) (*http.Response, error) {
	ctx := context.WithValue(context.Background(), HTTPClient, client)
	return c.PutContext(ctx, credentials, urlStr, form)
}

// PutContext uses Context to perform Put.
func (c *Client) PutContext(ctx context.Context, credentials *Credentials, urlStr string, form url.Values) (*http.Response, error) {
	return c.do(ctx, urlStr, &request{method: http.MethodPut, credentials: credentials, form: form})
}

func (c *Client) requestCredentials(ctx context.Context, u string, r *request) (*Credentials, url.Values, error) {
	if r.method == "" {
		r.method = http.MethodPost
	}
	resp, err := c.do(ctx, u, r)
	if err != nil {
		return nil, nil, err
	}
	p, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, nil, RequestCredentialsError{StatusCode: resp.StatusCode, Header: resp.Header,
			Body: p, msg: err.Error()}
	}
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return nil, nil, RequestCredentialsError{StatusCode: resp.StatusCode, Header: resp.Header,
			Body: p, msg: fmt.Sprintf("OAuth server status %d, %s", resp.StatusCode, string(p))}
	}
	m, err := url.ParseQuery(string(p))
	if err != nil {
		return nil, nil, RequestCredentialsError{StatusCode: resp.StatusCode, Header: resp.Header,
			Body: p, msg: err.Error()}
	}
	tokens := m["oauth_token"]
	if len(tokens) == 0 || tokens[0] == "" {
		return nil, nil, RequestCredentialsError{StatusCode: resp.StatusCode, Header: resp.Header,
			Body: p, msg: "oauth: token missing from server result"}
	}
	secrets := m["oauth_token_secret"]
	if len(secrets) == 0 { // allow "" as a valid secret.
		return nil, nil, RequestCredentialsError{StatusCode: resp.StatusCode, Header: resp.Header,
			Body: p, msg: "oauth: secret missing from server result"}
	}
	return &Credentials{Token: tokens[0], Secret: secrets[0]}, m, nil
}

// RequestTemporaryCredentials requests temporary credentials from the server.
// See http://tools.ietf.org/html/rfc5849#section-2.1 for information about
// temporary credentials.
func (c *Client) RequestTemporaryCredentials(client *http.Client, callbackURL string, additionalParams url.Values) (*Credentials, error) {
	ctx := context.WithValue(context.Background(), HTTPClient, client)
	return c.RequestTemporaryCredentialsContext(ctx, callbackURL, additionalParams)
}

// RequestTemporaryCredentialsContext uses Context to perform RequestTemporaryCredentials.
func (c *Client) RequestTemporaryCredentialsContext(ctx context.Context, callbackURL string, additionalParams url.Values) (*Credentials, error) {
	credentials, _, err := c.requestCredentials(ctx, c.TemporaryCredentialRequestURI,
		&request{method: c.TemporaryCredentialsMethod, form: additionalParams, callbackURL: callbackURL})
	return credentials, err
}

// RequestToken requests token credentials from the server. See
// http://tools.ietf.org/html/rfc5849#section-2.3 for information about token
// credentials.
func (c *Client) RequestToken(client *http.Client, temporaryCredentials *Credentials, verifier string) (*Credentials, url.Values, error) {
	ctx := context.WithValue(context.Background(), HTTPClient, client)
	return c.RequestTokenContext(ctx, temporaryCredentials, verifier)
}

// RequestTokenContext uses Context to perform RequestToken.
func (c *Client) RequestTokenContext(ctx context.Context, temporaryCredentials *Credentials, verifier string) (*Credentials, url.Values, error) {
	return c.requestCredentials(ctx, c.TokenRequestURI,
		&request{credentials: temporaryCredentials, method: c.TokenCredentailsMethod, verifier: verifier})
}

// RenewRequestCredentials requests new token credentials from the server.
// See http://wiki.oauth.net/w/page/12238549/ScalableOAuth#AccessTokenRenewal
// for information about access token renewal.
func (c *Client) RenewRequestCredentials(client *http.Client, credentials *Credentials, sessionHandle string) (*Credentials, url.Values, error) {
	ctx := context.WithValue(context.Background(), HTTPClient, client)
	return c.RenewRequestCredentialsContext(ctx, credentials, sessionHandle)
}

// RenewRequestCredentialsContext uses Context to perform RenewRequestCredentials.
func (c *Client) RenewRequestCredentialsContext(ctx context.Context, credentials *Credentials, sessionHandle string) (*Credentials, url.Values, error) {
	return c.requestCredentials(ctx, c.RenewCredentialRequestURI, &request{credentials: credentials, sessionHandle: sessionHandle})
}

// RequestTokenXAuth requests token credentials from the server using the xAuth protocol.
// See https://dev.twitter.com/oauth/xauth for information on xAuth.
func (c *Client) RequestTokenXAuth(client *http.Client, temporaryCredentials *Credentials, user, password string) (*Credentials, url.Values, error) {
	ctx := context.WithValue(context.Background(), HTTPClient, client)
	return c.RequestTokenXAuthContext(ctx, temporaryCredentials, user, password)
}

// RequestTokenXAuthContext uses Context to perform RequestTokenXAuth.
func (c *Client) RequestTokenXAuthContext(ctx context.Context, temporaryCredentials *Credentials, user, password string) (*Credentials, url.Values, error) {
	form := make(url.Values)
	form.Set("x_auth_mode", "client_auth")
	form.Set("x_auth_username", user)
	form.Set("x_auth_password", password)
	return c.requestCredentials(ctx, c.TokenRequestURI,
		&request{credentials: temporaryCredentials, method: c.TokenCredentailsMethod, form: form})
}

// AuthorizationURL returns the URL for resource owner authorization. See
// http://tools.ietf.org/html/rfc5849#section-2.2 for information about
// resource owner authorization.
func (c *Client) AuthorizationURL(temporaryCredentials *Credentials, additionalParams url.Values) string {
	params := make(url.Values)
	for k, vs := range additionalParams {
		params[k] = vs
	}
	params.Set("oauth_token", temporaryCredentials.Token)
	return c.ResourceOwnerAuthorizationURI + "?" + params.Encode()
}

// HTTPClient is the context key to use with context's
// WithValue function to associate an *http.Client value with a context.
var HTTPClient contextKey

type contextKey struct{}

func contextClient(ctx context.Context) *http.Client {
	if ctx != nil {
		if hc, ok := ctx.Value(HTTPClient).(*http.Client); ok && hc != nil {
			return hc
		}
	}
	return http.DefaultClient
}

// RequestCredentialsError is an error containing
// response information when requesting credentials.
type RequestCredentialsError struct {
	StatusCode int
	Header     http.Header
	Body       []byte
	msg        string
}

func (e RequestCredentialsError) Error() string {
	return e.msg
}
