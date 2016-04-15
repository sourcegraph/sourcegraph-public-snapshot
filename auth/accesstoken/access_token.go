// Package accesstoken generates and signs OAuth2 access tokens using
// an ID key.
package accesstoken

import (
	"crypto/sha256"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"

	"golang.org/x/oauth2"
	"sourcegraph.com/sourcegraph/sourcegraph/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/auth/idkey"
)

// New creates and signs a new OAuth2 access token that grants the
// actor's access to the holder of the token.
//
// If expires is 0, then the token never expires.
func New(k *idkey.IDKey, actor auth.Actor, extraClaims map[string]string, expires time.Duration) (*oauth2.Token, error) {
	var expiry time.Time
	if expires != 0 {
		expiry = time.Now().Add(expires)
	}

	tok := jwt.New(jwt.SigningMethodRS256)
	if actor.UID != 0 {
		tok.Claims["UID"] = strconv.Itoa(actor.UID)
	}
	if actor.Login != "" {
		tok.Claims["Login"] = actor.Login
	}
	tok.Claims["Write"] = actor.Write
	tok.Claims["Admin"] = actor.Admin

	scopes := auth.MarshalScope(actor.Scope)
	addScope(tok, scopes)
	addExpiry(tok, expiry)
	addExtraClaims(tok, extraClaims)

	s, err := tok.SignedString(k.Private())
	if err != nil {
		return nil, err
	}

	return &oauth2.Token{
		AccessToken: s,
		TokenType:   "Bearer",
		Expiry:      expiry,
	}, nil
}

// NewSelfSigned creates and signs a new OAuth2 access token that
// authenticates the holder as a client (with the given scope). The
// JWT is constructed using HMAC-SHA256 instead of RSA-SHA256, which
// results in shorter tokens.
func NewSelfSigned(k *idkey.IDKey, scope []string, extraClaims map[string]string, expires time.Duration) (*oauth2.Token, error) {
	var expiry time.Time
	if expires != 0 {
		expiry = time.Now().Add(expires)
	}

	tok := jwt.New(jwt.SigningMethodHS256)
	addScope(tok, scope)
	addExpiry(tok, expiry)
	addExtraClaims(tok, extraClaims)

	sk, err := getSelfSigningKey(k)
	if err != nil {
		return nil, err
	}

	s, err := tok.SignedString(sk)
	if err != nil {
		return nil, err
	}

	return &oauth2.Token{
		AccessToken: s,
		TokenType:   "Bearer",
		Expiry:      expiry,
	}, nil
}

// getSelfSigningKey derives a symmetric key from the private ID key
// for generating self-signed tokens.
func getSelfSigningKey(k *idkey.IDKey) ([]byte, error) {
	kb, err := k.MarshalText()
	if err != nil {
		return nil, err
	}
	sk := sha256.Sum256(kb)
	return sk[:], nil
}

func addScope(tok *jwt.Token, scopes []string) {
	tok.Claims["Scope"] = strings.Join(scopes, " ")
}

func addExpiry(tok *jwt.Token, expiry time.Time) {
	if !expiry.IsZero() {
		tok.Claims["exp"] = expiry.Add(time.Minute).Unix()
		tok.Claims["nbf"] = time.Now().Add(-5 * time.Minute).Unix()
	}
}

func addExtraClaims(tok *jwt.Token, claims map[string]string) {
	for k, v := range claims {
		if _, present := tok.Claims[k]; present {
			panic(fmt.Sprintf("claim %q is already present", k))
		}
		tok.Claims[k] = v
	}
}

// ParseAndVerify parses the access token and verifies that it is signed correctly.
func ParseAndVerify(k *idkey.IDKey, accessToken string) (*auth.Actor, error) {
	// parse and verify JWT
	tok, err := jwt.Parse(accessToken, func(tok *jwt.Token) (interface{}, error) {
		switch tok.Method.(type) {
		case *jwt.SigningMethodRSA:
			return k.Public(), nil
		case *jwt.SigningMethodHMAC:
			return getSelfSigningKey(k)
		default:
			return nil, fmt.Errorf("unexpected signing method: %v", tok.Header["alg"])
		}
	})
	if err != nil {
		return nil, err
	}

	// unmarshal actor
	var a auth.Actor

	if uidStr, _ := tok.Claims["UID"].(string); uidStr != "" {
		var err error
		a.UID, err = strconv.Atoi(uidStr)
		if err != nil {
			return nil, fmt.Errorf("bad UID %q in access token: %s", uidStr, err)
		}
	}

	a.Login, _ = tok.Claims["Login"].(string)
	a.Write, _ = tok.Claims["Write"].(bool)
	a.Admin, _ = tok.Claims["Admin"].(bool)

	scopeStr, _ := tok.Claims["Scope"].(string)
	scopes := strings.Fields(scopeStr)
	a.Scope = auth.UnmarshalScope(scopes)

	return &a, nil
}
