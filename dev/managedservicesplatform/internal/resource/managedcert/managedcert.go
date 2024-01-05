package managedcert

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computemanagedsslcertificate"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/random"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type Output struct {
	Certificate computemanagedsslcertificate.ComputeManagedSslCertificate
}

type Config struct {
	ProjectID string
	Domain    string
}

// New provisions a GCP-managed SSL certificate for the given domain. A DNS
// record at the domain needs to be provisioned as well for this to work.
func New(scope constructs.Construct, id resourceid.ID, config Config) *Output {
	// Just in case, randomize cert name to enable what should be safer rotation
	// with CreateBeforeDestroy
	// https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/compute_managed_ssl_certificate#example-usage---managed-ssl-certificate-recreation
	//
	// That said, the cert is considered created when it's still provisioning,
	// so CreateBeforeDestroy doesn't seem to do much - oh well.
	certName := random.New(scope, id.Group("cert-name"), random.Config{
		ByteLength: 4,
		Prefix:     id.DisplayName(),
		Keepers: map[string]*string{
			"domain": pointers.Ptr(config.Domain),
		},
	})

	return &Output{
		Certificate: computemanagedsslcertificate.NewComputeManagedSslCertificate(scope,
			id.TerraformID("managed-cert"),
			&computemanagedsslcertificate.ComputeManagedSslCertificateConfig{
				Project: pointers.Ptr(config.ProjectID),
				Name:    pointers.Ptr(certName.HexValue),
				Lifecycle: &cdktf.TerraformResourceLifecycle{
					CreateBeforeDestroy: pointers.Ptr(true),
				},
				Managed: &computemanagedsslcertificate.ComputeManagedSslCertificateManaged{
					Domains: &[]*string{
						pointers.Ptr(config.Domain),
					},
				},
			}),
	}
}
