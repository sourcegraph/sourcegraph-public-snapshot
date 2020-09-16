package secret

import (
	"database/sql/driver"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// StringValue implements driver.Valuer and sql.Scanner for string type secret.
type StringValue string

func (v *StringValue) Value() (driver.Value, error) {
	if !ConfiguredToEncrypt() {
		return []byte(*v), nil
	}
	if v == nil { // prefer all columns to be NON-NULL, but check in case
		return nil, nil
	}

	ciphertext, err := defaultEncryptor.EncryptBytes([]byte(*v))
	if err != nil {
		return nil, err
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (v *StringValue) Scan(src interface{}) error {

	var source []byte

	switch src := src.(type) {
	case string: // expect a base64 encoded string
		if !ConfiguredToEncrypt() {
			*v = StringValue(src)
			return nil
		}

		s, err := base64.StdEncoding.DecodeString(src)
		if err != nil {
			if _, ok := err.(base64.CorruptInputError); ok {
				*v = StringValue(src)
				return nil
			}
			return fmt.Errorf("unexpected error: %v", err)
		}
		source = s
	default:
		return errors.Errorf("incompatible type %T for StringValue", src)
	}

	plaintext, err := defaultEncryptor.DecryptBytes(source)
	if err != nil {
		return err
	}
	*v = StringValue(plaintext)
	return nil
}

func (v *StringValue) String() string {
	return string(*v)
}

// JSONValue implements driver.Valuer and sql.Scanner for JSON type secret.
type JSONValue json.RawMessage

func (v *JSONValue) Value() (driver.Value, error) {
	if v == nil { // prefer all columns to be NON-NULL, but check in case
		return `""`, nil
	}

	if !ConfiguredToEncrypt() {
		return `"` + string([]byte(*v)) + `"`, nil
	}

	ciphertext, err := defaultEncryptor.EncryptBytes([]byte(*v))
	if err != nil {
		return nil, err
	}
	return `"` + base64.StdEncoding.EncodeToString(ciphertext) + `"`, nil
}

func (v *JSONValue) Scan(src interface{}) error {
	var source []byte
	switch src := src.(type) {
	case string:

		// In case it's an empty JSON
		src = strings.Trim(src, `"{}`)
		if src == "" {
			*v = JSONValue(`""`)
			return nil
		}

		if !ConfiguredToEncrypt() {
			*v = JSONValue(src)
			return nil
		}

		s, err := base64.StdEncoding.DecodeString(src)
		if err != nil {
			return err
		}
		source = s

	default:
		return errors.Errorf("incompatible type %T for JSONValue", src)
	}

	plaintext, err := defaultEncryptor.DecryptBytes(source)
	if err != nil {
		return err
	}
	*v = JSONValue(plaintext)
	return nil
}

func (v *JSONValue) RawMessage() json.RawMessage {
	return json.RawMessage(*v)
}

func (v *JSONValue) String() string {
	return string(*v)
}
