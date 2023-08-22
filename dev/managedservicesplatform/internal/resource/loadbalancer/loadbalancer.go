package loadbalancer

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/cloudrunv2service"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computebackendservice"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computeregionnetworkendpointgroup"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computeurlmap"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/project"

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

// New instantiates a set of resources for a load-balancer frontend for a Cloud
// Run service.
func New(scope constructs.Construct, id string, config Config) *Output {
	// Endpoint represents the Cloud Run service.
	endpoint := computeregionnetworkendpointgroup.NewComputeRegionNetworkEndpointGroup(scope,
		pointer.Stringf("%s-endpoint", id),
		&computeregionnetworkendpointgroup.ComputeRegionNetworkEndpointGroupConfig{
			Name:    pointer.Value(id),
			Project: config.Project.ProjectId(),
			Region:  pointer.Value(config.Region),

			NetworkEndpointType: pointer.Value("SERVERLESS"),
			CloudRun: &computeregionnetworkendpointgroup.ComputeRegionNetworkEndpointGroupCloudRun{
				Service: config.TargetService.Name(),
			},
		})

	// Set up a group of virtual machines that will serve traffic for load balancing
	backendService := computebackendservice.NewComputeBackendService(scope,
		pointer.Stringf("%s-backend-service", id),
		&computebackendservice.ComputeBackendServiceConfig{
			Name:    pointer.Value(id),
			Project: config.Project.ProjectId(),

			Protocol: pointer.Value("HTTP"),
			PortName: pointer.Value("http"),

			// TODO: Parameterize
			SecurityPolicy: nil,

			Backend: endpoint.Id(),
		})

	// Enable routing requests to the backend service working serving traffic
	// for load balancing
	urlMap := computeurlmap.NewComputeUrlMap(scope,
		pointer.Stringf("%s-urlmap", id),
		&computeurlmap.ComputeUrlMapConfig{
			Name:           pointer.Value(id),
			Project:        config.Project.ProjectId(),
			DefaultService: backendService.Id(),
		})

	return &Output{
		URLMap: urlMap,
	}
}
