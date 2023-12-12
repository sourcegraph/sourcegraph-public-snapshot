package iam

import (
	"fmt"
	"sort"
	"strings"

	"github.com/grafana/regexp"
	"golang.org/x/exp/maps"

	"github.com/aws/jsii-runtime-go"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/projectiamcustomrole"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/projectiammember"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/secretmanagersecretiammember"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google_beta/googleprojectserviceidentity"
	google_beta "github.com/sourcegraph/managed-services-platform-cdktf/gen/google_beta/provider"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/serviceaccount"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack/options/googleprovider"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type CrossStackOutput struct {
	CloudRunWorkloadServiceAccount *serviceaccount.Output
}

type Variables struct {
	ProjectID string
	Image     string
	Service   spec.ServiceSpec

	// SecretEnv should be the environment config that sources from secrets.
	SecretEnv map[string]string
}

const StackName = "iam"

func NewStack(stacks *stack.Set, vars Variables) (*CrossStackOutput, error) {
	stack, locals, err := stacks.New(StackName,
		googleprovider.With(vars.ProjectID))
	if err != nil {
		return nil, err
	}

	id := resourceid.New("iam")

	// serviceAccountRoles are granted to the service account for the Cloud Run service.
	serviceAccountRoles := []serviceaccount.Role{
		// Allow env vars to source from secrets
		{ID: resourceid.New("role_secret_accessor"), Role: "roles/secretmanager.secretAccessor"},
		// Allow service to access private networks
		{ID: resourceid.New("role_compute_networkuser"), Role: "roles/compute.networkUser"},
		// Allow service to emit observability
		{ID: resourceid.New("role_cloudtrace_agent"), Role: "roles/cloudtrace.agent"},
		{ID: resourceid.New("role_monitoring_metricwriter"), Role: "roles/monitoring.metricWriter"},
		// Allow service to publish Cloud Profiler profiles
		{ID: resourceid.New("role_cloudprofiler_agent"), Role: "roles/cloudprofiler.agent"},
		// Allow service to connect to Cloud SQL
		{ID: resourceid.New("role_cloudsql_client"), Role: "roles/cloudsql.client"},
		{ID: resourceid.New("role_cloudsql_instanceuser"), Role: "roles/cloudsql.instanceUser"},
	}

	// Grant configured roles to the workload identity
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

	// Configure a role with custom permissions and grant it to the workload
	// identity
	if vars.Service.IAM != nil && len(vars.Service.IAM.Permissions) > 0 {
		customRole := projectiamcustomrole.NewProjectIamCustomRole(stack,
			id.TerraformID("custom-role"),
			&projectiamcustomrole.ProjectIamCustomRoleConfig{
				RoleId:      pointers.Ptr("msp_workload_custom_role"),
				Title:       pointers.Ptr("Managed Services Platform workload custom role"),
				Project:     &vars.ProjectID,
				Permissions: jsii.Strings(vars.Service.IAM.Permissions...),
			})
		serviceAccountRoles = append(serviceAccountRoles, serviceaccount.Role{
			ID:   resourceid.New("msp_workload_custom_role"),
			Role: *customRole.Name(),
		})
	}

	// Create a service account for the workload identity in Cloud Run
	workloadServiceAccount := serviceaccount.New(stack,
		id.Group("workload"),
		serviceaccount.Config{
			ProjectID: vars.ProjectID,
			AccountID: fmt.Sprintf("%s-sa", vars.Service.ID),
			DisplayName: fmt.Sprintf("%s Service Account",
				pointers.Deref(vars.Service.Name, vars.Service.ID)),
			Roles: serviceAccountRoles,
		})

	// Provision the default Cloud Run robot account so that we can grant it
	// access to prerequisite resources.
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
	identityMember := pointers.Ptr(fmt.Sprintf("serviceAccount:%s", *cloudRunIdentity.Email()))

	// If the service image is published to a private image repository, grant
	// the Cloud Run robot account access to the target project to pull the
	// image.
	if imageProject := extractImageGoogleProject(vars.Image); imageProject != nil {
		imageAccessID := id.Group("image_access")
		for _, r := range []serviceaccount.Role{{
			ID:   resourceid.New("object_viewer"),
			Role: "roles/storage.objectViewer", // for gcr.io
		}, {
			ID:   resourceid.New("artifact_reader"),
			Role: "roles/artifactregistry.reader", // for artifact registry
		}} {
			projectiammember.NewProjectIamMember(stack,
				imageAccessID.Append(r.ID).TerraformID("member"),
				&projectiammember.ProjectIamMemberConfig{
					Project: imageProject,
					Role:    pointers.Ptr(r.Role),
					Member:  identityMember,
				})
		}
	}

	// If any secret env secrets are external to this project, grant access to
	// the referenced secrets.
	if externalSecrets, err := extractExternalSecrets(vars.SecretEnv); err != nil {
		return nil, errors.Wrap(err, "extracting secret projects")
	} else {
		secretAccessID := id.Group("external_secret_access")
		for _, p := range externalSecrets {
			secretmanagersecretiammember.NewSecretManagerSecretIamMember(stack,
				secretAccessID.Group(p.key).TerraformID("member"),
				&secretmanagersecretiammember.SecretManagerSecretIamMemberConfig{
					Project:  &p.projectID,
					SecretId: &p.secretID,
					Role:     pointers.Ptr("roles/secretmanager.secretAccessor"),
					Member:   identityMember,
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

type externalSecret struct {
	key string // original configuration key, lowercased

	projectID string
	secretID  string
}

func extractExternalSecrets(secrets map[string]string) ([]externalSecret, error) {
	var externalSecrets []externalSecret

	// Sort for stability
	secretKeys := maps.Keys(secrets)
	sort.Strings(secretKeys)
	for _, k := range secretKeys {
		secretName := secrets[k]
		// Error on easy-to-make oopsies
		if strings.HasPrefix(secretName, "project/") {
			return nil, errors.New("invalid secret name %q: 'project/'-prefixed name provided, did you mean 'projects/'?")
		}
		// Check for 'projects/{project}/secrets/{secretName}'
		if strings.HasPrefix(secretName, "projects/") {
			secretNameParts := strings.SplitN(secretName, "/", 4)
			if len(secretNameParts) != 4 {
				return nil, errors.Newf("invalid secret name %q: expected 'projects/'-prefixed name to have 4 '/'-delimited parts",
					secretName)
			}
			// Error on easy-to-make oopsies
			if secretNameParts[2] != "secrets" {
				return nil, errors.Newf("invalid secret name %q: found '/secret/' segment, did you mean '/secrets/'?",
					secretName)
			}

			externalSecrets = append(externalSecrets, externalSecret{
				key:       strings.ToLower(k),
				projectID: secretNameParts[1],
				secretID:  secretNameParts[3],
			})
		}
	}

	return externalSecrets, nil
}
