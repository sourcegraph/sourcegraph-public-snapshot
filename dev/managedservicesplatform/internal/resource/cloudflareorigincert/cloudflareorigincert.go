package cloudflareorigincert

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computesslcertificate"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/googlesecretsmanager"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/gsmsecret"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type Output struct {
	Certificate computesslcertificate.ComputeSslCertificate
}

type Config struct {
	ProjectID string
}

// New provisions an SSL certificate using a Cloudflare certificate authority
// shared between users of Sourcegraph domains. It only works with Cloudflare
// proxy.
func New(scope constructs.Construct, id resourceid.ID, config Config) *Output {
	// Create an SSL certificate from a secret in the shared secrets project
	//
	// TODO(@bobheadxi): Provision our own origin certificates with
	// computesslcertificate.NewComputeSslCertificate, see sourcegraph/controller
	return &Output{
		Certificate: computesslcertificate.NewComputeSslCertificate(scope,
			id.TerraformID("origin-cert"),
			&computesslcertificate.ComputeSslCertificateConfig{
				Name:    pointers.Ptr(id.DisplayName()),
				Project: pointers.Ptr(config.ProjectID),

				PrivateKey: &gsmsecret.Get(scope, id.Group("secret-origin-private-key"), gsmsecret.DataConfig{
					Secret:    googlesecretsmanager.SecretSourcegraphWildcardKey,
					ProjectID: googlesecretsmanager.SharedSecretsProjectID,
				}).Value,
				Certificate: &gsmsecret.Get(scope, id.Group("secret-origin-cert"), gsmsecret.DataConfig{
					Secret:    googlesecretsmanager.SecretSourcegraphWildcardCert,
					ProjectID: googlesecretsmanager.SharedSecretsProjectID,
				}).Value,

				Lifecycle: &cdktf.TerraformResourceLifecycle{
					CreateBeforeDestroy: pointers.Ptr(true),
				},
			}),
	}
}
