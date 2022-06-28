package dbstore

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type DockerStep struct {
	Root     string   `json:"root"`
	Image    string   `json:"image"`
	Commands []string `json:"commands"`
}

func (s *DockerStep) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.Errorf("value is not []byte: %T", value)
	}

	return json.Unmarshal(b, &s)
}

func (s DockerStep) Value() (driver.Value, error) {
	return json.Marshal(s)
}
