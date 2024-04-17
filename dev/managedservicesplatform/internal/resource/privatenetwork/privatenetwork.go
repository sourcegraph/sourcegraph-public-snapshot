package privatenetwork

import (
	"fmt"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computeglobaladdress"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computenetwork"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computesubnetwork"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/servicenetworkingconnection"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/random"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type Config struct {
	ProjectID string
	ServiceID string
	Region    string
}

type Output struct {
	// Network is the private network for GCP resources that the Cloud Run
	// workload needs to access.
	Network    computenetwork.ComputeNetwork
	Subnetwork computesubnetwork.ComputeSubnetwork
	// ServiceNetworkingConnection is required for Cloud SQL access, and is
	// provisioned by default.
	ServiceNetworkingConnection servicenetworkingconnection.ServiceNetworkingConnection
}

// New sets up a network for the Cloud Run service to interface with other GCP
// services. This should only be created once, hence why it does not have accept
// a resourceid.ID
func New(scope constructs.Construct, id resourceid.ID, config Config) *Output {
	network := computenetwork.NewComputeNetwork(
		scope,
		pointers.Ptr("cloudrun-network"),
		&computenetwork.ComputeNetworkConfig{
			Project: &config.ProjectID,
			Name:    &config.ServiceID,
			// We will manually create a subnet.
			AutoCreateSubnetworks: false,
		})

	// Cloud Run supports a small set of IPv4 ranges for the subnet:
	// https://cloud.google.com/run/docs/configuring/vpc-direct-vpc#supported-ip-ranges
	// We choose one with a generous IP range to avoid limitations documented
	// here: https://cloud.google.com/run/docs/configuring/vpc-direct-vpc#limitations
	subnetworkIPCIDRRange := "172.16.0.0/12"
	subnetworkName := random.New(scope, id.Group("subnetwork-name"), random.Config{
		Prefix:     config.ServiceID,
		ByteLength: 4,
		Keepers: map[string]*string{
			// Range change requires recreation of the subnetwork, so we need
			// to change the randomized suffix to avoid a conflict and respect
			// CreateBeforeDestroy
			"ipcidrrange": pointers.Ptr(subnetworkIPCIDRRange),
		},
	})
	subnetwork := computesubnetwork.NewComputeSubnetwork(
		scope,
		pointers.Ptr("cloudrun-subnetwork"),
		&computesubnetwork.ComputeSubnetworkConfig{
			Project:     &config.ProjectID,
			Region:      &config.Region,
			Name:        &subnetworkName.HexValue,
			Network:     network.Id(),
			IpCidrRange: pointers.Ptr(subnetworkIPCIDRRange),

			// Allow usage of private Google access: https://cloud.google.com/vpc/docs/private-google-access
			PrivateIpGoogleAccess: true,

			//checkov:skip=CKV_GCP_76: Enable dual-stack support for subnetworks is destrutive and require re-creating the subnet and all dependent resources (e.g. NEG)
			PrivateIpv6GoogleAccess: pointers.Ptr("DISABLE_GOOGLE_ACCESS"),
			// Checkov requirement: https://docs.bridgecrew.io/docs/bc_gcp_logging_1
			LogConfig: &computesubnetwork.ComputeSubnetworkLogConfig{
				AggregationInterval: pointers.Ptr("INTERVAL_10_MIN"),
				FlowSampling:        pointers.Float64(0.5),
				Metadata:            pointers.Ptr("INCLUDE_ALL_METADATA"),
			},

			Lifecycle: &cdktf.TerraformResourceLifecycle{
				// Recreation also requires rename of randomized subnetworkName.
				CreateBeforeDestroy: pointers.Ptr(true),
			},
		},
	)

	// https://cloud.google.com/sql/docs/mysql/private-ip#network_requirements
	// This is configured per project and usually doesn't change
	serviceNetworkingConnectionIP := computeglobaladdress.NewComputeGlobalAddress(
		scope,
		pointers.Ptr("cloudrun-network-service-networking-connection-ip"),
		&computeglobaladdress.ComputeGlobalAddressConfig{
			Project:      &config.ProjectID,
			Name:         pointers.Ptr(fmt.Sprintf("%s-service-networking-connection", config.ServiceID)),
			Network:      network.Id(),
			Purpose:      pointers.Ptr("VPC_PEERING"),
			AddressType:  pointers.Ptr("INTERNAL"),
			PrefixLength: pointers.Float64(16),
		})
	serviceNetworkingConnection := servicenetworkingconnection.NewServiceNetworkingConnection(
		scope,
		pointers.Ptr("cloudrun-network-service-networking-connection"),
		&servicenetworkingconnection.ServiceNetworkingConnectionConfig{
			Network:               network.Id(),
			Service:               pointers.Ptr("servicenetworking.googleapis.com"),
			ReservedPeeringRanges: &[]*string{serviceNetworkingConnectionIP.Name()},
		})

	return &Output{
		Network:                     network,
		Subnetwork:                  subnetwork,
		ServiceNetworkingConnection: serviceNetworkingConnection,
	}
}
