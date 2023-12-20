package spec

import (
	"fmt"
	"os"
	"path/filepath"

	// We intentionally use sigs.k8s.io/yaml because it has some convenience features,
	// and nicer formatting. We use this in Sourcegraph Cloud as well.
	"sigs.k8s.io/yaml"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
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
	Monitoring   *MonitoringSpec   `json:"monitoring,omitempty"`
}

// Open a specification file, validate it, unmarshal the data as a MSP spec,
// and load any extraneous configuration.
func Open(specPath string) (*Spec, error) {
	specData, err := os.ReadFile(specPath)
	if err != nil {
		return nil, errors.Wrap(err, "ReadFile")
	}
	spec, err := parse(specData)
	if err != nil {
		return nil, errors.Wrap(err, "spec.parse")
	}

	// Load extraneous resources
	configDir := filepath.Dir(specPath)
	for _, e := range spec.Environments {
		if e.Resources != nil && e.Resources.BigQueryDataset != nil {
			if err := e.Resources.BigQueryDataset.LoadSchemas(configDir); err != nil {
				return spec, errors.Wrap(err, "BigQueryTable.LoadSchema")
			}
		}
	}

	return spec, nil
}

// parse validates and unmarshals data as a MSP spec.
func parse(data []byte) (*Spec, error) {
	var s Spec
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, err
	}

	// Assign zero value for top-level monitoring spec for covenience
	s.Monitoring = &MonitoringSpec{}

	if validationErrs := s.Validate(); len(validationErrs) > 0 {
		return nil, errors.Append(nil, validationErrs...)
	}
	return &s, nil
}

func (s Spec) Validate() []error {
	var errs []error

	if s.Service.Kind.Is(ServiceKindJob) {
		for _, e := range s.Environments {
			if e.EnvironmentServiceSpec != nil {
				errs = append(errs, errors.New("service specifications are not supported for 'kind: job'"))
			}
			if e.Instances.Scaling != nil {
				errs = append(errs, errors.New("'environments.instances.scaling' not supported for 'kind: job'"))
			}
		}
	}

	for _, env := range s.ListEnvironmentIDs() {
		projectName := fmt.Sprintf("%s - %s",
			pointers.Deref(s.Service.Name, s.Service.ID), env)
		if len(projectName) > 30 {
			errs = append(errs, errors.Newf(
				"full environment name %q exceeds 30 characters limit - try a shorter service name or environment ID",
				projectName,
			))
		}
	}

	errs = append(errs, s.Service.Validate()...)
	errs = append(errs, s.Build.Validate()...)
	for _, env := range s.Environments {
		errs = append(errs, env.Validate()...)
	}
	errs = append(errs, s.Monitoring.Validate()...)
	return errs
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

func (s Spec) ListEnvironmentIDs() []string {
	var ids []string
	for _, e := range s.Environments {
		ids = append(ids, e.ID)
	}
	return ids
}

// MarshalYAML marshals the spec to YAML using our YAML library of choice.
func (s Spec) MarshalYAML() ([]byte, error) {
	return yaml.Marshal(s)
}
