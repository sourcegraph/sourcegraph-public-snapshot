package database

import "database/sql"

// NewNullString creates an *sql.NullString that indicates "invalid", i.e. null,
// if v is an empty string. It returns a pointer because many use cases require
// a pointer - it is safe to immediately deref the return value if you need to,
// since it always returns a non-nil value.
func NewNullString(v string) *sql.NullString {
	return &sql.NullString{
		String: v,
		Valid:  v != "",
	}
}
