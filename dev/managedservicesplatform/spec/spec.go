package spec

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

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
	Service      ServiceSpec       `yaml:"service"`
	Build        BuildSpec         `yaml:"build"`
	Environments []EnvironmentSpec `yaml:"environments"`
	Monitoring   *MonitoringSpec   `yaml:"monitoring,omitempty"`
	// Rollout can be configured to indicate how releases should roll out
	// through a set of environments.
	Rollout *RolloutSpec `yaml:"rollout,omitempty"`

	// README is the contents of the README.md file adjacent to the service
	// specification. May be a zero-length byte slice if a README file is not
	// present, or nil if this spec was not opened with 'spec.Open(...)'.
	README []byte `yaml:"-"`
}

var ErrServiceDoesNotExist = errors.New("service does not exist")

// Open a specification file, validate it, unmarshal the data as a MSP spec,
// and load any extraneous configuration. Callsites that return an error to the
// user should wrap the error with the name of the service to avoid any confusion,
// e.g.
//
//	s, err := spec.Open(serviceSpecPath)
//	if err != nil {
//		return errors.Wrapf(err, "load service %q", serviceID)
//	}
//
// This is helpful because this function only accepts the spec path.
func Open(specPath string) (*Spec, error) {
	specData, err := os.ReadFile(specPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrServiceDoesNotExist
		}
		return nil, errors.Wrap(err, "read service specification")
	}
	spec, err := parse(specData)
	if err != nil {
		return spec, errors.Wrap(err, "spec.parse")
	}

	// Load extraneous resources
	configDir := filepath.Dir(specPath)
	if err := spec.loadREADME(configDir); err != nil {
		return spec, errors.Wrap(err, "spec.loadREADME")
	}
	for _, e := range spec.Environments {
		if e.Resources != nil && e.Resources.BigQueryDataset != nil {
			if err := e.Resources.BigQueryDataset.LoadSchemas(configDir); err != nil {
				return spec, errors.Wrapf(err, "spec.environments.[%s].Resources.BigQueryDataset.LoadSchema", e.ID)
			}
		}
	}

	return spec, nil
}

