package privatenetwork

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computenetwork"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computesubnetwork"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/vpcaccessconnector"

	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type Config struct {
	ProjectID string
	ServiceID string
	Region    string
}

type Output struct {
	Network    computenetwork.ComputeNetwork
	Subnetwork computesubnetwork.ComputeSubnetwork
	Connector  vpcaccessconnector.VpcAccessConnector
}

// New sets up a network for the Cloud Run service to interface
// with other GCP services. We create it directly in this stack package becasue
// it is very specific to the Cloud Run service.
func New(scope constructs.Construct, config Config) *Output {
	network := computenetwork.NewComputeNetwork(
		scope,
		pointers.Ptr("cloudrun-network"),
		&computenetwork.ComputeNetworkConfig{
			Project: &config.ProjectID,
			Name:    &config.ServiceID,
			// We will manually create a subnet.
			AutoCreateSubnetworks: false,
		})

	subnetwork := computesubnetwork.NewComputeSubnetwork(
		scope,
		pointers.Ptr("cloudrun-subnetwork"),
		&computesubnetwork.ComputeSubnetworkConfig{
			Project: &config.ProjectID,
			Region:  &config.Region,
			Name:    &config.ServiceID,
			Network: network.Id(),

			// This is similar to the setup in Cloud v1.1 for connecting to Cloud SQL - we
			// set up an arbitrary ip_cidr_range that covers enough IPs for most needs. The
			// private_x_google_access stuff is based on security requirements.
			// We must use a /28 range because that's the range supported by VPC connectors.
			IpCidrRange:           pointers.Ptr("10.0.0.0/28"),
			PrivateIpGoogleAccess: false,
			//checkov:skip=CKV_GCP_76: Enable dual-stack support for subnetworks is destrutive and require re-creating the subnet and all dependent resources (e.g. NEG)
			PrivateIpv6GoogleAccess: pointers.Ptr("DISABLE_GOOGLE_ACCESS"),
			// Checkov requirement: https://docs.bridgecrew.io/docs/bc_gcp_logging_1
			LogConfig: &computesubnetwork.ComputeSubnetworkLogConfig{
				AggregationInterval: pointers.Ptr("INTERVAL_10_MIN"),
				FlowSampling:        pointers.Float64(0.5),
				Metadata:            pointers.Ptr("INCLUDE_ALL_METADATA"),
			},
		},
	)

	// Cloud Run services can't connect directly to networks, and seem to require a
	// VPC connector, so we provision one to allow Cloud Run services to talk to
	// other GCP services (like Redis)
	connector := vpcaccessconnector.NewVpcAccessConnector(
		scope,
		pointers.Ptr("cloudrun-connector"),
		&vpcaccessconnector.VpcAccessConnectorConfig{
			Project: &config.ProjectID,
			Region:  &config.Region,
			Name:    pointers.Ptr(config.ServiceID),
			Subnet: &vpcaccessconnector.VpcAccessConnectorSubnet{
				Name: subnetwork.Name(),
			},
		},
	)

	return &Output{
		Network:    network,
		Subnetwork: subnetwork,
		Connector:  connector,
	}
}
