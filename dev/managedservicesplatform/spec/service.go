package spec

import (
	"fmt"
	"strings"

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
	// Description briefly summarizing what the service does. Required.
	//
	// ❗ We do NOT include this description in generated docs today - while it
	// might be helpful to include service descriptions, some services have
	// sensitive details or descriptions that are difficult to put into words
	// in a public-facing document. For now, this is used for reference in the
	// private service spec and for internal integrations like Opsgenie.
	Description string `yaml:"description"`
	// NotionPageID is the unique Notion ID representing the page that should
	// be populated with generated MSP operational guidance for this service
	// and its environments. This page is owned by the MSP doc generator - any
	// manual changes will be periodically overwritten.
	//
	// If not provided, no operational guidance will be generated for this
	// service.
	NotionPageID *string `yaml:"notionPageID,omitempty"`

	// Kind is the type of the service, either 'service' or 'job'. Defaults to
	// 'service'.
	Kind *ServiceKind `yaml:"kind,omitempty"`
	// Protocol is a protocol other than HTTP that the service communicates
	// with. If empty, the service uses HTTP. To use gRPC, configure 'h2c':
	// https://cloud.google.com/run/docs/configuring/http2
	//
	// To use gRPC and Cloudflare proxy (implicitly enabled unless
	// 'domain.cloudflare.proxied: false' is configured), the Cloudflare zone
	// must have gRPC enabled: https://developers.cloudflare.com/network/grpc-connections/
	// This is already enabled for the Cloudflare zones 'sourcegraph.com' and
	// 'sgdev.org'.
	Protocol *ServiceProtocol `yaml:"protocol,omitempty"`

	// IAM is an optional IAM configuration for the service account on the
	// service's GCP project.
	IAM *ServiceIAMSpec `yaml:"iam,omitempty"`
}

// GetName returns Name if configured, otherwise the ID.
func (s ServiceSpec) GetName() string {
	return pointers.Deref(s.Name, s.ID)
}

// GetKind returns Kind if configured, otherwise the default (ServiceKindService).
func (s ServiceSpec) GetKind() ServiceKind {
	return pointers.Deref(s.Kind, ServiceKindService)
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
	if len(s.Description) == 0 {
		errs = append(errs, errors.New("description is required"))
	}
	if s.NotionPageID != nil {
		page := *s.NotionPageID
		if len(page) == 0 {
			errs = append(errs, errors.New("notionPageID cannot be empty"))
		}
		// Should not be 'www.notion.so' or 'sourcegraph.notion.site'
		if strings.Contains(page, ".notion.") {
			errs = append(errs, errors.New("notionPageID must be a page ID, not a URL"))
		}
		// Should not have URL query parameters appended by Notion
		if strings.Contains(page, "?") {
			errs = append(errs, errors.New("notionPageID must be a page ID, found what looks like a URL query '?'"))
		}
	}

	if s.IAM != nil {
		errs = append(errs, s.IAM.Validate()...)
	}

	return errs
}

// GetHandbookPageURL returns the public URL of the Notion page that is populated
// with operational guidance for this service for embedding in docs or help
// text.
//
// If no NotionPageID is configured, this returns a warning message instead.
func (s ServiceSpec) GetHandbookPageURL() string {
	if s.NotionPageID == nil {
		return fmt.Sprintf("<%s service spec does not have a notionPageID configured for generated docs>", s.ID)
	}
	return fmt.Sprintf("https://sourcegraph.notion.site/%s", *s.NotionPageID)
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