// AppendEnvironment attaches environmentSpec, expressed as a map *yaml.Node,
// to the spec's "environments" list. It returns the updated spec data. The
// update preserves all formatting and docstrings.
func AppendEnvironment(specData []byte, environmentSpec *yaml.Node) ([]byte, error) {
	if environmentSpec.Kind != yaml.ScalarNode && environmentSpec.Tag != "!!map" {
		return nil, errors.Newf("environment spec must be a YAML map node, got kind: %v, tag: %q",
			environmentSpec.Kind, environmentSpec.Tag)
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(specData, &doc); err != nil {
		return nil, errors.Wrap(err, "parse spec YAML")
	}

	var added bool
	root := doc.Content[0]
	for i, n := range root.Content {
		if n.Value == "environments" {
			envList := root.Content[i+1]
			envList.Content = append(envList.Content, environmentSpec)
			added = true
			break
		}
	}
	if !added {
		return nil, errors.New("spec 'environments' field not found")
	}

	// This is the only place we marshal a spec, other than the hand-written
	// templates in dev/sg/msp/example. We need to set up an encoder to align
	// with our preferences.
	var update bytes.Buffer
	enc := yaml.NewEncoder(&update)
	enc.SetIndent(2)
	if err := enc.Encode(&doc); err != nil {
		return nil, errors.Wrap(err, "render updated spec to YAML")
	}

	return update.Bytes(), nil
}

// parse validates and unmarshals data as a MSP spec.
func parse(data []byte) (*Spec, error) {
	var s Spec
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, err
	}

	// Assign zero value for top-level monitoring spec for covenience
	if s.Monitoring == nil {
		s.Monitoring = &MonitoringSpec{}
	}

	if validationErrs := s.Validate(); len(validationErrs) > 0 {
		return &s, errors.Append(nil, validationErrs...)
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
			if e.Deploy.Type == EnvironmentDeployTypeRollout {
				errs = append(errs, errors.New("'deploy { type: \"rollout\" }' not supported for 'kind: job'"))
			}
			if e.Instances.Scaling != nil {
				errs = append(errs, errors.New("'environments.instances.scaling' not supported for 'kind: job'"))
			}
		}
	}
	if s.Service.Kind.Is(ServiceKindService) {
		for _, e := range s.Environments {
			if e.EnvironmentJobSpec != nil {
				errs = append(errs, errors.New("job specifications are not supported for 'kind: service'"))
			}
		}
	}

	configuredDomains := map[string]struct{}{}
	for _, env := range s.Environments {
		projectDisplayName := fmt.Sprintf("%s - %s", s.Service.GetName(), env.ID)
		if len(projectDisplayName) > 30 {
			errs = append(errs, errors.Newf(
				"full environment name %q exceeds 30 characters limit - try a shorter service name or environment ID",
				projectDisplayName,
			))
		}

		if !strings.HasPrefix(env.ProjectID, fmt.Sprintf("%s-", s.Service.ID)) {
			errs = append(errs, errors.Newf("environment %q projectID %q must contain service ID: expecting format '$SERVICE_ID-$ENVIRONMENT_ID-$RANDOM_SUFFIX'",
				env.ID, env.ProjectID))
		}

		if env.Category.IsProduction() {
			if s.Service.NotionPageID == nil {
				errs = append(errs, errors.Newf("environment %q in category %q requires 'service.notionPageID' to be set",
					env.ID, env.Category))
			}
		}

		if domain := pointers.DerefZero(env.EnvironmentServiceSpec).Domain.GetDNSName(); domain != "" {
			_, exists := configuredDomains[domain]
			if exists {
				errs = append(errs, errors.Newf("domain %q is configured more than once across environments",
					domain))
			} else {
				configuredDomains[domain] = struct{}{}
			}
		}

		if s.Rollout.GetStageByEnvironment(env.ID) != nil {
			if env.Deploy.Type != EnvironmentDeployTypeRollout {
				errs = append(errs, errors.Newf("environment %q is referenced in a rollout stage - deploy type must be '%s'",
					env.ID, EnvironmentDeployTypeRollout))
			}
		} else if env.Deploy.Type == EnvironmentDeployTypeRollout {
			errs = append(errs, errors.Newf("environment %q has deploy type '%s', but is not referenced in rollout stages",
				EnvironmentDeployTypeRollout, env.ID))
		}
	}

	if s.Rollout != nil {
		if len(s.Rollout.Stages) == 0 {
			errs = append(errs, errors.New("rollout spec is defined but contains no stages"))
		}

		seenStages := make(map[string]struct{})
		for _, stage := range s.Rollout.Stages {
			if s.GetEnvironment(stage.EnvironmentID) == nil {
				errs = append(errs, errors.Newf("rollout stage references unknown environment %q",
					stage.EnvironmentID))
			}
			if _, seen := seenStages[stage.EnvironmentID]; seen {
				errs = append(errs, errors.Newf("rollout stage references environment %q more than once",
					stage.EnvironmentID))
			} else {
				seenStages[stage.EnvironmentID] = struct{}{}
			}
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

// loadREADME populates s.readme by convention, looking for the `README.md` file
// in dir. It expects a specific header format, and populates s.readme with the
// specific header removed.
//
// It is called by spec.Open(...).
func (s *Spec) loadREADME(dir string) error {
	// Open by convention
	readme, err := os.ReadFile(filepath.Join(dir, "README.md"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return errors.Wrapf(err, "open README")
	}

	// Must have header as prefix
	mustStartWith := "# " + s.Service.GetName()
	if !bytes.HasPrefix(readme, []byte(mustStartWith)) {
		return errors.Newf("README must start with an H1 header that matches the service name, i.e. %q", mustStartWith)
	}

	// Trim the prefix
	s.README = bytes.TrimSpace(bytes.TrimPrefix(readme, append([]byte(mustStartWith), '\n')))
	return nil
}
