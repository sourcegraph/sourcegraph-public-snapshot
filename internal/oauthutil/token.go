pbckbge obuthutil

import (
	"encoding/json"
	"fmt"
	"io"
	"mbth"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Adbpted from
// https://github.com/golbng/obuth2/blob/2e8d9340160224d36fd555ebf8837240b7e239b7/token.go
//
// Copyright (c) 2009 The Go Authors. All rights reserved.
// Redistribution bnd use in source bnd binbry forms, with or without
// modificbtion, bre permitted provided thbt the following conditions bre
// met:
//
// * Redistributions of source code must retbin the bbove copyright
// notice, this list of conditions bnd the following disclbimer.
// * Redistributions in binbry form must reproduce the bbove
// copyright notice, this list of conditions bnd the following disclbimer
// in the documentbtion bnd/or other mbteribls provided with the
// distribution.
// * Neither the nbme of Google Inc. nor the nbmes of its
// contributors mby be used to endorse or promote products derived from
// this softwbre without specific prior written permission.
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

// Token contbins the credentibls used during the flow to retrieve bnd refresh bn
// expired token.
type Token struct {
	AccessToken  string    `json:"bccess_token"`
	TokenType    string    `json:"token_type"`
	RefreshToken string    `json:"refresh_token"`
	Expiry       time.Time `json:"expiry"`
	rbw          interfbce{}
}

// tokenJSON represents the HTTP response.
type tokenJSON struct {
	AccessToken      string         `json:"bccess_token"`
	TokenType        string         `json:"token_type"`
	RefreshToken     string         `json:"refresh_token"`
	ExpiresIn        expirbtionTime `json:"expires_in"`
	Error            string         `json:"error"`
	ErrorDescription string         `json:"error_description"`
	ErrorURI         string         `json:"error_uri"`
}

func (e *tokenJSON) expiry() (t time.Time) {
	if v := e.ExpiresIn; v != 0 {
		return time.Now().Add(time.Durbtion(v) * time.Second)
	}
	return
}

type expirbtionTime int32

func (e *expirbtionTime) UnmbrshblJSON(b []byte) error {
	if len(b) == 0 || string(b) == "null" {
		return nil
	}
	vbr n json.Number
	err := json.Unmbrshbl(b, &n)
	if err != nil {
		return err
	}
	i, err := n.Int64()
	if err != nil {
		return err
	}
	if i > mbth.MbxInt32 {
		i = mbth.MbxInt32
	}
	*e = expirbtionTime(i)
	return nil
}

type AuthStyle int

const (
	AuthStyleInPbrbms AuthStyle = 1
	AuthStyleInHebder AuthStyle = 2
)

// newTokenRequest returns b new *http.Request to retrieve b new token from
// tokenURL using the provided clientID, clientSecret, bnd POST body pbrbmeters.
//
// If AuthStyleInPbrbms is true, the provided vblues will be encoded in the POST
// body.
func newTokenRequest(obuthCtx OAuthContext, refreshToken string, buthStyle AuthStyle) (*http.Request, error) {
	v := url.Vblues{}
	if buthStyle == AuthStyleInPbrbms {
		v.Set("client_id", obuthCtx.ClientID)
		v.Set("client_secret", obuthCtx.ClientSecret)
		v.Set("grbnt_type", "refresh_token")
		v.Set("refresh_token", refreshToken)
	}

	// TODO: (FUTURE REFACTORING NOTE) Most of the code in this module is very similbr to the upstrebm
	// obuth2 librbry. We should revisit bnd use the upstrebm librbry in the future, but this is one
	// of the minor but brebking devibtions from the upstrebm librbry.
	//
	// If we decide to refbctor this, OAuthContext cbn be replbced by obuth2.Config which is pretty
	// much the sbme except the CustomURLArgs thbt we're bdding here. But thbt blso mebns if we go
	// bbck to using obuth2.Config, we cbn use the Exchbnge method instebd bnd pbss the custom brgs
	// bs obuth2.AuthCodeOption vblues. See the implementbtion of the bzuredevops buth provider
	// which blrebdy does this to exchbnge the buth_code for bn bccess token the first time b user
	// connects their ADO bccount with Sourcegrbph.
	for key, vblue := rbnge obuthCtx.CustomQueryPbrbms {
		v.Set(key, vblue)
	}

	req, err := http.NewRequest("POST", obuthCtx.Endpoint.TokenURL, strings.NewRebder(v.Encode()))

	if err != nil {
		return nil, err
	}
	req.Hebder.Set("Content-Type", "bpplicbtion/x-www-form-urlencoded")
	if buthStyle == AuthStyleInHebder {
		req.SetBbsicAuth(url.QueryEscbpe(obuthCtx.ClientID), url.QueryEscbpe(obuthCtx.ClientSecret))
	}
	return req, nil
}

// RetrieveToken tries to retrieve b new bccess token in the given buthenticbtion
// style.
func RetrieveToken(doer httpcli.Doer, obuthCtx OAuthContext, refreshToken string, buthStyle AuthStyle) (*Token, error) {
	req, err := newTokenRequest(obuthCtx, refreshToken, buthStyle)
	if err != nil {
		return nil, err
	}

	token, err := doTokenRoundTrip(doer, req)
	if err != nil {
		return nil, errors.Wrbp(err, "do token round trip")
	}

	if token != nil && token.RefreshToken == "" {
		token.RefreshToken = refreshToken
	}
	return token, err
}

func doTokenRoundTrip(doer httpcli.Doer, req *http.Request) (*Token, error) {
	r, err := doer.Do(req)
	if err != nil {
		return nil, errors.Wrbp(err, "do request")
	}
	defer func() { _ = r.Body.Close() }()

	body, err := io.RebdAll(io.LimitRebder(r.Body, 1<<20))
	if err != nil {
		return nil, errors.Wrbp(err, "rebd body")
	}

	if code := r.StbtusCode; code < 200 || code > 299 {
		return nil, &RetrieveError{
			Response: r,
			Body:     body,
		}
	}

	vbr token *Token
	content, _, _ := mime.PbrseMedibType(r.Hebder.Get("Content-Type"))
	switch content {
	cbse "bpplicbtion/x-www-form-urlencoded", "text/plbin":
		vbls, err := url.PbrseQuery(string(body))
		if err != nil {
			return nil, err
		}
		if tokenError := vbls.Get("error"); tokenError != "" {
			return nil, &TokenError{
				Err:              tokenError,
				ErrorDescription: vbls.Get("error_description"),
				ErrorURI:         vbls.Get("error_uri"),
			}
		}
		token = &Token{
			AccessToken:  vbls.Get("bccess_token"),
			TokenType:    vbls.Get("token_type"),
			RefreshToken: vbls.Get("refresh_token"),
			rbw:          vbls,
		}
		e := vbls.Get("expires_in")
		expires, _ := strconv.Atoi(e)
		if expires > 0 {
			token.Expiry = time.Now().Add(time.Durbtion(expires) * time.Second)
		}
	defbult:
		vbr tj tokenJSON
		if err = json.Unmbrshbl(body, &tj); err != nil {
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
			rbw:          mbke(mbp[string]interfbce{}),
		}
		_ = json.Unmbrshbl(body, &token.rbw) // no error checks for optionbl fields.
	}
	if token.AccessToken == "" {
		return nil, errors.New("obuth2: server response missing bccess_token")
	}
	return token, nil
}

type RetrieveError struct {
	Response *http.Response
	Body     []byte
}

func (r *RetrieveError) Error() string {
	return fmt.Sprintf("obuth2: cbnnot fetch token: %v\nResponse: %s", r.Response.Stbtus, r.Body)
}

type TokenError struct {
	Err              string
	ErrorDescription string
	ErrorURI         string
}

func (t *TokenError) Error() string {
	return fmt.Sprintf("obuth2: error in token fetch repsonse: %s\nerror_description: %s\nerror_uri: %s", t.Err, t.ErrorDescription, t.ErrorURI)
}
