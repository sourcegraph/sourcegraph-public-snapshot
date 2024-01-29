package deliverytarget

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/hashicorp/terraform-cdk-go/cdktf"

	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/clouddeploytarget"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/serviceaccount"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/stacks/cloudrun/cloudrunresource"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type Config struct {
	Service spec.ServiceSpec

	CloudRunEnvironmentID    string
	CloudRunProjectID        string
	CloudRunResourceLocation string

	RequireApproval bool

	ExecutionServiceAccount *serviceaccount.Output

	DependsOn []cdktf.ITerraformDependable
}

type Output struct {
	Target clouddeploytarget.ClouddeployTarget
}

// New provisions resources for a google_clouddeploy_target:
// https://cloud.google.com/deploy/docs/overview
func New(scope constructs.Construct, id resourceid.ID, config Config) (*Output, error) {
	cloudRunServiceName := pointers.Ptr(cloudrunresource.NewName(
		config.Service.ID, config.CloudRunEnvironmentID, config.CloudRunResourceLocation))

	// TODO REMOVE THIS WHEN test IS UPDATED WITH NEW NAMING SCHEME
	if config.CloudRunEnvironmentID == "test" {
		cloudRunServiceName = &config.Service.ID
	}

	target := clouddeploytarget.NewClouddeployTarget(scope, id.TerraformID("target"), &clouddeploytarget.ClouddeployTargetConfig{
		Description: pointers.Stringf("%s - %s",
			config.Service.GetName(), config.CloudRunEnvironmentID),

		// Configure Cloud Run as the target
		Name: cloudRunServiceName,

		Location: &config.CloudRunResourceLocation,
		Run: &clouddeploytarget.ClouddeployTargetRun{
			Location: pointers.Stringf("projects/%s/locations/%s",
				config.CloudRunProjectID, config.CloudRunResourceLocation),
		},

		// Target configuration
		RequireApproval: &config.RequireApproval, // TODO
		ExecutionConfigs: &[]*clouddeploytarget.ClouddeployTargetExecutionConfigs{{
			Usages: &[]*string{
				pointers.Ptr("RENDER"),
				pointers.Ptr("DEPLOY"),
			},
			ExecutionTimeout: pointers.Ptr("3600s"),
			ServiceAccount:   &config.ExecutionServiceAccount.Email,
		}},

		DependsOn: &config.DependsOn,
	})
	return &Output{
		Target: target,
	}, nil
}
