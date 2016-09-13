package auth

import (
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
)

// ActiveIDKey is the key used by New and ParseAndVerify.
// The default is only to be used in dev environments.
var ActiveIDKey *IDKey

func init() {
	var err error
	ActiveIDKey, err = NewIDKey([]byte(`-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----
`))
	if err != nil {
		panic(err)
	}
}

// IDKey holds a Sourcegraph identity key (which identifies a
// Sourcegraph instance or cluster).
type IDKey struct {
	rsaKey  *rsa.PrivateKey
	hmacKey []byte
}

// New creates a new Sourcegraph identity key from PEM-encoded bytes
// of the form:
//
//  -----BEGIN RSA PRIVATE KEY-----
//  ...
//  -----END RSA PRIVATE KEY-----
func NewIDKey(data []byte) (*IDKey, error) {
	privBlock, _ := pem.Decode(data)
	if privBlock == nil {
		return nil, errors.New("invalid private key PEM-encoded data")
	}
	if privBlock.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("invalid private key block type %q", privBlock.Type)
	}
	rsaKey, err := x509.ParsePKCS1PrivateKey(privBlock.Bytes)
	if err != nil {
		return nil, err
	}
	rsaKey.Precompute()

	sk := sha256.Sum256(data)
	hmacKey := sk[:]

	return &IDKey{
		rsaKey:  rsaKey,
		hmacKey: hmacKey,
	}, nil
}

// FromString creates a new Sourcegraph identity key from a PEM-encoded
// string. It allows encoding the PEM data in base64, to make it easier to
// pass in env vars (which are often serialized/deserialized via buggy
// bash scripts).
func FromString(idKeyData string) (*IDKey, error) {
	if strings.HasPrefix(idKeyData, "base64:") {
		b, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(idKeyData, "base64:"))
		if err != nil {
			return nil, err
		}
		return NewIDKey(b)
	}
	return NewIDKey([]byte(idKeyData))
}
