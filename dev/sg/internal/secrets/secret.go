package secrets

import (
	"fmt"
	"time"
)

type ExternalSecret struct {
	Provider string `yaml:"provider"`
	Project  string `yaml:"project"`
	Name     string `yaml:"name"`
}

func (s *ExternalSecret) id() string {
	return fmt.Sprintf("%s/%s/%s", s.Provider, s.Project, s.Name)
}

// externalSecretValue is the stored representation of an external secret's value
type externalSecretValue struct {
	Fetched time.Time
	Value   string
}
