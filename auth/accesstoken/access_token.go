// Package accesstoken generates and signs OAuth2 access tokens using
// an ID key.
package accesstoken

import (
	"crypto"
	"crypto/sha256"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/auth/idkey"
	"src.sourcegraph.com/sourcegraph/fed"
	"src.sourcegraph.com/sourcegraph/svc"
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
	AddDomain(tok, actor.Domain)
	if actor.ClientID != "" {
		tok.Claims["ClientID"] = actor.ClientID
	}
	addScope(tok, actor.Scope)
	addExpiry(tok, expiry)
	addExtraClaims(tok, extraClaims)

	tok.Claims["kid"] = k.ID
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

func addScope(tok *jwt.Token, scope []string) {
	tok.Claims["Scope"] = strings.Join(scope, " ")
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

// ParseAndVerify parses tok's access token and verifies that it is
// signed correctly (and has the correct client ID). An unverified
// (potentially spoofed) actor is returned even if verification
// failed. Callers must check that the error is nil before assuming
// that the actor is verified.
func ParseAndVerify(ctx context.Context, accessToken string) (a *auth.Actor, allClaims map[string]interface{}, err error) {
	var err2 error
	idKey := idkey.FromContext(ctx)
	tok, err := parseToken(ctx, idKey, accessToken)
	if tok != nil {
		a, err2 = newActorWithVerifiedClaims(idKey, tok)
		allClaims = tok.Claims
	}
	if err == nil {
		err = err2
	}
	return a, allClaims, err
}

// UnsafeParseNoVerify parses tok's access token but DOES NOT verify
// its signature. This is unsafe! Someone could spoof the access
// token.
func UnsafeParseNoVerify(accessToken string) (*jwt.Token, error) {
	return jwt.Parse(accessToken, func(*jwt.Token) (interface{}, error) {
		return nil, nil
	})
}

// PublicKeyUnavailableError occurs when an access token (JWT) is
// signed by an external server. The current server can't verify it,
// but it can ignore the error and pass the access token along on
// outgoing requests. It's important that the server verify its own
// access tokens, but it can treat access tokens from other servers as
// opaque values.
type PublicKeyUnavailableError struct {
	ID  string // ID of server that signed the token
	Err error  // underlying error
}

func (e *PublicKeyUnavailableError) Error() string {
	s := fmt.Sprintf("JWT was signed by unavailable public key %q", e.ID)
	if e.Err != nil {
		s += fmt.Sprintf(" (reason: %s)", e.Err)
	}
	return s
}

func parseToken(ctx context.Context, idKey *idkey.IDKey, tokStr string) (*jwt.Token, error) {
	var innerErr error

	isSelfSigned := false

	// Unwrap and verify JWT.
	tok, err := jwt.Parse(tokStr, func(tok *jwt.Token) (interface{}, error) {
		if _, ok := tok.Method.(*jwt.SigningMethodHMAC); ok {
			// Assume token is self signed.
			isSelfSigned = true
			return getSelfSigningKey(idKey)
		}

		if _, ok := tok.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", tok.Header["alg"])
		}

		clientID, _ := tok.Claims["kid"].(string)
		if clientID == idKey.ID {
			return idKey.Public(), nil
		}

		if !fed.Config.IsRoot {
			rootPubKey := getRootPublicKey(ctx, clientID)
			if rootPubKey != nil {
				return rootPubKey, nil
			}
		}

		pubKey, err := getClientPublicKey(ctx, clientID)
		if pubKey == nil {
			innerErr = &PublicKeyUnavailableError{ID: clientID, Err: err}
			return nil, innerErr
		}
		return pubKey, err
	})
	if innerErr != nil {
		err = innerErr
	}

	// By convention, self-signed tokens do not include a "kid" or
	// "ClientID" claim because they are redundant (their only
	// possible values are the signer's/parser's own kid/ClientID
	// values). But set these explicitly if parsing succeeded so that
	// callers of this function can use those values.
	if err == nil && isSelfSigned {
		tok.Claims["ClientID"] = idKey.ID
		tok.Claims["kid"] = idKey.ID
	}

	return tok, err
}

func getClientPublicKey(ctx context.Context, clientID string) (crypto.PublicKey, error) {
	regClient, err := svc.RegisteredClients(ctx).Get(ctx, &sourcegraph.RegisteredClientSpec{ID: clientID})
	if err != nil {
		return nil, err
	}

	// Get the client's registered public key.
	if regClient.JWKS == "" {
		return nil, fmt.Errorf("client ID %s has no JWKS", clientID)
	}
	pubKey, err := idkey.UnmarshalJWKSPublicKey([]byte(regClient.JWKS))
	if err != nil {
		return nil, fmt.Errorf("parsing client ID %s JWKS public key: %s", clientID, err)
	}
	return pubKey, nil
}

func getRootPublicKey(ctx context.Context, clientID string) crypto.PublicKey {
	rootKey := idkey.RootPubKey(ctx)
	if rootKey != nil && rootKey.ID == clientID {
		return rootKey.Key
	}
	return nil
}

func newActorWithVerifiedClaims(idKey *idkey.IDKey, tok *jwt.Token) (*auth.Actor, error) {
	// Retrieve claims.
	var a auth.Actor
	var err error

	uidStr, _ := tok.Claims["UID"].(string)

	if uidStr != "" {
		a.UID, err = strconv.Atoi(uidStr)
		if err != nil {
			return nil, fmt.Errorf("bad UID %q in access token: %s", uidStr, err)
		}
	}

	a.Login, _ = tok.Claims["Login"].(string)
	a.Domain, _ = tok.Claims["Domain"].(string)
	a.ClientID, _ = tok.Claims["ClientID"].(string)

	scopeStr, _ := tok.Claims["Scope"].(string)
	a.Scope = strings.Fields(scopeStr)

	return &a, nil
}

// AddDomain adds a domain claim to the JWT.
func AddDomain(tok *jwt.Token, domain string) {
	tok.Claims["Domain"] = domain
}
