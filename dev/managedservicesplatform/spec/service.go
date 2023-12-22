package spec

import (
	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type ServiceSpec struct {
	// ID is an all-lowercase, hyphen-delimited identifier for the service,
	// e.g. "cody-gateway". It MUST be at most 20 characters long.
	ID string `json:"id"`
	// Name is an optional human-readable display name for the service,
	// e.g. "Cody Gateway".
	Name *string `json:"name"`
	// Owners denotes the teams or individuals primarily responsible for the
	// service.
	Owners []string `json:"owners"`

	// Kind is the type of the service, either 'service' or 'job'. Defaults to
	// 'service'.
	Kind *ServiceKind `json:"kind,omitempty"`
	// Protocol is a protocol other than HTTP that the service communicates
	// with. If empty, the service uses HTTP. To use gRPC, configure 'h2c':
	// https://cloud.google.com/run/docs/configuring/http2
	Protocol *ServiceProtocol `json:"protocol,omitempty"`

	// ProjectIDSuffixLength can be configured to truncate the length of the
	// service's generated project IDs.
	ProjectIDSuffixLength *int `json:"projectIDSuffixLength,omitempty"`

	// IAM is an optional IAM configuration for the service account on the
	// service's GCP project.
	IAM *ServiceIAMSpec `json:"iam,omitempty"`
}

func (s ServiceSpec) Validate() []error {
	var errs []error

	if len(s.ID) > 20 {
		errs = append(errs, errors.New("id must be at most 20 characters"))
	}

	if s.ProjectIDSuffixLength != nil && *s.ProjectIDSuffixLength < 4 {
		errs = append(errs, errors.New("projectIDSuffixLength must be >= 4"))
	}

	if s.IAM != nil {
		errs = append(errs, s.IAM.Validate()...)
	}

	// TODO: Add validation
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
	Services []string `json:"services,omitempty"`

	// Roles is a list of IAM roles to grant to the service account.
	Roles []string `json:"roles,omitempty"`
	// Permissions is a list of IAM permissions to grant to the service account.
	//
	// MSP will create a custom role with these permissions and grant it to the
	// service account.
	Permissions []string `json:"permissions,omitempty"`
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
