package secrets

import (
	"database/sql/driver"
	"encoding/base64"
	"errors"
	"fmt"
)

type EncryptedStringValue string

func (v *EncryptedStringValue) Value() (driver.Value, error) {
	ciphertext, err := defaultEncryptor.EncryptBytes([]byte(*v))
	if err != nil {
		return nil, err
	}
	encoded := base64.StdEncoding.EncodeToString(ciphertext)
	return encoded, nil
}

func (v *EncryptedStringValue) Scan(src interface{}) error {

	var source []byte

	switch src.(type) {
	case string: //expect a base64 encoded string
		s, err := base64.StdEncoding.DecodeString(src.(string))
		if err != nil {
			return fmt.Errorf("expected base64 encoded string: %v", err)
		}
		source = s
	case []byte:
		source = src.([]byte)
	default:
		return errors.New("incompatible type for EncryptedStringValue")
	}

	plaintext, err := defaultEncryptor.DecryptBytes(source)
	if err != nil {
		return err
	}
	*v = EncryptedStringValue(plaintext)
	return nil
}
