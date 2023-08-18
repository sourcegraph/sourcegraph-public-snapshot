package codehosts

import (
	"database/sql/driver"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type NullInt struct{ N *int }

// NewNullInt returns a NullInt treating zero value as null.
func NewNullInt(i int) NullInt {
	if i == 0 {
		return NullInt{}
	}
	return NullInt{N: &i}
}

// Scan implements the Scanner interface.
func (n *NullInt) Scan(value any) error {
	switch value := value.(type) {
	case int64:
		*n.N = int(value)
	case int32:
		*n.N = int(value)
	case nil:
		return nil
	default:
		return errors.Errorf("value is not int: %T", value)
	}
	return nil
}

// Value implements the driver Valuer interface.
func (n NullInt) Value() (driver.Value, error) {
	if n.N == nil {
		return nil, nil
	}
	return *n.N, nil
}

type NullInt32 struct{ N **int32 }

// Scan implements the Scanner interface.
func (n *NullInt32) Scan(value any) error {
	switch value := value.(type) {
	case int64:
		*n.N = pointers.Ptr(int32(value))
	case int32:
		*n.N = pointers.Ptr(int32(value))
	case nil:
		return nil
	default:
		return errors.Errorf("value is not int: %T", value)
	}
	return nil
}

// Value implements the driver Valuer interface.
func (n NullInt32) Value() (driver.Value, error) {
	if n.N == nil {
		return nil, nil
	}
	return *n.N, nil
}
