package cloudrun

import (
	"fmt"
	"strconv"

	"github.com/aws/jsii-runtime-go"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/cloudrunv2service"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/project"

	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/internal/provider/google"
	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/internal/stack"
	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/spec"
)

type Output struct{}

type Variables struct {
	Project project.Project

	Service     spec.ServiceSpec
	Image       string
	Environment spec.EnvironmentSpec
}

const StackName = "cloudrun"

var (
	serviceAccountRoles = []string{
		"roles/secretmanager.secretAccessor",
		"roles/compute.networkUser",
		"roles/cloudtrace.agent",
		"roles/monitoring.metricWriter",
	}
	region = "us-central1"
)

// NewStack instantiates the MSP cloudrun stack, which is currently a pretty
// monolithic stack that encompasses all the core components of an MSP service,
// including networking and dependencies like Redis.
func NewStack(stacks *stack.Set, vars Variables) (*Output, error) {
	stack := stacks.New(StackName,
		google.StackWithProject(vars.Project))

	tag, err := vars.Environment.Deploy.ResolveTag()
	if err != nil {
		return nil, err
	}

	_ = cloudrunv2service.NewCloudRunV2Service(stack, jsii.String("default"),
		&cloudrunv2service.CloudRunV2ServiceConfig{
			Name:     jsii.String(vars.Service.ID),
			Location: jsii.String(region),
			//  Disallows direct traffic from public internet, we have a LB set up for that.
			Ingress: jsii.String("INGRESS_TRAFFIC_INTERNAL_LOAD_BALANCER"),

			Template: &cloudrunv2service.CloudRunV2ServiceTemplate{
				ServiceAccount: nil, // TODO

				// So set a high limit that matches our default Cloudflare zone's
				// timeout:
				//
				//   export CF_API_TOKEN=$(gcloud secrets versions access latest --secret CLOUDFLARE_API_TOKEN --project sourcegraph-secrets)
				//   curl -H "Authorization: Bearer $CF_API_TOKEN" https://api.cloudflare.com/client/v4/zones | jq '.result[]  | select(.name == "sourcegraph.com") | .id'
				//   curl -H "Authorization: Bearer $CF_API_TOKEN" https://api.cloudflare.com/client/v4/zones/$CF_ZONE_ID/settings | jq '.result[] | select(.id == "proxy_read_timeout")'
				//
				// Result should be something like:
				//
				//   {
				//     "id": "proxy_read_timeout",
				//     "value": "300",
				//     "modified_on": "2022-02-08T23:10:35.772888Z",
				//     "editable": true
				//   }
				//
				// The service should implement tighter timeouts on its own if desired.
				Timeout: jsii.String("300s"),

				MaxInstanceRequestConcurrency: jsii.Number(vars.Environment.Instances.Scaling.MaxRequestConcurrency),
				Scaling: &cloudrunv2service.CloudRunV2ServiceTemplateScaling{
					MinInstanceCount: jsii.Number(vars.Environment.Instances.Scaling.MinCount),
					MaxInstanceCount: jsii.Number(vars.Environment.Instances.Scaling.MaxCount),
				},

				Containers: []*cloudrunv2service.CloudRunV2ServiceTemplateContainers{{
					Name:  jsii.String(vars.Service.ID),
					Image: jsii.String(fmt.Sprintf("%s:%s", vars.Image, tag)),

					Resources: &cloudrunv2service.CloudRunV2ServiceTemplateContainersResources{
						Limits: &map[string]*string{
							"cpu":    jsii.String(strconv.Itoa(vars.Environment.Instances.Resources.CPU)),
							"memory": jsii.String(vars.Environment.Instances.Resources.Memory),
						},
					},

					Ports: []*cloudrunv2service.CloudRunV2ServiceTemplateContainersPorts{{
						// TODO: Should this be configurable?
						ContainerPort: jsii.Number(9992),
					}},

					// Env: &cloudrunv2service.CloudRunV2ServiceTemplateContainersEnv{
					// 	// TODO
					// },

					// StartupProbe: &cloudrunv2service.CloudRunV2ServiceTemplateContainersStartupProbe{
					// 	// TODO
					// },

					// LivenessProbe: &cloudrunv2service.CloudRunV2ServiceTemplateContainersLivenessProbe{
					// 	// TODO
					// },

					// VolumeMounts: &cloudrunv2service.CloudRunV2ServiceTemplateContainersVolumeMounts{
					// 	// TODO
					// },
				}},

				// Volumes: &cloudrunv2service.CloudRunV2ServiceTemplateVolumes{
				// 	// TODO
				// },

				// VpcAccess: &cloudrunv2service.CloudRunV2ServiceTemplateVpcAccess{
				// 	// TODO
				// },
			},
		})

	return &Output{}, nil
}
