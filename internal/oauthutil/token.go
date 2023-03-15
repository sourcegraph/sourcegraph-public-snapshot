package oauthutil

import (
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

// Adapted from
// https://github.com/golang/oauth2/blob/2e8d9340160224d36fd555eaf8837240a7e239a7/token.go
//
// Copyright (c) 2009 The Go Authors. All rights reserved.
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
// * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
// * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
// * Neither the name of Google Inc. nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// Token contains the credentials used during the flow to retrieve and refresh an
// expired token.
type Token struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	RefreshToken string    `json:"refresh_token"`
	Expiry       time.Time `json:"expiry"`
	raw          interface{}
}

// tokenJSON represents the HTTP response.
type tokenJSON struct {
	AccessToken      string         `json:"access_token"`
	TokenType        string         `json:"token_type"`
	RefreshToken     string         `json:"refresh_token"`
	ExpiresIn        expirationTime `json:"expires_in"`
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

type AuthStyle int

const (
	AuthStyleInParams AuthStyle = 1
	AuthStyleInHeader AuthStyle = 2
)

// newTokenRequest returns a new *http.Request to retrieve a new token from
// tokenURL using the provided clientID, clientSecret, and POST body parameters.
//
// If AuthStyleInParams is true, the provided values will be encoded in the POST
// body.
func newTokenRequest(oauthCtx OAuthContext, refreshToken string, authStyle AuthStyle) (*http.Request, error) {
	v := url.Values{}
	if authStyle == AuthStyleInParams {
		v.Set("client_id", oauthCtx.ClientID)
		v.Set("client_secret", oauthCtx.ClientSecret)
		v.Set("grant_type", "refresh_token")
		v.Set("refresh_token", refreshToken)
	}

	// TODO: (FUTURE REFACTORING NOTE) Most of the code in this module is very similar to the upstream
	// oauth2 library. We should revisit and use the upstream library in the future, but this is one
	// of the minor but breaking deviations from the upstream library.
	//
	// If we decide to refactor this, OAuthContext can be replaced by oauth2.Config which is pretty
	// much the same except the CustomURLArgs that we're adding here. But that also means if we go
	// back to using oauth2.Config, we can use the Exchange method instead and pass the custom args
	// as oauth2.AuthCodeOption values. See the implementation of the azuredevops auth provider
	// which already does this to exchange the auth_code for an access token the first time a user
	// connects their ADO account with Sourcegraph.
	for key, value := range oauthCtx.CustomQueryParams {
		v.Set(key, value)
	}

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

// RetrieveToken tries to retrieve a new access token in the given authentication
// style.
func RetrieveToken(doer httpcli.Doer, oauthCtx OAuthContext, refreshToken string, authStyle AuthStyle) (*Token, error) {
	req, err := newTokenRequest(oauthCtx, refreshToken, authStyle)
	if err != nil {
		return nil, err
	}

	token, err := doTokenRoundTrip(doer, req)
	if err != nil {
		return nil, errors.Wrap(err, "do token round trip")
	}

	if token != nil && token.RefreshToken == "" {
		token.RefreshToken = refreshToken
	}
	return token, err
}

func doTokenRoundTrip(doer httpcli.Doer, req *http.Request) (*Token, error) {
	r, err := doer.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "do request")
	}
	defer func() { _ = r.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return nil, errors.Wrap(err, "read body")
	}

	if code := r.StatusCode; code < 200 || code > 299 {
		return nil, &RetrieveError{
			Response: r,
			Body:     body,
		}
	}

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
			raw:          vals,
		}
		e := vals.Get("expires_in")
		expires, _ := strconv.Atoi(e)
		if expires > 0 {
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
			raw:          make(map[string]interface{}),
		}
		_ = json.Unmarshal(body, &token.raw) // no error checks for optional fields.
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
