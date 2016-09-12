// Package idkey deals with Sourcegarph identity keys (which identify
// a Sourcegraph instance or cluster).
package idkey

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"context"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/jwt"
)

// Default id key used in dev environments
var Default *IDKey

func init() {
	var err error
	Default, err = New([]byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAqvPY0u1RugKaCQRW+e5DSsrc6PzwRt7kGTS/jcmO36QK34WJ
5DbkaX+KaiZ6jivk4PYb9EbRrwQlwddT/sQo3liT25a8Bw19kYFd6fGHgIrbrIHQ
h2F3faBtKXvbfBct5ScV+gxrnSh5lXF3HAUIvUrE3U0Frub8IW2+IMhXBFWpk99s
dPuENbUK2i+z7bsQ79lT4o0omUH7GqqBEgUn+ooUjBZKuxNOU8Qe8zgfxyo4aAFL
VJdNz8tdsps2XOpPAyIIDz4QXblsej+Sj3vd/e+2AGgWHWkTeqF4560ebGZTkzDU
iuRIENJB6zvVEkqbN2bqK5ipUeukJrneHcpK4wIDAQABAoIBAQCcWlIg+HUbD24a
eSGjjUt2iHvrjAumhg1REHFyGLrXyI05SkWHuLzH0KKj23WTuomnRvDiRjNZQw3V
cD+eb4KBebohyIdIXApQnmVqpkEsS9QGvuQeLgK/n463tlRT9k8/mrP68okqP+6T
xCcQNXp8xnmvfdaI1TIc0OZnzVPo3YBHBaLW0mYYd4q375e3QbHTogGhudjMJrmt
hrebayDbWb9AzCUfJ+SXN1ye6e+zGsPJHtj5vyq2AM6WWZ6cy9M3fNoeesqndyUi
ogQt9/G94WXpgIBzbAJvdFvUU7VYDl0RjTyknU0eUeQE4ZVXkf8DifvIv6Y4uf9O
ENcvVL8BAoGBAOK3YwT8rKcs9KqO1jhhqAAs2odyO23E00a+2jpEBi3jFieuJTXG
++I8DUIJvzPz6kY+wv5sqW/iUk8+BKg5vDelHqfilsGgUpz+RXdeHgTvbrpP3hX5
RVTtySjvzAuFthLt17SeGXW6XWFeIWGNVX1hN7GNBw4GlRmfoHvDvyThAoGBAMEI
kikynj9oQn1OBZX2q8ohvhR3CQ/UfFi8x32eB71diU2RND84vystnapEG5k85vp/
m4jZrkMcFFQByM++7BRhaZ106fap2aC37W9mEcZx1Rq8SEnBYGI7XFJSqgo5PQjv
pxGIVN42aWMKyBkbc3lZBk74aiXGCFlCB2ggFCRDAoGAQdB81UjIkitRx2V5uJpY
29wpgCJgMChwMNxcm4d9x7phhxldwfPG0VEfhCkyMVHAk63Ki3Nd6JXp0Qku7ur7
waeUc6Yqn4D8GokR/2n6CvK60Sk4TmazgskYmWKreDjTt0EGWm9+8pcsXoSl6hzO
UXx0X32SK7crd2nIe8VCauECgYBXtJ48S7xtEOEM7/NHxPEfAR/NSpx6bdAlvXhi
kffwSVyGOtBjXVQ2uR4m65UilfJYpLw1fLpZ0ZtG5Byqj5PSWsRS/3kCUwAHryoZ
cGXpJXVvFVc+87boSxJScS7DQGiD8+eT5r6wzEYr8w0ho0BfRVzBastH6GeIrqCQ
5epg4wKBgQC5gONk50H+mkx/H5wjAz/3/BudIpKh8ommdjo+Zd3ZRrxoOeRnygLW
fKg4HdqJ8iSj0Nc/Ykruno/F+BSFJq2TqyTm/kxxpeG3FTdWIY8DEPICiNmoJD29
p0bK8eAQfTtdU67p7umWzYxZS1Nre0aWc6bdfES1g/xOr465k5iIEA==
-----END RSA PRIVATE KEY-----`))
	if err != nil {
		panic(err)
	}
}

func getPrivKeyBits() int {
	s := strings.TrimSpace(os.Getenv("ID_KEY_SIZE"))
	if s == "" {
		return 2048 // default
	}
	var err error
	n, err := strconv.Atoi(s)
	if err != nil {
		log.Fatalf("Invalid ID_KEY_SIZE: %s.", err)
	}
	if n < 1024 {
		log.Fatalf("ID_KEY_SIZE must be at least 1024 (got %d).", n)
	}
	return n
}

// IDKey holds a Sourcegraph identity key (which identifies a
// Sourcegraph instance or cluster).
type IDKey struct {
	key      *rsa.PrivateKey
	pemBytes []byte

	// ID is k's public key fingerprint, which can act as a client's
	// or server's identity.
	ID string
}

// New creates a new Sourcegraph identity key from PEM-encoded bytes
// of the form:
//
//  -----BEGIN RSA PRIVATE KEY-----
//  ...
//  -----END RSA PRIVATE KEY-----
func New(pem []byte) (*IDKey, error) {
	var k IDKey
	if err := k.UnmarshalText(pem); err != nil {
		return nil, err
	}
	return &k, nil
}

// FromString creates a new Sourcegraph identity key from a PEM-encoded
// string. It allows encoding the PEM data in base64, to make it easier to
// pass in env vars (which are often serialized/deserialized via buggy
// bash scripts).
func FromString(idKeyData string) (*IDKey, error) {
	if strings.HasPrefix(idKeyData, "base64:") {
		idKeyData = strings.TrimPrefix(idKeyData, "base64:")
		b, err := base64.StdEncoding.DecodeString(idKeyData)
		if err != nil {
			return nil, err
		}
		idKeyData = string(b)
	}
	return New([]byte(idKeyData))
}

// Private returns k's private key.
func (k *IDKey) Private() *rsa.PrivateKey { return k.key }

// Public returns k's public key.
func (k *IDKey) Public() *rsa.PublicKey { return k.key.Public().(*rsa.PublicKey) }

func (k *IDKey) MarshalText() ([]byte, error) {
	return k.pemBytes, nil
}

func (k *IDKey) UnmarshalText(data []byte) error {
	privBlock, _ := pem.Decode(data)
	if privBlock == nil {
		return errors.New("invalid private key PEM-encoded data")
	}
	if privBlock.Type != "RSA PRIVATE KEY" {
		return fmt.Errorf("invalid private key block type %q", privBlock.Type)
	}
	privKey, err := x509.ParsePKCS1PrivateKey(privBlock.Bytes)
	if err != nil {
		return err
	}
	k.key = privKey
	return k.precompute()
}

func (k *IDKey) precompute() error {
	k.key.Precompute()

	// Precompute bytes.
	k.pemBytes = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k.key)})

	fp, err := Fingerprint(k.Public())
	if err != nil {
		return err
	}
	k.ID = fp
	return nil
}

// Fingerprint returns the fingerprint used as the ID (generated from
// the ID key's public key).
func Fingerprint(pubKey crypto.PublicKey) (string, error) {
	// Precompute ID.
	b, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return base64.StdEncoding.EncodeToString(sum[:]), nil
}

func (k *IDKey) TokenSource(ctx context.Context, tokenURL string) oauth2.TokenSource {
	c := &jwt.Config{
		Email:      k.ID,
		Subject:    k.ID,
		PrivateKey: k.pemBytes,
		TokenURL:   tokenURL,
	}
	return c.TokenSource(ctx)
}

type contextKey int

const (
	idKeyKey contextKey = iota
)

// FromContext returns the Sourcegraph identity key from the context,
// or nil if none is set.
func FromContext(ctx context.Context) *IDKey {
	idkey, _ := ctx.Value(idKeyKey).(*IDKey)
	return idkey
}

// NewContext returns a child context with the given ID key.
func NewContext(ctx context.Context, idkey *IDKey) context.Context {
	return context.WithValue(ctx, idKeyKey, idkey)
}
