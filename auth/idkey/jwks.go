package idkey

import (
	"crypto"
	"encoding/json"
	"errors"

	"github.com/square/go-jose"
)

// MarshalJWKSPublicKey returns the public key encoded as a JSON Web
// Key Set (JWKS). See
// https://tools.ietf.org/html/draft-ietf-jose-json-web-key-31 for
// more information.
//
// This method is usually used to encode the OAuth2 client public key
// when registering an OAuth2 client. The return value is used as the
// "jwks" metadata value (see
// http://openid.net/specs/openid-connect-registration-1_0.html#ClientMetadata).
func (k *IDKey) MarshalJWKSPublicKey() ([]byte, error) {
	return json.Marshal(jwkSet{Keys: []jose.JsonWebKey{{Key: k.Public()}}})
}

// UnmarshalJWKSPublicKey unmarshals an IDKey's public key from JWKS
// JSON bytes.
func UnmarshalJWKSPublicKey(data []byte) (crypto.PublicKey, error) {
	var jwks jwkSet
	if err := json.Unmarshal(data, &jwks); err != nil {
		return nil, err
	}
	if len(jwks.Keys) == 0 {
		return nil, errors.New("no keys in JWK set")
	}
	pubKey, ok := jwks.Keys[0].Key.(crypto.PublicKey)
	if !ok {
		return nil, errors.New("invalid public key")
	}
	return pubKey, nil
}

type jwkSet struct {
	Keys []jose.JsonWebKey `json:"keys"`
}
