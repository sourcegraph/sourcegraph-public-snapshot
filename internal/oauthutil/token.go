package oauthutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
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
func newTokenRequest(oauthCtx OAuthContext, refreshToken string) (*http.Request, error) {
	requestBody := struct {
		RefreshToken string `json:"refresh_token"`
		GrantType    string `json:"grant_type"`
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
	}{
		refreshToken,
		"refresh_token",
		oauthCtx.ClientID,
		oauthCtx.ClientSecret,
	}

	payload := new(bytes.Buffer)
	json.NewEncoder(payload).Encode(requestBody)

	req, err := http.NewRequest("POST", oauthCtx.Endpoint.TokenURL, payload)

	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// RetrieveToken tries to retrieve a new access token in the given authentication
// style.
func RetrieveToken(doer httpcli.Doer, oauthCtx OAuthContext, refreshToken string) (*auth.OAuthBearerToken, error) {
	req, err := newTokenRequest(oauthCtx, refreshToken)
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

func doTokenRoundTrip(doer httpcli.Doer, req *http.Request) (*auth.OAuthBearerToken, error) {
	r, err := doer.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "do request")
	}

	defer func() { _ = r.Body.Close() }()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Wrap(err, "read body")
	}

	if code := r.StatusCode; code < 200 || code > 299 {
		return nil, &RetrieveError{
			Response: r,
			Body:     body,
		}
	}

	var token *auth.OAuthBearerToken
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
	token = &auth.OAuthBearerToken{
		AccessToken:  tj.AccessToken,
		TokenType:    tj.TokenType,
		RefreshToken: tj.RefreshToken,
		Expiry:       tj.expiry(),
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
