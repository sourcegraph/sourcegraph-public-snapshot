package deliverypipeline

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/hashicorp/terraform-cdk-go/cdktf"

	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/clouddeploycustomtargettype"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/clouddeploydeliverypipeline"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/clouddeploytarget"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/serviceaccount"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type Target struct {
	ID        string
	ProjectID string
}

type Config struct {
	// Name and Description for the delivery pipeline.
	Name        string
	Description string
	// Location currently must also be the location of all targets being
	// deployed by this pipeline.
	Location string

	// ServiceID and ServiceImage are used to configure target defaults.
	ServiceID    string
	ServiceImage string

	// ExecutionSA is the service account that will be used to apply deployments.
	ExecutionSA *serviceaccount.Output

	// TargetStages lists targets in the order they should appear in the pipeline.
	TargetStages []Target

	// Repository is the source code repository for the images delivered to this pipeline.
	Repository string

	// Suspended prevents releases and rollouts from being created, rolled back,
	// etc using this pipeline: https://cloud.google.com/deploy/docs/suspend-pipeline
	Suspended bool

	DependsOn []cdktf.ITerraformDependable
}

type Output struct {
	PipelineID string
}

// New provisions resources for a google_clouddeploy_delivery_pipeline:
// https://cloud.google.com/deploy/docs/overview
func New(scope constructs.Construct, id resourceid.ID, config Config) (*Output, error) {
	targetType := clouddeploycustomtargettype.NewClouddeployCustomTargetType(
		scope,
		id.TerraformID("custom_cloudrun_targettype"),
		&clouddeploycustomtargettype.ClouddeployCustomTargetTypeConfig{
			Name:     pointers.Stringf("cloud-run-service"),
			Location: pointers.Stringf(config.Location),
			Labels: &map[string]*string{
				"msp": pointers.Stringf("true"),
			},
			Description: pointers.Stringf("MSP Cloud Run Service"),
			CustomActions: &clouddeploycustomtargettype.ClouddeployCustomTargetTypeCustomActions{
				DeployAction: pointers.Stringf("cloud-run-image-deploy"),
				RenderAction: pointers.Stringf("cloud-run-image-deploy-render"),
				// We can point this to the GCS bucket we generate, but it's
				// unclear why we need to, because the '--source' parameter
				// pointing to the same bucket (or something with a skaffold.yaml)
				// is strangely a required parameter in 'gcloud deploy releases create':
				//
				// - https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/clouddeploy_custom_target_type#include_skaffold_modules
				// - https://cloud.google.com/sdk/gcloud/reference/deploy/releases/create#--source
				//
				// Because of this strange behaviour, we omit this for now.
				// IncludeSkaffoldModules: []clouddeploycustomtargettype.ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModules{{}},
			},
		})

	pipeline := clouddeploydeliverypipeline.NewClouddeployDeliveryPipeline(scope,
		id.TerraformID("pipeline"),
		&clouddeploydeliverypipeline.ClouddeployDeliveryPipelineConfig{
			Location: &config.Location,

			Name:        &config.Name,
			Description: &config.Description,
			Suspended:   &config.Suspended,

			SerialPipeline: &clouddeploydeliverypipeline.ClouddeployDeliveryPipelineSerialPipeline{
				Stages: pointers.Ptr(newStages(config)),
			},

			Annotations: &map[string]*string{
				"source.repository": pointers.Ptr(config.Repository),
			},

			DependsOn: pointers.Ptr(append(config.DependsOn, targetType)),
		})

	for _, target := range config.TargetStages {
		id := id.Group("target_stage")
		_ = clouddeploytarget.NewClouddeployTarget(scope, id.TerraformID(target.ID), &clouddeploytarget.ClouddeployTargetConfig{
			Name:     &target.ID,
			Location: &config.Location,
			CustomTarget: &clouddeploytarget.ClouddeployTargetCustomTarget{
				CustomTargetType: targetType.Id(),
			},
			DeployParameters: &map[string]*string{
				"customTarget/serviceID": &config.ServiceID,
				"customTarget/image":     &config.ServiceImage,
				"customTarget/projectID": &target.ProjectID,
				// Tag must be provided in 'gcloud deploy releases create' via the
				// flag '--deploy-parameters="customTarget/tag=$TAG"'.
				// customTarget/tag: nil,
			},
			ExecutionConfigs: []*clouddeploytarget.ClouddeployTargetExecutionConfigs{{
				Usages:         pointers.Ptr(pointers.Slice([]string{"RENDER", "DEPLOY"})),
				ServiceAccount: &config.ExecutionSA.Email,
			}},

			DependsOn: pointers.Ptr(append(config.DependsOn, targetType)),
		})
	}

	return &Output{
		PipelineID: *pipeline.Uid(),
	}, nil
}

func newStages(config Config) []*clouddeploydeliverypipeline.ClouddeployDeliveryPipelineSerialPipelineStages {
	var stages []*clouddeploydeliverypipeline.ClouddeployDeliveryPipelineSerialPipelineStages
	for _, target := range config.TargetStages {
		stages = append(stages, &clouddeploydeliverypipeline.ClouddeployDeliveryPipelineSerialPipelineStages{
			TargetId: pointers.Ptr(target.ID),
		})
	}
	return stages
}
