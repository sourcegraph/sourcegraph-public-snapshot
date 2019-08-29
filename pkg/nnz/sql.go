package nnz

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// String represents a Go string whose zero value maps to SQL null (when used with database/sql).
type String string

// Scan implements the database/sql.Scanner interface.
func (v *String) Scan(value interface{}) error {
	if value == nil {
		*v = ""
		return nil
	}
	str, err := scanString(value)
	if err != nil {
		return err
	}
	*v = String(str)
	return nil
}

func scanString(value interface{}) (string, error) {
	switch value := value.(type) {
	case string:
		return value, nil
	case *string:
		if value != nil {
			return *value, nil
		}
		return "", nil
	case []byte:
		return string(value), nil
	case *[]byte:
		if value != nil {
			return string(*value), nil
		}
		return "", nil
	case json.RawMessage:
		return string(value), nil
	case *json.RawMessage:
		if value != nil {
			return string(*value), nil
		}
		return "", nil
	default:
		return "", fmt.Errorf("invalid type %T for nnz.String", value)
	}
}

// Value implements the database/sql/driver.Valuer interface.
func (v String) Value() (driver.Value, error) {
	if v == "" {
		return nil, nil
	}
	return string(v), nil
}

// Int64 represents a Go int64 whose zero value maps to SQL null (when used with database/sql).
type Int64 int64

// Scan implements the database/sql.Scanner interface.
func (v *Int64) Scan(value interface{}) error {
	if value == nil {
		*v = 0
		return nil
	}
	i64, err := scanInt64(value)
	if err != nil {
		return err
	}
	*v = Int64(i64)
	return nil
}

func scanInt64(value interface{}) (int64, error) {
	switch value := value.(type) {
	case int:
		return int64(value), nil
	case int32:
		return int64(value), nil
	case int64:
		return value, nil
	case *int:
		if value != nil {
			return int64(*value), nil
		}
		return 0, nil
	case *int32:
		if value != nil {
			return int64(*value), nil
		}
		return 0, nil
	case *int64:
		if value != nil {
			return *value, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("invalid type %T for nnz.int64", value)
	}
}

// Value implements the database/sql/driver.Valuer interface.
func (v Int64) Value() (driver.Value, error) {
	if v == 0 {
		return nil, nil
	}
	return int64(v), nil
}

// Int32 returns a driver Value that maps int32(0) to SQL null.
func Int32(v int32) driver.Value {
	if v == 0 {
		return nil
	}
	return int64(v)
}

// ToInt32 returns a value that implements database/sql.Scanner so that SQL null maps to int32(0).
func ToInt32(v *int32) sql.Scanner {
	return (*int32Scanner)(v)
}

type int32Scanner int32

// Scan implements the database/sql.Scanner interface.
func (v *int32Scanner) Scan(value interface{}) error {
	if value == nil {
		*v = 0
		return nil
	}
	i64, err := scanInt64(value)
	if err != nil {
		return err
	}
	*v = int32Scanner(i64)
	return nil
}

// JSON returns a driver Value that maps json.RawMessage(nil) to SQL null.
func JSON(v json.RawMessage) driver.Value {
	if v == nil {
		return nil
	}
	return []byte(v)
}

// ToJSON returns a value that implements database/sql.Scanner so that SQL null maps to
// json.RawMessage(nil).
func ToJSON(v *json.RawMessage) sql.Scanner {
	return (*jsonScanner)(v)
}

type jsonScanner json.RawMessage

// Scan implements the database/sql.Scanner interface.
func (v *jsonScanner) Scan(value interface{}) error {
	if value == nil {
		*v = nil
		return nil
	}
	str, err := scanString(value)
	if err != nil {
		return err
	}
	if str == "" {
		*v = nil
		return nil
	}
	*v = jsonScanner(str)
	return nil
}
