package spec

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type EnvironmentPrivateNetworkingSpec struct {
	// PrivateGoogleAccess configures routing to Private Google Access (i.e.
	// via internal networking) for selected Google services.
	//
	// If no value is provided, the default is currently to not provision any
	// routing.
	PrivateGoogleAccess *EnvironmentPrivateGoogleAccessSpec `yaml:"privateGoogleAccess,omitempty"`
	// PrivateAccessPerimeter, if configured to a non-nil value, will provision
	// a VPC Service Controls perimeter around this environment's GCP project's
	// Cloud Run APIs (including the Cloud Run service's internal direct-access
	// URL).
	//
	// Note that this prevents real users from accessing the Cloud Run service
	// via Cloud Console, even with the prerequisite Entitle permissions. Users
	// must be added to AllowlistedIdentities - GCP permissions will still be
	// enforced.
	//
	// Only supported for services of 'kind: service'.
	PrivateAccessPerimeter *EnvironmentPrivateAccessPerimeterSpec `yaml:"privateAccessPerimeter,omitempty"`
}

func (s *EnvironmentPrivateNetworkingSpec) Validate() []error {
	if s == nil {
		return nil
	}
	var errs []error
	errs = append(errs, s.PrivateAccessPerimeter.Validate()...)
	return errs
}

type EnvironmentPrivateGoogleAccessSpec struct {
	// CloudRunApps toggles Private Google Access directly TO the internal URLs
	// of Cloud Run Apps (i.e. '*.run.app' URLs).
	CloudRunApps *bool `yaml:"cloudRunApps,omitempty"`
}

type EnvironmentPrivateAccessPerimeterSpec struct {
	// AllowlistedProjects is a list of GCP projects whose internal traffic
	// (e.g. egress from a VPC network within the listed projects) is allowed
	// direct ingress to the Cloud Run service in the perimeter, typically via
	// the service's internal Cloud Run URL (i.e. '*.run.app' URLs).
	//
	// If you want to allow private ingress to this service from another MSP
	// service's project, include the desired environment's project ID here.
	AllowlistedProjects []string `yaml:"allowlistedProjects"`
	// AllowlistedIdentities is a list of Google identities that are allowed
	// ingress to the Cloud Run service in the perimeter. Values are expected to
	// use the https://cloud.google.com/iam/docs/principal-identifiers.md#v1
	// format.
	//
	// Currently, only 'serviceAccount:' and 'user:' identities are supported.
	AllowlistedIdentities []string `yaml:"allowlistedIdentities,omitempty"`
}

var allowedPrivateAccessServerAllowlistedIdentityTypes = []string{"serviceAccount:", "user:"}

func (e *EnvironmentPrivateAccessPerimeterSpec) Validate() []error {
	if e == nil {
		return nil
	}

	var errs []error

	for idx, identity := range e.AllowlistedIdentities {
		var hasAllowedType bool
		for _, allowedType := range allowedPrivateAccessServerAllowlistedIdentityTypes {
			if strings.HasPrefix(identity, allowedType) {
				hasAllowedType = true
				break
			}
		}
		if !hasAllowedType {
			errs = append(errs, errors.Newf("allowlistedIdentities[%d]: identity %s must be of one of the allowed types: [%s]",
				idx, identity, strings.Join(allowedPrivateAccessServerAllowlistedIdentityTypes, ", ")))
		}
	}

	return errs
}
