package nnz

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// Value is a convenience interface implemented by this package's types.
type Value interface {
	// IsZero reports whether the value is the zero value for its type.
	IsZero() bool
}

// String represents a Go string whose zero value maps to SQL null (when used with database/sql).
type String string

// IsZero reports whether v == "".
func (v String) IsZero() bool { return v == "" }

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

// Int32 represents a Go int32 whose zero value maps to SQL null (when used with database/sql).
type Int32 int32

// IsZero reports whether v == 0.
func (v Int32) IsZero() bool { return v == 0 }

// Scan implements the database/sql.Scanner interface.
func (v *Int32) Scan(value interface{}) error {
	if value == nil {
		*v = 0
		return nil
	}
	switch value := value.(type) {
	case int:
		*v = Int32(value)
	case int32:
		*v = Int32(value)
	case int64:
		*v = Int32(value)
	case *int:
		if value != nil {
			*v = Int32(*value)
		} else {
			*v = 0
		}
	case *int32:
		if value != nil {
			*v = Int32(*value)
		} else {
			*v = 0
		}
	case *int64:
		if value != nil {
			*v = Int32(*value)
		} else {
			*v = 0
		}
	default:
		return fmt.Errorf("invalid type %T for nnz.Int32", value)
	}
	return nil
}

// Value implements the database/sql/driver.Valuer interface.
func (v Int32) Value() (driver.Value, error) {
	if v == 0 {
		return nil, nil
	}
	return int64(v), nil
}

// Int64 represents a Go int64 whose zero value maps to SQL null (when used with database/sql).
type Int64 int64

// IsZero reports whether v == 0.
func (v Int64) IsZero() bool { return v == 0 }

// Scan implements the database/sql.Scanner interface.
func (v *Int64) Scan(value interface{}) error {
	if value == nil {
		*v = 0
		return nil
	}
	switch value := value.(type) {
	case int:
		*v = Int64(value)
	case int32:
		*v = Int64(value)
	case int64:
		*v = Int64(value)
	case *int:
		if value != nil {
			*v = Int64(*value)
		} else {
			*v = 0
		}
	case *int32:
		if value != nil {
			*v = Int64(*value)
		} else {
			*v = 0
		}
	case *int64:
		if value != nil {
			*v = Int64(*value)
		} else {
			*v = 0
		}
	default:
		return fmt.Errorf("invalid type %T for nnz.Int64", value)
	}
	return nil
}

// Value implements the database/sql/driver.Valuer interface.
func (v Int64) Value() (driver.Value, error) {
	if v == 0 {
		return nil, nil
	}
	return int64(v), nil
}
