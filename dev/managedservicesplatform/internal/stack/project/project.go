package project

import (
	"strings"

	"github.com/aws/jsii-runtime-go"
	"github.com/grafana/regexp"

	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/project"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/projectservice"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/random"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack/options/googleprovider"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack/options/randomprovider"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

var gcpServices = []string{
	"run.googleapis.com",
	"containerregistry.googleapis.com",
	"cloudbuild.googleapis.com",
	"logging.googleapis.com",
	"monitoring.googleapis.com",
	"iam.googleapis.com",
	"secretmanager.googleapis.com",
	"redis.googleapis.com",
	"compute.googleapis.com",
	"networkmanagement.googleapis.com",
	"vpcaccess.googleapis.com",
	"servicenetworking.googleapis.com",
	"storage-api.googleapis.com",
	"storage-component.googleapis.com",
	"bigquery.googleapis.com",
}

const (
	BillingAccountID = "017005-C370B2-0E3030"
	// DefaultProjectFolderID points to the 'Managed Services' folder
	DefaultProjectFolderID = "folders/26336759932"
)

type Output struct {
	// Project is created with a generated project ID.
	Project project.Project
}

type Variables struct {
	// ProjectIDPrefix is the prefix for a project ID. A suffix of the format
	// '-${randomizedSuffix}' will be added, as project IDs must be unique.
	ProjectIDPrefix string

	// Name is a display name for the project. It does not need to be unique.
	Name string

	// Labels to apply to the project.
	Labels map[string]string

	// ProjectFolderID is a 'folders/'-prefixed folder ID to create the project
	// in. A default project is set.
	ProjectFolderID *string

	// EnableAuditLogs ships GCP audit logs to security cluster.
	// TODO: Not yet implemented
	EnableAuditLogs bool
}

const StackName = "project"

// NewStack creates a stack that provisions a GCP project.
func NewStack(stacks *stack.Set, vars Variables) (*Output, error) {
	stack := stacks.New(StackName,
		randomprovider.With(),
		// ID is not known ahead of time, we can omit it
		googleprovider.With(""))

	// Name all stack resources after the desired project ID
	id := resourceid.New(vars.ProjectIDPrefix)

	projectID := random.New(stack, id, random.Config{
		ByteLength: 6,
		Prefix:     vars.ProjectIDPrefix,
	})

	output := &Output{
		Project: project.NewProject(stack,
			id.ResourceID("project"),
			&project.ProjectConfig{
				Name:              pointers.Ptr(vars.Name),
				ProjectId:         &projectID.HexValue,
				AutoCreateNetwork: false,
				BillingAccount:    pointers.Ptr(BillingAccountID),
				FolderId:          pointers.Ptr(pointers.Deref(vars.ProjectFolderID, DefaultProjectFolderID)),
				Labels: func(input map[string]string) *map[string]*string {
					labels := make(map[string]*string)
					for k, v := range input {
						labels[sanitizeName(k)] = pointers.Ptr(v)
					}
					return &labels
				}(vars.Labels),
			}),
	}

	for _, service := range gcpServices {
		projectservice.NewProjectService(stack,
			id.ResourceID("project-service-%s", strings.ReplaceAll(service, ".", "-")),
			&projectservice.ProjectServiceConfig{
				Project:                  output.Project.ProjectId(),
				Service:                  pointers.Ptr(service),
				DisableDependentServices: jsii.Bool(false),
				// prevent accidental deletion of services
				DisableOnDestroy: jsii.Bool(false),
			})
	}

	return output, nil
}

var regexpMatchNonLowerAlphaNumericUnderscoreDash = regexp.MustCompile(`[^a-z0-9_-]`)

// sanitizeName ensures the name contains only lowercase letters, numerals, underscores, and dashes.
// non-compliant characters are replaced with underscores.
func sanitizeName(name string) string {
	s := strings.ToLower(name)
	s = regexpMatchNonLowerAlphaNumericUnderscoreDash.ReplaceAllString(s, "_")
	return s
}
