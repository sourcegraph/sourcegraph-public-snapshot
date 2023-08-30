package project

import (
	"strings"

	"github.com/aws/jsii-runtime-go"
	"github.com/grafana/regexp"

	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/project"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/projectservice"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack/options/googleprovider"
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
	ProjectFolderID  = "26336759932"
)

type Output struct {
	Project project.Project
}

type Variables struct {
	ProjectID string
	Name      string
	Labels    map[string]string

	// EnableAuditLogs ships GCP audit logs to security cluster
	EnableAuditLogs bool
}

const StackName = "project"

// NewStack creates a stack that provisions a GCP project.
func NewStack(stacks *stack.Set, vars Variables) (*Output, error) {
	stack := stacks.New(StackName,
		googleprovider.With(vars.ProjectID))

	// Name all stack resources after the desired project ID
	id := resourceid.New(vars.ProjectID)

	output := &Output{
		Project: project.NewProject(stack,
			id.ResourceID("project"),
			&project.ProjectConfig{
				Name:              pointers.Ptr(vars.Name),
				ProjectId:         pointers.Ptr(vars.ProjectID),
				AutoCreateNetwork: false,
				BillingAccount:    pointers.Ptr(BillingAccountID),
				FolderId:          pointers.Ptr(ProjectFolderID),
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
