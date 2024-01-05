package spec

import (
	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type ServiceSpec struct {
	// ID is an all-lowercase, hyphen-delimited identifier for the service,
	// e.g. "cody-gateway". It MUST be at most 20 characters long.
	ID string `yaml:"id"`
	// Name is an optional human-readable display name for the service,
	// e.g. "Cody Gateway".
	Name *string `yaml:"name"`
	// Owners denotes the teams or individuals primarily responsible for the
	// service. Each owner MUST be a valid Opsgenie team name - this is validated
	// in each environment's monitoring stack.
	Owners []string `yaml:"owners"`

	// Kind is the type of the service, either 'service' or 'job'. Defaults to
	// 'service'.
	Kind *ServiceKind `yaml:"kind,omitempty"`
	// Protocol is a protocol other than HTTP that the service communicates
	// with. If empty, the service uses HTTP. To use gRPC, configure 'h2c':
	// https://cloud.google.com/run/docs/configuring/http2
	Protocol *ServiceProtocol `yaml:"protocol,omitempty"`

	// IAM is an optional IAM configuration for the service account on the
	// service's GCP project.
	IAM *ServiceIAMSpec `yaml:"iam,omitempty"`
}

func (s ServiceSpec) Validate() []error {
	var errs []error

	if s.ID == "" {
		errs = append(errs, errors.New("id is required"))
	}
	if len(s.ID) > 20 {
		errs = append(errs, errors.New("id must be at most 20 characters"))
	}
	if !regexp.MustCompile(`^[a-z0-9-]+$`).MatchString(s.ID) {
		errs = append(errs, errors.New("id can only contain lowercase alphanumeric characters and hyphens"))
	}
	if len(s.Owners) == 0 {
		errs = append(errs, errors.New("owners requires at least one value"))
	}
	for i, o := range s.Owners {
		if o == "" {
			errs = append(errs, errors.Newf("owners[%d] is invalid", i))
		}
	}

	if s.IAM != nil {
		errs = append(errs, s.IAM.Validate()...)
	}

	return errs
}

type ServiceProtocol string

const ServiceProtocolH2C ServiceProtocol = "h2c"

type ServiceKind string

const (
	ServiceKindService ServiceKind = "service"
	ServiceKindJob     ServiceKind = "job"
)

func (s *ServiceKind) Is(kind ServiceKind) bool {
	configuredKind := pointers.Deref(s, ServiceKindService)
	return configuredKind == kind
}

type ServiceIAMSpec struct {
	// Services is a list of GCP services to enable in the service's project.
	Services []string `yaml:"services,omitempty"`

	// Roles is a list of IAM roles to grant to the service account.
	Roles []string `yaml:"roles,omitempty"`
	// Permissions is a list of IAM permissions to grant to the service account.
	//
	// MSP will create a custom role with these permissions and grant it to the
	// service account.
	Permissions []string `yaml:"permissions,omitempty"`
}

func (s ServiceIAMSpec) Validate() []error {
	var errs []error

	for _, role := range s.Roles {
		if !validIAMRole(role) {
			errs = append(errs, errors.Errorf("invalid IAM role %q, must be one of custom role or predefined role", role))
		}
	}

	return errs
}

func validIAMRole(role string) bool {
	return matchCustomRole.MatchString(role) || matchPredefinedRole.MatchString(role)
}

var (
	matchCustomRole     = regexp.MustCompile(`^(projects|organizations)/[a-z0-9_-]+/roles/[a-zA-Z_\.]+$`)
	matchPredefinedRole = regexp.MustCompile(`^roles/[a-zA-Z\.]+$`)
)
