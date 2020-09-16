package secret

import (
	"database/sql/driver"
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
)

// StringValue implements driver.Valuer and sql.Scanner for string type secret.
type StringValue string

func (v *StringValue) Value() (driver.Value, error) {
	if v == nil { // prefer all columns to be NON-NULL, but check in case
		return nil, nil
	}

	return Encrypt(string(*v))
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

// JSONValue implements driver.Valuer and sql.Scanner for JSON type secret.
type JSONValue json.RawMessage

func (v *JSONValue) Value() (driver.Value, error) {
	if v == nil { // prefer all columns to be NON-NULL, but check in case
		return `""`, nil
	}

	return Encrypt(string(*v))
}

func (v *JSONValue) Scan(src interface{}) (err error) {
	var plaintext string
	switch src := src.(type) {
	case string:
		// In case it's an empty JSON
		src = strings.Trim(src, `"{}`)
		if src == "" {
			plaintext = `""`
			break
		}

		plaintext, err = Decrypt(src)
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
