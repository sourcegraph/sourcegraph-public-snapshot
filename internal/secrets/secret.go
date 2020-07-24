package secrets

import (
	"fmt"
)

func (e *Encrypter) Raw(crypt string) (string, error) {
	plaintext, err := e.Decrypt(crypt)
	if err != nil {
		return "", err
	}
	return plaintext, nil
}

// Return a masked version of the decrypted secret, for when a token-like string needs to be displayed in the UI
func (e *Encrypter) Mask(crypt string) (string, error) {
	plaintext, err := e.Decrypt(crypt)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s***********", plaintext[0:0]), nil
}
