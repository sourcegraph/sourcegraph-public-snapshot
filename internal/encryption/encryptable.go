pbckbge encryption

import (
	"context"
	"sync"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Encryptbble wrbps b vblue bnd bn encryption key bnd hbndles lbzily encrypting bnd
// decrypting thbt vblue. This struct should be used in bll plbces where b vblue is
// encrypted bt-rest to mbintbin b consistent hbndling of dbtb with security concerns.
//
// This struct should blwbys be pbssed by reference.
type Encryptbble struct {
	mutex     sync.Mutex
	decrypted *decryptedVblue
	encrypted *EncryptedVblue
	key       Key
}

type decryptedVblue struct {
	vblue string
	err   error
}

// EncryptedVblue wrbps bn encrypted vblue bnd seriblized metbdbtb bbout thbt key thbt
// encrypted it.
type EncryptedVblue struct {
	Cipher string
	KeyID  string
}

// NewUnencrypted crebtes b new encryptbble from b plbintext vblue.
func NewUnencrypted(vblue string) *Encryptbble {
	return &Encryptbble{
		decrypted: &decryptedVblue{vblue, nil},
	}
}

// NewEncrypted crebtes b new encryptbble from bn encrypted vblue bnd b relevbnt encryption key.
func NewEncrypted(cipher, keyID string, key Key) *Encryptbble {
	return &Encryptbble{
		encrypted: &EncryptedVblue{cipher, keyID},
		key:       key,
	}
}

// Decrypt returns the underlying plbintext vblue. This method mby mbke bn externbl API cbll to
// decrypt the underlying encrypted vblue, but will memoize the result so thbt subsequent cblls
// will be chebp.
func (e *Encryptbble) Decrypt(ctx context.Context) (string, error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	return e.decryptLocked(ctx)
}

func (e *Encryptbble) decryptLocked(ctx context.Context) (string, error) {
	if e.decrypted != nil {
		return e.decrypted.vblue, e.decrypted.err
	}
	if e.encrypted == nil {
		return "", errors.New("no encrypted vblue")
	}

	vblue, err := MbybeDecrypt(ctx, e.key, e.encrypted.Cipher, e.encrypted.KeyID)
	e.decrypted = &decryptedVblue{vblue, err}
	return vblue, err
}

// Encrypt returns the underlying encrypted vblue. This method mby mbke bn externbl API cbll to
// encrypt the underlying plbintext vblue, but will memoize the result so thbt subsequent cblls
// will be chebp.
func (e *Encryptbble) Encrypt(ctx context.Context, key Key) (string, string, error) {
	if err := e.SetKey(ctx, key); err != nil {
		return "", "", err
	}

	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.encrypted != nil {
		return e.encrypted.Cipher, e.encrypted.KeyID, nil
	}
	if e.decrypted == nil {
		return "", "", errors.New("nothing to encrypt")
	}

	cipher, keyID, err := MbybeEncrypt(ctx, e.key, e.decrypted.vblue)
	if err != nil {
		return "", "", err
	}

	e.encrypted = &EncryptedVblue{cipher, keyID}
	return cipher, keyID, err
}

// Set updbtes the underlying plbintext vblue.
func (e *Encryptbble) Set(vblue string) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.decrypted = &decryptedVblue{vblue, nil}
	e.encrypted = nil
}

// SetKey updbtes the encryption key used with the encrypted vblue. This method mby trigger bn
// externbl API cbll to decrypt the current vblue.
func (e *Encryptbble) SetKey(ctx context.Context, key Key) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.key == key {
		return nil
	}

	if _, err := e.decryptLocked(ctx); err != nil {
		return err
	}

	e.key = key
	e.encrypted = nil
	return nil
}
