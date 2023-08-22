package loadbalancer

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/cloudrunv2service"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computebackendservice"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computeregionnetworkendpointgroup"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computeurlmap"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/project"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/internal/pointer"
)

type Output struct {
	URLMap computeurlmap.ComputeUrlMap
}

type Config struct {
	Project project.Project
	Region  string

	TargetService      cloudrunv2service.CloudRunV2Service
	TargetHTTPPortName string
}

// New instantiates a set of resources for a load-balancer backend that routes
// requests to a Cloud Run service:
//
//	URLMap -> BackendService -> NetworkEndpointGroup -> CloudRun
//
// Typically some other frontend will then be placed in front of URLMap, e.g.
// resource/cloudflare.
func New(scope constructs.Construct, id resourceid.ID, config Config) *Output {
	// Endpoint group represents the Cloud Run service.
	endpointGroup := computeregionnetworkendpointgroup.NewComputeRegionNetworkEndpointGroup(scope,
		id.ResourceID("endpoint_group"),
		&computeregionnetworkendpointgroup.ComputeRegionNetworkEndpointGroupConfig{
			Name:    pointer.Value(id.DisplayName()),
			Project: config.Project.ProjectId(),
			Region:  pointer.Value(config.Region),

			NetworkEndpointType: pointer.Value("SERVERLESS"),
			CloudRun: &computeregionnetworkendpointgroup.ComputeRegionNetworkEndpointGroupCloudRun{
				Service: config.TargetService.Name(),
			},
		})

	// Set up a group of virtual machines that will serve traffic for load balancing
	backendService := computebackendservice.NewComputeBackendService(scope,
		id.ResourceID("backend_service"),
		&computebackendservice.ComputeBackendServiceConfig{
			Name:    pointer.Value(id.DisplayName()),
			Project: config.Project.ProjectId(),

			Protocol: pointer.Value("HTTP"),
			PortName: pointer.Value("http"),

			// TODO: Parameterize
			SecurityPolicy: nil,

			Backend: []*computebackendservice.ComputeBackendServiceBackend{{
				Group: endpointGroup.Id(),
			}},
		})

	// Enable routing requests to the backend service working serving traffic
	// for load balancing
	urlMap := computeurlmap.NewComputeUrlMap(scope,
		id.ResourceID("url_map"),
		&computeurlmap.ComputeUrlMapConfig{
			Name:           pointer.Value(id.DisplayName()),
			Project:        config.Project.ProjectId(),
			DefaultService: backendService.Id(),
		})

	return &Output{
		URLMap: urlMap,
	}
}
