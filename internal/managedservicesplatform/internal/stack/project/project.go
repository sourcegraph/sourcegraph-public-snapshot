package project

import (
	"fmt"
	"strings"

	"github.com/aws/jsii-runtime-go"
	"github.com/grafana/regexp"

	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/project"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/projectservice"

	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/internal/provider/google"
	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/internal/stack"
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

type Output struct {
	Project project.Project
}

type Variables struct {
	ProjectID        string `validate:"required"`
	Name             string `validate:"required"`
	ParentFolderID   string `validate:"required"`
	BillingAccountID string `validate:"required"`
	// Ship audit logs to security cluster
	EnableAuditLogs bool
	Labels          map[string]string
}

const StackName = "project"

func NewStack(stacks *stack.Set, vars Variables) (*Output, error) {
	stack := stacks.New(StackName,
		google.WithProjectID(vars.ProjectID))

	output := &Output{
		Project: project.NewProject(stack, jsii.String("project"), &project.ProjectConfig{
			Name:              jsii.String(vars.Name),
			ProjectId:         jsii.String(vars.ProjectID),
			AutoCreateNetwork: false,
			BillingAccount:    jsii.String(vars.BillingAccountID),
			FolderId:          jsii.String(vars.ParentFolderID),
			Labels: func(input map[string]string) *map[string]*string {
				labels := make(map[string]*string)
				for k, v := range input {
					labels[sanitizeName(k)] = jsii.String(v)
				}
				return &labels
			}(vars.Labels),
		}),
	}

	for i, service := range gcpServices {
		projectservice.NewProjectService(stack, jsii.String(fmt.Sprintf("project_service_%d", i)), &projectservice.ProjectServiceConfig{
			Project:                  output.Project.ProjectId(),
			Service:                  jsii.String(service),
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
