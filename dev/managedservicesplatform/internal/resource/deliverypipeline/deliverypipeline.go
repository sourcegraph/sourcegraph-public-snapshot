package deliverypipeline

import (
	"github.com/aws/constructs-go/constructs/v10"

	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/clouddeploydeliverypipeline"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/deliverytarget"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type Config struct {
	Location string

	Name        string
	Description string

	Stages []*deliverytarget.Output
}

type Output struct{}

// New provisions resources for a google_clouddeploy_delivery_pipeline:
// https://cloud.google.com/deploy/docs/overview
func New(scope constructs.Construct, id resourceid.ID, config Config) (*Output, error) {
	_ = clouddeploydeliverypipeline.NewClouddeployDeliveryPipeline(scope,
		id.TerraformID("pipeline"),
		&clouddeploydeliverypipeline.ClouddeployDeliveryPipelineConfig{
			Location: &config.Location,

			Name:        &config.Name,
			Description: &config.Description,

			SerialPipeline: &clouddeploydeliverypipeline.ClouddeployDeliveryPipelineSerialPipeline{
				Stages: pointers.Ptr(newStages(config)),
			},
		})

	return &Output{}, nil
}

func newStages(config Config) []*clouddeploydeliverypipeline.ClouddeployDeliveryPipelineSerialPipelineStages {
	var stages []*clouddeploydeliverypipeline.ClouddeployDeliveryPipelineSerialPipelineStages
	for _, target := range config.Stages {
		stages = append(stages, &clouddeploydeliverypipeline.ClouddeployDeliveryPipelineSerialPipelineStages{
			TargetId: target.Target.Name(),
		})
	}
	return stages
}
