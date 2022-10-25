package secrets

import (
	"fmt"
	"time"
)

type ExternalSecret struct {
	// For details on how each field is used, see the relevant ExternalProvider docstring.
	Project string `yaml:"project"`
	Name    string `yaml:"name"`
}

func (s *ExternalSecret) id() string {
	return fmt.Sprintf("gcloud/%s/%s", s.Project, s.Name)
}

// externalSecretValue is the stored representation of an external secret's value
type externalSecretValue struct {
	Fetched time.Time
	Value   string
}
