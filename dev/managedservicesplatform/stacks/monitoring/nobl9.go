package monitoring

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"

	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/serviceaccountkey"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/nobl9/directgcm"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/nobl9/project"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/nobl9/service"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/serviceaccount"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// WIP(jac): For evaluation purposes
func createNobl9Project(
	stack cdktf.TerraformStack,
	id resourceid.ID,
	vars Variables,
) {
	// GCP Setup
	sa := serviceaccount.New(stack, id.Group("nobl9-sa"), serviceaccount.Config{
		ProjectID:   vars.ProjectID,
		AccountID:   "nobl9-sa",
		DisplayName: "Nobl9 SA",
		Roles: []serviceaccount.Role{
			{
				ID:   id.Group("monitoring"),
				Role: "roles/monitoring.viewer",
			},
		},
	})

	key := serviceaccountkey.NewServiceAccountKey(stack, id.TerraformID("nobl9-key"), &serviceaccountkey.ServiceAccountKeyConfig{
		ServiceAccountId: &sa.Email,
	})

	// Nobl9
	id = id.Group("nobl9")
	project := project.NewProject(stack, id.TerraformID("project"), &project.ProjectConfig{
		Name:        pointers.Stringf("%s-%s", vars.Service.ID, vars.EnvironmentID),
		DisplayName: pointers.Stringf("%s - %s", vars.Service.GetName(), vars.EnvironmentID),
		Label: []project.ProjectLabel{
			{
				Key:    pointers.Ptr("category"),
				Values: pointers.Ptr(pointers.Slice([]string{"msp"})),
			},
		},
	})

	directgcm.NewDirectGcm(stack, id.TerraformID("gcm"), &directgcm.DirectGcmConfig{
		Name:              pointers.Stringf("%s-%s", vars.Service.ID, vars.EnvironmentID),
		DisplayName:       pointers.Stringf("%s - %s", vars.Service.GetName(), vars.EnvironmentID),
		Project:           project.Id(),
		SourceOf:          pointers.Ptr(pointers.Slice([]string{"Metrics"})),
		ReleaseChannel:    pointers.Ptr("beta"),
		ServiceAccountKey: cdktf.Fn_Base64decode(key.PrivateKey()),
	})

	service.NewService(stack, id.TerraformID("cloudrun-service"), &service.ServiceConfig{
		Name:    pointers.Stringf("%s-%s-%s", vars.Service.ID, vars.EnvironmentID, "cloudrun"),
		Project: project.Id(),
		Label: []service.ServiceLabel{
			{
				Key:    pointers.Ptr("service"),
				Values: pointers.Ptr(pointers.Slice([]string{"cloudrun"})),
			},
		},
	})
}
