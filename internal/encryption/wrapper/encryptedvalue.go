package wrapper

import (
	"encoding/base64"
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type StorableEncryptedKey struct {
	// Mechanism is used to store the way the value was encrypted, like envelope AES encryption.
	Mechanism string
	// KeyName is the name of the key used to encrypt the value.
	KeyName string
	// WrappedKey optionally contains an encrypted key that was used to encrypt the ciphertext.
	WrappedKey []byte
	// Ciphertext contains the encrypted value.
	Ciphertext []byte
	// Nonce is an additional value to store a nonce that was used to encrypt Ciphertext.
	Nonce []byte
}

func FromCiphertext(ciphertext []byte) (*StorableEncryptedKey, error) {
	data, err := base64.StdEncoding.DecodeString(string(ciphertext))
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode ciphertext")
	}
	var sek StorableEncryptedKey
	if err := json.Unmarshal(data, &sek); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal ciphertext into StorableEncryptedKey")
	}

	return &sek, nil
}

func (sek *StorableEncryptedKey) Serialize() ([]byte, error) {
	jsonKey, err := json.Marshal(sek)
	if err != nil {
		return nil, errors.Wrap(err, "marshaling StorableEncryptedKey to JSON")
	}
	return []byte(base64.StdEncoding.EncodeToString(jsonKey)), nil
}
