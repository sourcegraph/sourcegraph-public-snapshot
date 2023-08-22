package spec

import (
	// We intentionally use sigs.k8s.io/yaml because it has some convenience features,
	// and nicer formatting. We use this in Sourcegraph Cloud as well.
	"sigs.k8s.io/yaml"
)

// Spec is a Managed Services Platform (MSP) service.
//
// All MSP services must:
//
//   - Serve its API on ":$PORT", if $PORT is provided
//   - Export a /-/healthz endpoint that authenticates requests using
//     "Authorization: Bearer $DIAGNOSTICS_SECRET", if $DIAGNOSTICS_SECRET is provided.
//
// Package dev/managedservicesplatform handles generating Terraform manifests
// from a given spec.
type Spec struct {
	Service      ServiceSpec       `json:"service"`
	Build        BuildSpec         `json:"build"`
	Environments []EnvironmentSpec `json:"environments"`
}

// Parse validates and unmarshals data as a MSP spec.
func Parse(data []byte) (*Spec, error) {
	// TODO: Add validaton
	var s Spec
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// GetEnvironment retrieves the environment with the given ID, returning nil if
// it doesn't exist.
func (s Spec) GetEnvironment(id string) *EnvironmentSpec {
	for _, e := range s.Environments {
		if e.ID == id {
			return &e
		}
	}
	return nil
}

func (s Spec) MarshalYAML() ([]byte, error) {
	return yaml.Marshal(s)
}
