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
	switch value := value.(type) {
	case string:
		*v = String(value)
	case *string:
		if value != nil {
			*v = String(*value)
		} else {
			*v = ""
		}
	case []byte:
		*v = String(value)
	case *[]byte:
		if value != nil {
			*v = String(*value)
		} else {
			*v = ""
		}
	case json.RawMessage:
		*v = String(value)
	case *json.RawMessage:
		if value != nil {
			*v = String(*value)
		} else {
			*v = ""
		}
	default:
		return fmt.Errorf("invalid type %T for nnz.String", value)
	}
	return nil
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
