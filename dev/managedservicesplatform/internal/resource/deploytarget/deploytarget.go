package deploytarget

import (
	"github.com/aws/constructs-go/constructs/v10"

	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/clouddeploytarget"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type Config struct {
	Service spec.ServiceSpec

	CloudRunEnvironmentID    string
	CloudRunProjectID        string
	CloudRunResourceName     string
	CloudRunResourceLocation string
}

type Output struct{}

func New(scope constructs.Construct, id resourceid.ID, config Config) (*Output, error) {
	_ = clouddeploytarget.NewClouddeployTarget(scope, id.TerraformID("target"), &clouddeploytarget.ClouddeployTargetConfig{
		Description: pointers.Stringf("%s - %s",
			config.Service.GetName(), config.CloudRunEnvironmentID),

		// Configure Cloud Run as the target
		Name:     &config.CloudRunResourceName,
		Location: &config.CloudRunResourceLocation,
		Run: &clouddeploytarget.ClouddeployTargetRun{
			Location: pointers.Stringf("projects/%s/locations/%s",
				config.CloudRunProjectID, config.CloudRunResourceLocation),
		},

		// Target configuration
		RequireApproval: nil, // TODO
		ExecutionConfigs: &[]*clouddeploytarget.ClouddeployTargetExecutionConfigs{{
			Usages: &[]*string{
				pointers.Ptr("RENDER"),
				pointers.Ptr("DEPLOY"),
			},
			ExecutionTimeout: pointers.Ptr("3600s"),
		}},
	})
	return nil, nil
}
