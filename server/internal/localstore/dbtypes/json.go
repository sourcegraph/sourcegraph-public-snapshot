package dbtypes

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// JSONMapStringString is a map[string]string that serializes to JSON
// bytes when written to the DB, and deserializes accordingly.
type JSONMapStringString map[string]string

// Value implements database/sql/driver.Valuer.
func (j JSONMapStringString) Value() (driver.Value, error) {
	data, err := json.Marshal(j)
	return []byte(data), err
}

// Scan implements database/sql.Scanner.
func (j *JSONMapStringString) Scan(v interface{}) error {
	var data []byte
	switch v.(type) {
	case string:
		data = []byte(v.(string))
	case []byte:
		data = v.([]byte)
	case nil:
		data = []byte("null")
	default:
		return fmt.Errorf("invalid type %T for JSONMapStringString", v)
	}
	return json.Unmarshal(data, &j)
}
