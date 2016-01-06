package secure

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"github.com/square/go-jose"
	"gopkg.in/yaml.v2"
)

type Secure struct {
	Checksum    string        `yaml:"checksum"`
	Environment MapEqualSlice `yaml:"environment"`
}

// Parse parses and returns the secure section of the
// yaml file as plaintext parameters.
func Parse(in, privKey string) (*Secure, error) {
	// unarmshal the private key from PEM
	rsaPrivKey, err := decodePrivateKey(privKey)
	if err != nil {
		return nil, err
	}

	// decrypt the Yaml file
	plain, err := decrypt(in, rsaPrivKey)
	if err != nil {
		return nil, err
	}

	// unmarshal the yaml
	out := &Secure{}
	err = yaml.Unmarshal(plain, out)
	return out, err
}

// decrypt decrypts a JOSE string and returns the
// plaintext value.
func decrypt(secret string, privKey *rsa.PrivateKey) ([]byte, error) {
	object, err := jose.ParseEncrypted(secret)
	if err != nil {
		return nil, err
	}
	return object.Decrypt(privKey)
}

// decodePrivateKey is a helper function that unmarshals a PEM
// bytes to an RSA Private Key
func decodePrivateKey(privateKey string) (*rsa.PrivateKey, error) {
	derBlock, _ := pem.Decode([]byte(privateKey))
	return x509.ParsePKCS1PrivateKey(derBlock.Bytes)
}

// encodePrivateKey is a helper function that marshals an RSA
// Private Key to a PEM encoded file.
func encodePrivateKey(privkey *rsa.PrivateKey) string {
	privateKeyMarshaled := x509.MarshalPKCS1PrivateKey(privkey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Headers: nil, Bytes: privateKeyMarshaled})
	return string(privateKeyPEM)
}

// encrypt encrypts a plaintext variable using JOSE with
// RSA_OAEP and A128GCM algorithms.
func encrypt(text string, pubKey *rsa.PublicKey) (string, error) {
	var encrypted string
	var plaintext = []byte(text)

	// Creates a new encrypter using defaults
	encrypter, err := jose.NewEncrypter(jose.RSA_OAEP, jose.A128GCM, pubKey)
	if err != nil {
		return encrypted, err
	}
	// Encrypts the plaintext value and serializes
	// as a JOSE string.
	object, err := encrypter.Encrypt(plaintext)
	if err != nil {
		return encrypted, err
	}
	return object.CompactSerialize()
}
