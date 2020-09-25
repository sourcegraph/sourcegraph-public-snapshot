package secrets

import (
	"database/sql"
	"database/sql/driver"

	"github.com/pkg/errors"
)

var (
	_ driver.Valuer = StringValue{}
	_ sql.Scanner   = (*StringValue)(nil)
)

// StringValue implements driver.Valuer and sql.Scanner for string type secret.
type StringValue struct {
	S *string
}

func (v StringValue) Value() (driver.Value, error) {
	if v.S == nil {
		return nil, errors.New("unable to encrypt a nil string pointer")
	}
	return Encrypt(*v.S)
}

func (v *StringValue) Scan(src interface{}) (err error) {
	if v.S == nil {
		return errors.New("unable to decrypt to a nil string pointer")
	}

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

	*v.S = plaintext
	return nil
}

var (
	_ driver.Valuer = NullStringValue{}
	_ sql.Scanner   = (*NullStringValue)(nil)
)

// NullStringValue implements driver.Valuer and sql.Scanner for NULLABLE string type secret.
type NullStringValue struct {
	S     *string
	Valid bool // Valid is true if String is not NULL
}

func (nv *NullStringValue) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		es := StringValue{S: nv.S}
		err := es.Scan(v)
		if err != nil {
			return err
		}
		nv.Valid = true
	}
	return nil
}

func (nv NullStringValue) Value() (driver.Value, error) {
	if nv.S == nil {
		return nil, nil
	}

	return StringValue{S: nv.S}.Value()
}
