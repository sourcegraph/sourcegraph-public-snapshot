package datastreamconnection

import (
	"fmt"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/hashicorp/terraform-cdk-go/cdktf"

	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computefirewall"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computeinstance"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/datastreamprivateconnection"

	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/datastreamconnectionprofile"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/cloudsql"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/postgresqllogicalreplication"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/privatenetwork"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/serviceaccount"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type Config struct {
	VPC      *privatenetwork.Output
	CloudSQL *cloudsql.Output
	// CloudSQLClientServiceAccount is used for establishing a proxy that can
	// connect to the Cloud SQL instance.
	CloudSQLClientServiceAccount serviceaccount.Output

	Publications          []postgresqllogicalreplication.PublicationOutput
	PublicationUserGrants []cdktf.ITerraformDependable
}

type Output struct {
}

// New provisions everything needed for Datastream to connect to Cloud SQL proxy:
//
//	Datastream --peering-> Private Network -> Cloud SQL Proxy VM -> Cloud SQL
//
// We need an additional VM proxying connections to Cloud SQL because Datastream
// and Cloud SQL both have their own internal VPCs, and we cannot transitively
// peer them over the private network we manage.
func New(scope constructs.Construct, id resourceid.ID, config Config) (*Output, error) {
	const proxyInstanceName = "msp-datastream-cloudsqlproxy"

	cloudsqlproxyInstance := computeinstance.NewComputeInstance(scope, id.TerraformID("cloudsqlproxy"), &computeinstance.ComputeInstanceConfig{
		Name:        pointers.Ptr(proxyInstanceName),
		Description: pointers.Ptr("Cloud SQL proxy to allow Datastream to connect to Cloud SQL over private network"),

		// Just use a random zone in the same region as the Cloud SQL instance
		Zone: pointers.Stringf("%s-a", *config.CloudSQL.Instance.Region()),

		MachineType: pointers.Ptr("e2-micro"),
		NetworkInterface: []computeinstance.ComputeInstanceNetworkInterface{{
			Network:    config.VPC.Network.Name(),
			Subnetwork: config.VPC.Subnetwork.Name(),
		}},
		ServiceAccount: &computeinstance.ComputeInstanceServiceAccount{
			Email:  &config.CloudSQLClientServiceAccount.Email,
			Scopes: &[]*string{pointers.Ptr("https://www.googleapis.com/auth/cloud-platform")},
		},
		BootDisk: &computeinstance.ComputeInstanceBootDisk{
			InitializeParams: &computeinstance.ComputeInstanceBootDiskInitializeParams{
				Image: pointers.Ptr("cos-cloud/cos-stable"),
				Size:  pointers.Float64(10), // Gb
			},
		},
		Tags: &[]*string{pointers.Ptr(proxyInstanceName)},

		// See docstring of newMetadataGCEContainerDeclaration for details about
		// the label and metadata.
		Labels: &map[string]*string{
			"container-vm": pointers.Ptr(proxyInstanceName),
			"msp":          pointers.Ptr("true"),
		},
		Metadata: &map[string]*string{
			"gce-container-declaration": pointers.Ptr(
				newMetadataGCEContainerDeclaration(proxyInstanceName, *config.CloudSQL.Instance.ConnectionName())),
		},
	})

	const dsPrivateConnectionSubnet = "10.126.0.0/29" // any '/29' range
	datastreamConnection := datastreamprivateconnection.NewDatastreamPrivateConnection(scope, id.TerraformID("cloudsqlproxy-privateconnection"), &datastreamprivateconnection.DatastreamPrivateConnectionConfig{
		DisplayName:         pointers.Ptr(proxyInstanceName),
		PrivateConnectionId: pointers.Ptr(proxyInstanceName),
		Location:            config.CloudSQL.Instance.Region(),
		VpcPeeringConfig: &datastreamprivateconnection.DatastreamPrivateConnectionVpcPeeringConfig{
			Vpc:    config.VPC.Network.Id(),
			Subnet: pointers.Ptr(dsPrivateConnectionSubnet),
		},
		Labels: &map[string]*string{"msp": pointers.Ptr("true")},
	})

	// Allow ingress from Datastream
	datastreamIngressFirewall := computefirewall.NewComputeFirewall(scope, id.TerraformID("cloudsqlproxy-firewall-datastream-ingress"), &computefirewall.ComputeFirewallConfig{
		Name:        pointers.Stringf("%s-datastream-ingress", proxyInstanceName),
		Description: pointers.Ptr("Allow incoming connections from a Datastream private connection to the Cloud SQL Proxy VM"),
		Network:     config.VPC.Network.Name(),
		Priority:    pointers.Float64(1000),

		Direction: pointers.Ptr("INGRESS"),
		SourceRanges: &[]*string{
			pointers.Ptr(dsPrivateConnectionSubnet),
		},
		Allow: []computefirewall.ComputeFirewallAllow{{
			Protocol: pointers.Ptr("tcp"),
			Ports:    &[]*string{pointers.Ptr("5432")},
		}},
		TargetTags: cloudsqlproxyInstance.Tags(),
	})

	// Allow IAP ingress for debug https://cloud.google.com/iap/docs/using-tcp-forwarding
	_ = computefirewall.NewComputeFirewall(scope, id.TerraformID("cloudsqlproxy-firewall-iap-ingress"), &computefirewall.ComputeFirewallConfig{
		Name:        pointers.Stringf("%s-iap-ingress", proxyInstanceName),
		Description: pointers.Ptr("Allow incoming connections from GCP IAP to the Cloud SQL Proxy VM"),
		Network:     config.VPC.Network.Name(),
		Priority:    pointers.Float64(1000),

		Direction: pointers.Ptr("INGRESS"),
		SourceRanges: &[]*string{
			pointers.Ptr("35.235.240.0/20"),
		},
		Allow: []computefirewall.ComputeFirewallAllow{{
			Protocol: pointers.Ptr("tcp"),
			Ports:    &[]*string{pointers.Ptr("22")},
		}},
		TargetTags: cloudsqlproxyInstance.Tags(),
	})

	for _, pub := range config.Publications {
		id := id.Group(pub.Name)

		// The Datastream Connection Profile is what the data team will click-ops
		// during their creation of the actual Datastream "Stream".
		// https://cloud.google.com/datastream/docs/create-a-stream
		//
		// This is where we stop managing things directly in MSP.
		_ = datastreamconnectionprofile.NewDatastreamConnectionProfile(scope, id.TerraformID("cloudsqlproxy-connectionprofile"), &datastreamconnectionprofile.DatastreamConnectionProfileConfig{
			DisplayName:         pointers.Stringf("MSP Publication - %s", pub.Name),
			ConnectionProfileId: pointers.Stringf("msp-publication-%s", pub.Name),
			Labels: &map[string]*string{
				"msp":                 pointers.Ptr("true"),
				"database":            pointers.Ptr(pub.Database),
				"pg_replication_slot": pub.ReplicationSlotName,
				"pg_publication":      pub.PublicationName,
			},
			Location: config.CloudSQL.Instance.Region(),
			PostgresqlProfile: &datastreamconnectionprofile.DatastreamConnectionProfilePostgresqlProfile{
				Hostname: cloudsqlproxyInstance.NetworkInterface().
					Get(pointers.Float64(0)).
					NetworkIp(), // internal IP of the instance
				Port: pointers.Float64(5432),

				Database: pointers.Ptr(pub.Database),
				Username: pub.User.Name(),
				Password: pub.User.Password(),
			},
			PrivateConnectivity: &datastreamconnectionprofile.DatastreamConnectionProfilePrivateConnectivity{
				PrivateConnection: datastreamConnection.Name(),
			},
			DependsOn: pointers.Ptr(append(config.PublicationUserGrants,
				datastreamIngressFirewall)),
		})
	}

	return &Output{}, nil
}

// newMetadataGCEContainerDeclaration recreates the metadata value that GCP
// provides when you click-ops a Compute Engine VM that runs a container. GCP
// manages the container lifecycle which is quite nice. Sadly this isn't
// available via an official Terraform API, but we can replicate that GCP does
// and hope they don't change anything.
func newMetadataGCEContainerDeclaration(containerName, cloudSQLConnectionString string) string {
	// Note the docstring about how this format is not a public API - it's
	// generated by GCP, and we include that as well
	return fmt.Sprintf(`
spec:
  restartPolicy: Always
  containers:
  - name: %s
    image: gcr.io/cloud-sql-connectors/cloud-sql-proxy
    args:
    - '--auto-iam-authn'
    - '--private-ip'
    - '--address=0.0.0.0'
    - '%s'

# This container declaration format is not public API and may change without notice. Please
# use gcloud command-line tool or Google Cloud Console to run Containers on Google Compute Engine.`,
		containerName, cloudSQLConnectionString)
}
