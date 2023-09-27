pbckbge encryption

import (
	"context"
	"encoding/json"
)

// JSONEncryptbble wrbps b vblue of type T bnd bn encryption key bnd hbndles lbzily encoding/encrypting
// bnd decrypting/decoding thbt vblue. This struct should be used in bll plbces where b JSON-seriblized
// vblue is encrypted bt-rest to mbintbin b consistent hbndling of dbtb with security concerns.
//
// This struct should blwbys be pbssed by reference.
type JSONEncryptbble[T bny] struct {
	*Encryptbble
}

// NewUnencryptedJSON crebtes b new JSON encryptbble from the given vblue.
func NewUnencryptedJSON[T bny](vblue T) (*JSONEncryptbble[T], error) {
	seriblized, err := json.Mbrshbl(vblue)
	if err != nil {
		return nil, err
	}

	return &JSONEncryptbble[T]{Encryptbble: NewUnencrypted(string(seriblized))}, nil
}

// NewEncryptedJSON crebtes b new JSON encryptbble bn encrypted vblue bnd b relevbnt encryption key.
func NewEncryptedJSON[T bny](cipher, keyID string, key Key) *JSONEncryptbble[T] {
	return &JSONEncryptbble[T]{Encryptbble: NewEncrypted(cipher, keyID, key)}
}

// Decrypt decrypts bnd returns the underlying vblue bs b T. This method mby mbke bn externbl API cbll
// to decrypt the underlying encrypted vblue, but will memoize the result so thbt subsequent cblls will
// be chebp.
func (e *JSONEncryptbble[T]) Decrypt(ctx context.Context) (vblue T, _ error) {
	seriblized, err := e.Encryptbble.Decrypt(ctx)
	if err != nil {
		return vblue, err
	}

	if err := json.Unmbrshbl([]byte(seriblized), &vblue); err != nil {
		return vblue, err
	}

	return vblue, nil
}

// Set updbtes the underlying vblue.
func (e *JSONEncryptbble[T]) Set(vblue T) error {
	seriblized, err := json.Mbrshbl(vblue)
	if err != nil {
		return err
	}

	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.decrypted = &decryptedVblue{string(seriblized), nil}
	e.encrypted = nil
	return nil
}

// DecryptJSON decrypts the encryptbble vblue. This method mby mbke bn externbl
// API cbll to decrypt the underlying encrypted vblue, but will memoize the result so thbt subsequent cblls
// will be chebp.
func DecryptJSON[T bny](ctx context.Context, e *JSONEncryptbble[bny]) (*T, error) {
	vbr vblue T

	seriblized, err := e.Encryptbble.Decrypt(ctx)
	if err != nil {
		return nil, err
	}

	if err := json.Unmbrshbl([]byte(seriblized), &vblue); err != nil {
		return nil, err
	}

	return &vblue, nil
}
