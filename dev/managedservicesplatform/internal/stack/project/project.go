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
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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
	"cloudprofiler.googleapis.com",
	"cloudscheduler.googleapis.com",
}

const (
	BillingAccountID = "017005-C370B2-0E3030"
	// DefaultProjectFolderID points to the 'Managed Services' folder:
	// https://console.cloud.google.com/welcome?folder=26336759932
	DefaultProjectFolderID = "folders/26336759932"
)

var EnvironmentCategoryFolders = map[spec.EnvironmentCategory]string{
	// Engineering Projects - https://console.cloud.google.com/welcome?folder=795981974432
	spec.EnvironmentCategoryTest: "folders/795981974432",

	// Internal Projects - https://console.cloud.google.com/welcome?folder=387815626940
	spec.EnvironmentCategoryInternal: "folders/387815626940",

	// Use default folder - see DefaultProjectFolderID
	spec.EnvironmentCategoryExternal: DefaultProjectFolderID,
	spec.EnvironmentCategory(""):     DefaultProjectFolderID,
}

type CrossStackOutput struct {
	// Project is created with a generated project ID.
	Project project.Project
}

type Variables struct {
	// ProjectIDPrefix is the prefix for a project ID. A suffix of the format
	// '-${randomizedSuffix}' will be added, as project IDs must be unique.
	ProjectIDPrefix string

	// ProjectIDSuffixLength is the length of the randomized suffix added to
	// to the project.
	ProjectIDSuffixLength *int

	// DisplayName is a display name for the project. It does not need to be unique.
	DisplayName string

	// Labels to apply to the project.
	Labels map[string]string

	// Category determines what folder the project will be created in.
	Category *spec.EnvironmentCategory

	// EnableAuditLogs ships GCP audit logs to security cluster.
	// TODO: Not yet implemented
	EnableAuditLogs bool

	// Services is a list of additional GCP services to enable.
	Services []string
}

const StackName = "project"

const (
	// https://cloud.google.com/resource-manager/reference/rest/v1/projects
	projectIDMaxLength                 = 30
	projectIDRandomizedSuffixLength    = 6
	projectIDMinRandomizedSuffixLength = 2
)

// NewStack creates a stack that provisions a GCP project.
func NewStack(stacks *stack.Set, vars Variables) (*CrossStackOutput, error) {
	stack, locals, err := stacks.New(StackName,
		randomprovider.With(),
		// ID is not known ahead of time, we can omit it
		googleprovider.With(""))
	if err != nil {
		return nil, err
	}

	// Name all stack resources after the desired project ID
	id := resourceid.New(vars.ProjectIDPrefix)

	// The project ID must leave room for a randomized suffix and a separator.
	suffixLength := projectIDRandomizedSuffixLength
	if vars.ProjectIDSuffixLength != nil {
		suffixLength = *vars.ProjectIDSuffixLength / 2
	}
	realSuffixLength := suffixLength * 2 // after converting to hex
	if afterSuffixLength := len(vars.ProjectIDPrefix) + 1 + realSuffixLength; afterSuffixLength > projectIDMaxLength {
		return nil, errors.Newf("project ID prefix %q is too long after adding random suffix (%d characters) - got %d characters, but maximum is %d characters",
			vars.ProjectIDPrefix, projectIDRandomizedSuffixLength, afterSuffixLength, projectIDMaxLength)
	}
	projectID := random.New(stack, id, random.Config{
		ByteLength: suffixLength,
		Prefix:     vars.ProjectIDPrefix,
	})

	project := project.NewProject(stack,
		id.TerraformID("project"),
		&project.ProjectConfig{
			Name:              pointers.Ptr(vars.DisplayName),
			ProjectId:         &projectID.HexValue,
			AutoCreateNetwork: false,
			BillingAccount:    pointers.Ptr(BillingAccountID),
			FolderId: func() *string {
				folder, ok := EnvironmentCategoryFolders[pointers.Deref(vars.Category, spec.EnvironmentCategoryExternal)]
				if ok {
					return &folder
				}
				return pointers.Ptr(string(DefaultProjectFolderID))
			}(),
			Labels: func(input map[string]string) *map[string]*string {
				labels := make(map[string]*string)
				for k, v := range input {
					labels[sanitizeName(k)] = pointers.Ptr(v)
				}
				return &labels
			}(vars.Labels),
		})

	for _, service := range append(gcpServices, vars.Services...) {
		projectservice.NewProjectService(stack,
			id.TerraformID("project-service-%s", strings.ReplaceAll(service, ".", "-")),
			&projectservice.ProjectServiceConfig{
				Project:                  project.ProjectId(),
				Service:                  pointers.Ptr(service),
				DisableDependentServices: jsii.Bool(false),
				// prevent accidental deletion of services
				DisableOnDestroy: jsii.Bool(false),
			})
	}

	// Collect outputs
	locals.Add("project_id", project.ProjectId(), "Generated project ID")
	return &CrossStackOutput{Project: project}, nil
}

var regexpMatchNonLowerAlphaNumericUnderscoreDash = regexp.MustCompile(`[^a-z0-9_-]`)

// sanitizeName ensures the name contains only lowercase letters, numerals, underscores, and dashes.
// non-compliant characters are replaced with underscores.
func sanitizeName(name string) string {
	s := strings.ToLower(name)
	s = regexpMatchNonLowerAlphaNumericUnderscoreDash.ReplaceAllString(s, "_")
	return s
}
