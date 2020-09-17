package secret

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
)

var (
	_ driver.Valuer = StringValue("")
	_ sql.Scanner   = (*StringValue)(nil)
)

// StringValue implements driver.Valuer and sql.Scanner for string type secret.
type StringValue string

func (v StringValue) Value() (driver.Value, error) {
	return Encrypt(string(v))
}

func (v *StringValue) Scan(src interface{}) (err error) {
	var plaintext string
	switch src := src.(type) {
	case string: // expect a base64 encoded string
		plaintext, err = Decrypt(src)
		if err != nil {
			return err
		}

	default:
		return errors.Errorf("incompatible type %T for StringValue", src)
	}

	*v = StringValue(plaintext)
	return nil
}

func (v *StringValue) String() string {
	return string(*v)
}

var (
	_ driver.Valuer = JSONValue("")
	_ sql.Scanner   = (*JSONValue)(nil)
)

// JSONValue implements driver.Valuer and sql.Scanner for JSON type secret.
type JSONValue json.RawMessage

func (v JSONValue) Value() (driver.Value, error) {
	if len(v) == 0 {
		return `""`, nil
	} else if !ConfiguredToEncrypt() {
		return string(v), nil
	}

	ciphertext, err := Encrypt(string(v))
	if err != nil {
		return nil, err
	}
	return `"` + ciphertext + `"`, nil
}

func (v *JSONValue) Scan(src interface{}) (err error) {
	var plaintext string
	switch src := src.(type) {
	case []byte:
		ciphertext := string(src)

		// In case it's not configured to encrypt or an empty JSON
		if !ConfiguredToEncrypt() ||
			ciphertext == `""` || ciphertext == `{}` {
			plaintext = ciphertext
			break
		}

		ciphertext = strings.Trim(string(src), `"`)
		plaintext, err = Decrypt(ciphertext)
		if err != nil {
			return err
		}

	default:
		return errors.Errorf("incompatible type %T for JSONValue", src)
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
