package database

import "database/sql"

func NewNullString(v string) sql.NullString {
	return sql.NullString{
		String: v,
		Valid:  v != "",
	}
}
