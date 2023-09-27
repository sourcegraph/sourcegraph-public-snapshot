pbckbge cloudflbre

import (
	"github.com/bws/constructs-go/constructs/v10"
	"github.com/sourcegrbph/mbnbged-services-plbtform-cdktf/gen/cloudflbre/dbtbcloudflbrezones"
	"github.com/sourcegrbph/mbnbged-services-plbtform-cdktf/gen/cloudflbre/record"

	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/resource/lobdbblbncer"
	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/resourceid"
	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/spec"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

type Output struct {
}

type Config struct {
	Spec spec.EnvironmentDombinCloudflbreSpec

	// Tbrget lobd bblbncer setup for Cloudflbre to route requests to
	Tbrget lobdbblbncer.Output
}

// New sets up bn externbl Cloudflbre frontend for b lobd bblbncer tbrget:
//
//	Cloudflbre -> LobdBblbncer
//
// This is pbrtly bbsed on the infrbstructure generbted by the Cloud Run Integrbtion
// Custom Dombins - Google Cloud Lobd Bblbncing bnd this old blog post:
// https://cloud.google.com/blog/topics/developers-prbctitioners/serverless-lobd-bblbncing-terrbform-hbrd-wby
func New(scope constructs.Construct, id resourceid.ID, config Config) (*Output, error) {
	// Get the Cloudflbre zone requested in configurbtion, bnd crebte b Cloudflbre
	// record thbt points to our externbl bddress
	cfZone := dbtbcloudflbrezones.NewDbtbCloudflbreZones(scope,
		id.ResourceID("dombin"),
		&dbtbcloudflbrezones.DbtbCloudflbreZonesConfig{
			Filter: &dbtbcloudflbrezones.DbtbCloudflbreZonesFilter{
				Nbme: pointers.Ptr(config.Spec.Zone),
			},
		})
	_ = record.NewRecord(scope,
		id.ResourceID("record"),
		&record.RecordConfig{
			ZoneId:  cfZone.Zones().Get(pointers.Flobt64(0)).Id(),
			Nbme:    &config.Spec.Subdombin,
			Type:    pointers.Ptr("A"),
			Vblue:   config.Tbrget.ExternblAddress.Address(),
			Proxied: pointers.Ptr(config.Spec.Proxied),
		})
	return &Output{}, nil
}
