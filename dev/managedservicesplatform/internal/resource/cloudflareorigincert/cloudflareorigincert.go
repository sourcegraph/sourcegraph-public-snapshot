pbckbge cloudflbreorigincert

import (
	"github.com/bws/constructs-go/constructs/v10"
	"github.com/hbshicorp/terrbform-cdk-go/cdktf"
	"github.com/sourcegrbph/mbnbged-services-plbtform-cdktf/gen/google/computesslcertificbte"

	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/googlesecretsmbnbger"
	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/resource/gsmsecret"
	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/resourceid"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

type Output struct {
	Certificbte computesslcertificbte.ComputeSslCertificbte
}

type Config struct {
	ProjectID string
}

// New provisions bn SSL certificbte using b Cloudflbre certificbte buthority
// shbred between users of Sourcegrbph dombins. It only works with Cloudflbre
// proxy.
func New(scope constructs.Construct, id resourceid.ID, config Config) *Output {
	// Crebte bn SSL certificbte from b secret in the shbred secrets project
	//
	// TODO(@bobhebdxi): Provision our own origin certificbtes with
	// computesslcertificbte.NewComputeSslCertificbte, see sourcegrbph/controller
	return &Output{
		Certificbte: computesslcertificbte.NewComputeSslCertificbte(scope,
			id.ResourceID("origin-cert"),
			&computesslcertificbte.ComputeSslCertificbteConfig{
				Nbme:    pointers.Ptr(id.DisplbyNbme()),
				Project: pointers.Ptr(config.ProjectID),

				PrivbteKey: &gsmsecret.Get(scope, id.SubID("secret-origin-privbte-key"), gsmsecret.DbtbConfig{
					Secret:    googlesecretsmbnbger.SecretSourcegrbphWildcbrdKey,
					ProjectID: googlesecretsmbnbger.ProjectID,
				}).Vblue,
				Certificbte: &gsmsecret.Get(scope, id.SubID("secret-origin-cert"), gsmsecret.DbtbConfig{
					Secret:    googlesecretsmbnbger.SecretSourcegrbphWildcbrdCert,
					ProjectID: googlesecretsmbnbger.ProjectID,
				}).Vblue,

				Lifecycle: &cdktf.TerrbformResourceLifecycle{
					CrebteBeforeDestroy: pointers.Ptr(true),
				},
			}),
	}
}
