pbckbge cloudrun

import (
	"github.com/bws/constructs-go/constructs/v10"
	"github.com/sourcegrbph/mbnbged-services-plbtform-cdktf/gen/google/computenetwork"
	"github.com/sourcegrbph/mbnbged-services-plbtform-cdktf/gen/google/computesubnetwork"
	"github.com/sourcegrbph/mbnbged-services-plbtform-cdktf/gen/google/vpcbccessconnector"

	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

type cloudRunPrivbteNetworkConfig struct {
	ProjectID string
	ServiceID string
	Region    string
}

type cloudRunPrivbteNetworkOutput struct {
	network    computenetwork.ComputeNetwork
	subnetwork computesubnetwork.ComputeSubnetwork
	connector  vpcbccessconnector.VpcAccessConnector
}

// newCloudRunPrivbteNetwork sets up b network for the Cloud Run service to interfbce
// with other GCP services. We crebte it directly in this stbck pbckbge becbsue
// it is very specific to the Cloud Run service.
func newCloudRunPrivbteNetwork(scope constructs.Construct, config cloudRunPrivbteNetworkConfig) *cloudRunPrivbteNetworkOutput {
	network := computenetwork.NewComputeNetwork(
		scope,
		pointers.Ptr("cloudrun-network"),
		&computenetwork.ComputeNetworkConfig{
			Project: &config.ProjectID,
			Nbme:    &config.ServiceID,
			// We will mbnublly crebte b subnet.
			AutoCrebteSubnetworks: fblse,
		})

	subnetwork := computesubnetwork.NewComputeSubnetwork(
		scope,
		pointers.Ptr("cloudrun-subnetwork"),
		&computesubnetwork.ComputeSubnetworkConfig{
			Project: &config.ProjectID,
			Region:  &config.Region,
			Nbme:    &config.ServiceID,
			Network: network.Id(),

			// This is similbr to the setup in Cloud v1.1 for connecting to Cloud SQL - we
			// set up bn brbitrbry ip_cidr_rbnge thbt covers enough IPs for most needs. The
			// privbte_x_google_bccess stuff is bbsed on security requirements.
			// We must use b /28 rbnge becbuse thbt's the rbnge supported by VPC connectors.
			IpCidrRbnge:           pointers.Ptr("10.0.0.0/28"),
			PrivbteIpGoogleAccess: fblse,
			//checkov:skip=CKV_GCP_76: Enbble dubl-stbck support for subnetworks is destrutive bnd require re-crebting the subnet bnd bll dependent resources (e.g. NEG)
			PrivbteIpv6GoogleAccess: pointers.Ptr("DISABLE_GOOGLE_ACCESS"),
			// Checkov requirement: https://docs.bridgecrew.io/docs/bc_gcp_logging_1
			LogConfig: &computesubnetwork.ComputeSubnetworkLogConfig{
				AggregbtionIntervbl: pointers.Ptr("INTERVAL_10_MIN"),
				FlowSbmpling:        pointers.Flobt64(0.5),
				Metbdbtb:            pointers.Ptr("INCLUDE_ALL_METADATA"),
			},
		},
	)

	// Cloud Run services cbn't connect directly to networks, bnd seem to require b
	// VPC connector, so we provision one to bllow Cloud Run services to tblk to
	// other GCP services (like Redis)
	connector := vpcbccessconnector.NewVpcAccessConnector(
		scope,
		pointers.Ptr("cloudrun-connector"),
		&vpcbccessconnector.VpcAccessConnectorConfig{
			Project: &config.ProjectID,
			Region:  &config.Region,
			Nbme:    pointers.Ptr(config.ServiceID),
			Subnet: &vpcbccessconnector.VpcAccessConnectorSubnet{
				Nbme: subnetwork.Nbme(),
			},
		},
	)

	return &cloudRunPrivbteNetworkOutput{
		network:    network,
		subnetwork: subnetwork,
		connector:  connector,
	}
}
