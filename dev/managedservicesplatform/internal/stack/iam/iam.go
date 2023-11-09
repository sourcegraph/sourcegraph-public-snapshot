package iam

import (
	"fmt"
	"strings"

	"github.com/grafana/regexp"

	"github.com/aws/jsii-runtime-go"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/projectiamcustomrole"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/projectiammember"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google_beta/googleprojectserviceidentity"
	google_beta "github.com/sourcegraph/managed-services-platform-cdktf/gen/google_beta/provider"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/serviceaccount"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack/options/googleprovider"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type CrossStackOutput struct {
	CloudRunWorkloadServiceAccount *serviceaccount.Output
}

type Variables struct {
	ProjectID string
	Image     string
	Service   spec.ServiceSpec
}

const StackName = "iam"

var (
	// serviceAccountRoles are granted to the service account for the Cloud Run service.
	serviceAccountRoles = []serviceaccount.Role{
		// Allow env vars to source from secrets
		{ID: resourceid.New("role_secret_accessor"), Role: "roles/secretmanager.secretAccessor"},
		// Allow service to access private networks
		{ID: resourceid.New("role_compute_networkuser"), Role: "roles/compute.networkUser"},
		// Allow service to emit observability
		{ID: resourceid.New("role_cloudtrace_agent"), Role: "roles/cloudtrace.agent"},
		{ID: resourceid.New("role_monitoring_metricwriter"), Role: "roles/monitoring.metricWriter"},
		// Allow service to publish Cloud Profiler profiles
		{ID: resourceid.New("role_cloudprofiler_agent"), Role: "roles/cloudprofiler.agent"},
	}
)

func NewStack(stacks *stack.Set, vars Variables) (*CrossStackOutput, error) {
	stack, locals, err := stacks.New(StackName,
		googleprovider.With(vars.ProjectID))
	if err != nil {
		return nil, err
	}

	id := resourceid.New("iam")

	var customRole projectiamcustomrole.ProjectIamCustomRole
	const customRoleID = "msp_workload_custom_role"
	if vars.Service.IAM != nil && len(vars.Service.IAM.Permissions) > 0 {
		customRole = projectiamcustomrole.NewProjectIamCustomRole(stack,
			id.TerraformID("custom-role"),
			&projectiamcustomrole.ProjectIamCustomRoleConfig{
				RoleId:      pointers.Ptr(customRoleID),
				Title:       pointers.Ptr("Managed Services Platform workload custom role"),
				Project:     &vars.ProjectID,
				Permissions: jsii.Strings(vars.Service.IAM.Permissions...),
			})
	}
	workloadServiceAccount := serviceaccount.New(stack,
		id.Group("workload"),
		serviceaccount.Config{
			ProjectID: vars.ProjectID,
			AccountID: fmt.Sprintf("%s-sa", vars.Service.ID),
			DisplayName: fmt.Sprintf("%s Service Account",
				pointers.Deref(vars.Service.Name, vars.Service.ID)),
			Roles: func() []serviceaccount.Role {
				if vars.Service.IAM != nil && len(vars.Service.IAM.Roles) > 0 {
					var rs []serviceaccount.Role
					for _, r := range vars.Service.IAM.Roles {
						rs = append(rs, serviceaccount.Role{
							ID:   resourceid.New(matchNonAlphaNumericRegex.ReplaceAllString(r, "_")),
							Role: r,
						})
					}
					serviceAccountRoles = append(rs, serviceAccountRoles...)
				}
				if customRole != nil {
					serviceAccountRoles = append(serviceAccountRoles, serviceaccount.Role{
						ID:   resourceid.New(customRoleID),
						Role: *customRole.Name(),
					})
				}
				return serviceAccountRoles
			}(),
		})

	// If the service image is published to a private image repository, grant
	// the Cloud Run robot account access to the target project to pull the
	// image.
	cloudRunIdentity := googleprojectserviceidentity.NewGoogleProjectServiceIdentity(stack,
		id.TerraformID("cloudrun-identity"),
		&googleprojectserviceidentity.GoogleProjectServiceIdentityConfig{
			Project: &vars.ProjectID,
			Service: pointers.Ptr("run.googleapis.com"),
			// Only available via beta provider:
			// https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/project_service_identity
			Provider: google_beta.NewGoogleBetaProvider(stack, pointers.Ptr("google_beta"), &google_beta.GoogleBetaProviderConfig{
				Project: &vars.ProjectID,
			}),
		})
	if imageProject := extractImageGoogleProject(vars.Image); imageProject != nil {
		for _, r := range []serviceaccount.Role{{
			ID:   resourceid.New("object_viewer"),
			Role: "roles/storage.objectViewer", // for gcr.io
		}, {
			ID:   resourceid.New("artifact_reader"),
			Role: "roles/artifactregistry.reader", // for artifact registry
		}} {
			projectiammember.NewProjectIamMember(stack,
				id.TerraformID("member_image_access_%s", r.ID),
				&projectiammember.ProjectIamMemberConfig{
					Project: imageProject,
					Role:    pointers.Ptr(r.Role),
					Member: pointers.Ptr(fmt.Sprintf("serviceAccount:%s",
						*cloudRunIdentity.Email())),
				})
		}
	}

	// Collect outputs
	locals.Add("cloud_run_service_account", workloadServiceAccount.Email,
		"Service Account email used as Cloud Run resource workload identity")
	return &CrossStackOutput{
		CloudRunWorkloadServiceAccount: workloadServiceAccount,
	}, nil
}

var matchNonAlphaNumericRegex = regexp.MustCompile("[^a-zA-Z0-9]+")

func extractImageGoogleProject(image string) *string {
	// Private images in our GCP projects have patterns like:
	// - us-central1-docker.pkg.dev/control-plane-5e9ee072/docker/apiserver
	// - us.gcr.io/sourcegraph-dev/abuse-banbot
	// If the root matches particular patterns, the second component is the
	// project ID.
	imageRepoParts := strings.SplitN(image, "/", 3)
	if len(imageRepoParts) != 3 {
		return nil
	}
	repoRoot := imageRepoParts[0]
	if strings.HasSuffix(repoRoot, ".gcr.io") ||
		strings.HasSuffix(repoRoot, "-docker.pkg.dev") {
		return &imageRepoParts[1]
	}
	return nil
}
