package dbstore

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type LocalStep struct {
	Commands []string `json:"commands"`
}

func (s *LocalStep) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("value is not []byte: %T", value)
	}

	return json.Unmarshal(b, &s)
}

func (s LocalStep) Value() (driver.Value, error) {
	return json.Marshal(s)
}
