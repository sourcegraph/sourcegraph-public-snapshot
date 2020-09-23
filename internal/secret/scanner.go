package secret

import (
	"database/sql"
	"database/sql/driver"

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
	_ driver.Valuer = NullStringValue{}
	_ sql.Scanner   = (*NullStringValue)(nil)
)

type NullStringValue struct {
	S     *StringValue
	Valid bool // Valid is true if StringValue is not NULL
}

func (nv *NullStringValue) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		*nv.S = StringValue(v)
		nv.Valid = true
	}
	return nil
}

func (nv NullStringValue) Value() (driver.Value, error) {
	if nv.S == nil {
		return nil, nil
	}
	return *nv.S, nil
}
