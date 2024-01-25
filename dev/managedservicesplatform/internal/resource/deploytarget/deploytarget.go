package deploystrategy

import (
	"github.com/aws/constructs-go/constructs/v10"

	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/clouddeploytarget"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type Config struct {
	Service spec.ServiceSpec
}

type Output struct{}

func New(scope constructs.Construct, id resourceid.ID, config Config) (*Output, error) {
	_ = clouddeploytarget.NewClouddeployTarget(scope, id.TerraformID("target"), &clouddeploytarget.ClouddeployTargetConfig{
		Project:          nil,
		Location:         nil,
		Name:             nil,
		DeployParameters: &map[string]*string{},
		ExecutionConfigs: &[]*clouddeploytarget.ClouddeployTargetExecutionConfigs{{
			Usages: &[]*string{
				pointers.Ptr("RENDER"),
				pointers.Ptr("DEPLOY"),
			},
			ExecutionTimeout: pointers.Ptr("3600s"),
		}},
		RequireApproval: nil,
		Run: &clouddeploytarget.ClouddeployTargetRun{
			Location: nil, // `projects/{project}/locations/{location}`
		},
	})
	return nil, nil
}
