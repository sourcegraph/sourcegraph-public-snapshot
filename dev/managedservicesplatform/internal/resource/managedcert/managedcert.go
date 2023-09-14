package managedcert

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computemanagedsslcertificate"

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
	return &Output{
		Certificate: computemanagedsslcertificate.NewComputeManagedSslCertificate(scope,
			id.ResourceID("managed-cert"),
			&computemanagedsslcertificate.ComputeManagedSslCertificateConfig{
				Project: pointers.Ptr(config.ProjectID),
				Name:    pointers.Ptr(id.DisplayName()),
				Managed: &computemanagedsslcertificate.ComputeManagedSslCertificateManaged{
					Domains: &[]*string{
						pointers.Ptr(config.Domain),
					},
				},
			}),
	}
}
