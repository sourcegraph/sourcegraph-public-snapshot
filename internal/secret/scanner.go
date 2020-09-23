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

//// NullString represents a string that may be null. NullString implements the
//// sql.Scanner interface so it can be used as a scan destination, similar to
//// sql.NullString. When the scanned value is null, String is set to the zero value.
//type NullString struct{ S *string }
//
//// Scan implements the Scanner interface.
//func (nt *NullString) Scan(value interface{}) error {
//	switch v := value.(type) {
//	case []byte:
//		*nt.S = string(v)
//	case string:
//		*nt.S = v
//	}
//	return nil
//}
//
//// Value implements the driver Valuer interface.
//func (nt NullString) Value() (driver.Value, error) {
//	if nt.S == nil {
//		return nil, nil
//	}
//	return *nt.S, nil
//}
// ---------

type NullStringValue struct{ S *StringValue }

func (s *NullStringValue) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		*s.S = StringValue(v)
	}
	return nil
}
