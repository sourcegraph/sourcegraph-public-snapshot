package auth

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// New creates and signs a new OAuth2 access token that grants the
// actor's access to the holder of the token. The given scopes are
// applied as well. The retuned token is assumed to be
// public and must not include any secret data.
func NewAccessToken(actor *Actor, scopes []string, expiryDuration time.Duration) (string, error) {
	tok := jwt.New(jwt.SigningMethod(jwt.SigningMethodHS256))

	if actor != nil {
		if actor.UID != 0 {
			tok.Claims["UID"] = strconv.Itoa(actor.UID)
		}
		if actor.Login != "" {
			tok.Claims["Login"] = actor.Login
		}
		tok.Claims["Write"] = actor.Write
		tok.Claims["Admin"] = actor.Admin
		scopes = append(scopes, marshalScope(actor.Scope)...)
	}

	tok.Claims["Scope"] = strings.Join(scopes, " ")

	if expiryDuration != 0 {
		expiry := time.Now().Add(expiryDuration)
		tok.Claims["exp"] = expiry.Add(time.Minute).Unix()
		tok.Claims["nbf"] = time.Now().Add(-5 * time.Minute).Unix()
	}

	s, err := tok.SignedString(ActiveIDKey.hmacKey)
	if err != nil {
		return "", err
	}

	return s, nil
}

// ParseAndVerify parses the access token and verifies that it is signed correctly.
func ParseAndVerify(accessToken string) (*Actor, error) {
	// parse and verify JWT
	tok, err := jwt.Parse(accessToken, func(tok *jwt.Token) (interface{}, error) {
		switch tok.Method.(type) {
		case *jwt.SigningMethodRSA: // legacy
			return ActiveIDKey.rsaKey.Public(), nil
		case *jwt.SigningMethodHMAC:
			return ActiveIDKey.hmacKey, nil
		default:
			return nil, fmt.Errorf("unexpected signing method: %v", tok.Header["alg"])
		}
	})
	if err != nil {
		return nil, err
	}

	// unmarshal actor
	var a Actor

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
	a.Scope = unmarshalScope(scopes)

	return &a, nil
}

func unmarshalScope(scope []string) map[string]bool {
	scopeMap := make(map[string]bool)
	for _, s := range scope {
		scopeMap[s] = true
	}
	return scopeMap
}

func marshalScope(scopeMap map[string]bool) []string {
	scope := make([]string, 0)
	for s, ok := range scopeMap {
		if !ok {
			continue
		}
		scope = append(scope, s)
	}
	return scope
}
