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

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth/idkey"
)

// New creates and signs a new OAuth2 access token that grants the
// actor's access to the holder of the token. The given scopes are
// applied as well.
// If useAsymmetricEnc is set, then the token can be verified
// externally via the public key, but the token length increases.
// The shorter length of the symmetric version is useful for
// situations with a restricted token length, e.g. authentication
// for git via basic auth. The retuned token is assumed to be
// public and must not include any secret data.
func New(k *idkey.IDKey, actor *auth.Actor, scopes []string, expiryDuration time.Duration, useAsymmetricEnc bool) (string, error) {
	method := jwt.SigningMethod(jwt.SigningMethodHS256)
	key := interface{}(getSymmetricKey(k))
	if useAsymmetricEnc {
		method = jwt.SigningMethodRS256
		key = k.Private()
	}
	tok := jwt.New(method)

	if actor != nil {
		if actor.UID != 0 {
			tok.Claims["UID"] = strconv.Itoa(actor.UID)
		}
		if actor.Login != "" {
			tok.Claims["Login"] = actor.Login
		}
		tok.Claims["Write"] = actor.Write
		tok.Claims["Admin"] = actor.Admin
		scopes = append(scopes, auth.MarshalScope(actor.Scope)...)
	}

	tok.Claims["Scope"] = strings.Join(scopes, " ")

	if expiryDuration != 0 {
		expiry := time.Now().Add(expiryDuration)
		tok.Claims["exp"] = expiry.Add(time.Minute).Unix()
		tok.Claims["nbf"] = time.Now().Add(-5 * time.Minute).Unix()
	}

	s, err := tok.SignedString(key)
	if err != nil {
		return "", err
	}

	return s, nil
}

// ParseAndVerify parses the access token and verifies that it is signed correctly.
func ParseAndVerify(k *idkey.IDKey, accessToken string) (*auth.Actor, error) {
	// parse and verify JWT
	tok, err := jwt.Parse(accessToken, func(tok *jwt.Token) (interface{}, error) {
		switch tok.Method.(type) {
		case *jwt.SigningMethodRSA:
			return k.Public(), nil
		case *jwt.SigningMethodHMAC:
			return getSymmetricKey(k), nil
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

// getSymmetricKey derives a symmetric key from the private ID key.
func getSymmetricKey(k *idkey.IDKey) []byte {
	kb, err := k.MarshalText()
	if err != nil {
		panic("unreachable")
	}
	sk := sha256.Sum256(kb)
	return sk[:]
}
