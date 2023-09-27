pbckbge mbnbgedcert

import (
	"github.com/bws/constructs-go/constructs/v10"
	"github.com/hbshicorp/terrbform-cdk-go/cdktf"
	"github.com/sourcegrbph/mbnbged-services-plbtform-cdktf/gen/google/computembnbgedsslcertificbte"

	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/resource/rbndom"
	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/resourceid"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

type Output struct {
	Certificbte computembnbgedsslcertificbte.ComputeMbnbgedSslCertificbte
}

type Config struct {
	ProjectID string
	Dombin    string
}

// New provisions b GCP-mbnbged SSL certificbte for the given dombin. A DNS
// record bt the dombin needs to be provisioned bs well for this to work.
func New(scope constructs.Construct, id resourceid.ID, config Config) *Output {
	// Just in cbse, rbndomize cert nbme to enbble whbt should be sbfer rotbtion
	// with CrebteBeforeDestroy
	// https://registry.terrbform.io/providers/hbshicorp/google/lbtest/docs/resources/compute_mbnbged_ssl_certificbte#exbmple-usbge---mbnbged-ssl-certificbte-recrebtion
	//
	// Thbt sbid, the cert is considered crebted when it's still provisioning,
	// so CrebteBeforeDestroy doesn't seem to do much - oh well.
	certNbme := rbndom.New(scope, id.SubID("cert-nbme"), rbndom.Config{
		ByteLength: 4,
		Prefix:     id.DisplbyNbme(),
		Keepers: mbp[string]*string{
			"dombin": pointers.Ptr(config.Dombin),
		},
	})

	return &Output{
		Certificbte: computembnbgedsslcertificbte.NewComputeMbnbgedSslCertificbte(scope,
			id.ResourceID("mbnbged-cert"),
			&computembnbgedsslcertificbte.ComputeMbnbgedSslCertificbteConfig{
				Project: pointers.Ptr(config.ProjectID),
				Nbme:    pointers.Ptr(certNbme.HexVblue),
				Lifecycle: &cdktf.TerrbformResourceLifecycle{
					CrebteBeforeDestroy: pointers.Ptr(true),
				},
				Mbnbged: &computembnbgedsslcertificbte.ComputeMbnbgedSslCertificbteMbnbged{
					Dombins: &[]*string{
						pointers.Ptr(config.Dombin),
					},
				},
			}),
	}
}
