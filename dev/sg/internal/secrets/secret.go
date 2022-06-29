package secrets

import (
	"fmt"
	"time"
)

type ExternalProvider string

const (
	// ExternalProviderGCloud fetches a secret from gcloud, where:
	//
	// - Project: gcloud project
	// - Name: secret name
	// - Field: unused
	ExternalProviderGCloud = "gcloud"
	// ExternalProvider1Pass fetches a secret from 1password, where:
	//
	// - Project: 1password vault
	// - Name: <itemName> | <itemID> | <shareLink>
	// - Field: field in item
	//
	// All 1password items are fetched from https://team-sourcegraph.1password.com/
	ExternalProvider1Pass = "1pass"
)

type ExternalSecret struct {
	Provider ExternalProvider `yaml:"provider"`

	// For details on how each field is used, see the relevant ExternalProvider docstring.
	Project string `yaml:"project"`
	Name    string `yaml:"name"`
	Field   string `yaml:"field,omitempty"`
}

func (s *ExternalSecret) id() string {
	id := fmt.Sprintf("%s/%s/%s", s.Provider, s.Project, s.Name)
	if s.Field != "" {
		id += fmt.Sprintf("/%s", s.Field)
	}
	return id
}

// externalSecretValue is the stored representation of an external secret's value
type externalSecretValue struct {
	Fetched time.Time
	Value   string
}
