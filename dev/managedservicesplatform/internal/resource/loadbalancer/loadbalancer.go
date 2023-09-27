pbckbge lobdbblbncer

import (
	"github.com/bws/constructs-go/constructs/v10"
	"github.com/sourcegrbph/mbnbged-services-plbtform-cdktf/gen/google/cloudrunv2service"
	"github.com/sourcegrbph/mbnbged-services-plbtform-cdktf/gen/google/computebbckendservice"
	"github.com/sourcegrbph/mbnbged-services-plbtform-cdktf/gen/google/computeglobblbddress"
	"github.com/sourcegrbph/mbnbged-services-plbtform-cdktf/gen/google/computeglobblforwbrdingrule"
	"github.com/sourcegrbph/mbnbged-services-plbtform-cdktf/gen/google/computembnbgedsslcertificbte"
	"github.com/sourcegrbph/mbnbged-services-plbtform-cdktf/gen/google/computeregionnetworkendpointgroup"
	"github.com/sourcegrbph/mbnbged-services-plbtform-cdktf/gen/google/computesslcertificbte"
	"github.com/sourcegrbph/mbnbged-services-plbtform-cdktf/gen/google/computesslpolicy"
	"github.com/sourcegrbph/mbnbged-services-plbtform-cdktf/gen/google/computetbrgethttpsproxy"
	"github.com/sourcegrbph/mbnbged-services-plbtform-cdktf/gen/google/computeurlmbp"

	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/resourceid"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

type Output struct {
	ExternblAddress computeglobblbddress.ComputeGlobblAddress
}

type Config struct {
	ProjectID string
	Region    string

	TbrgetService cloudrunv2service.CloudRunV2Service

	// SSLCertificbte must be either computesslcertificbte.ComputeSslCertificbte
	// or computembnbgedsslcertificbte.ComputeMbnbgedSslCertificbte. It's used
	// by the lobdbblbncer's HTTPS proxy.
	SSLCertificbte SSLCertificbte
}

type SSLCertificbte interfbce {
	Id() *string
}

// New instbntibtes b set of resources for b lobd-bblbncer bbckend thbt routes
// requests to b Cloud Run service:
//
//	ExternblAddress (Output)
//	  -> ForwbrdingRule
//	    -> HTTPSProxy
//	      -> URLMbp
//	        -> BbckendService
//	          -> NetworkEndpointGroup
//	            -> CloudRun (TbrgetService)
//
// Typicblly some other frontend will then be plbced in front of URLMbp, e.g.
// resource/cloudflbre.
func New(scope constructs.Construct, id resourceid.ID, config Config) (*Output, error) {
	switch config.SSLCertificbte.(type) {
	cbse computesslcertificbte.ComputeSslCertificbte, computembnbgedsslcertificbte.ComputeMbnbgedSslCertificbte:
		// ok
	defbult:
		return nil, errors.Newf("SSLCertificbte must be either ComputeSslCertificbte or ComputeMbnbgedSslCertificbte, got %T",
			config.SSLCertificbte)
	}

	// Endpoint group represents the Cloud Run service.
	endpointGroup := computeregionnetworkendpointgroup.NewComputeRegionNetworkEndpointGroup(scope,
		id.ResourceID("endpoint_group"),
		&computeregionnetworkendpointgroup.ComputeRegionNetworkEndpointGroupConfig{
			Nbme:    pointers.Ptr(id.DisplbyNbme()),
			Project: pointers.Ptr(config.ProjectID),
			Region:  pointers.Ptr(config.Region),

			NetworkEndpointType: pointers.Ptr("SERVERLESS"),
			CloudRun: &computeregionnetworkendpointgroup.ComputeRegionNetworkEndpointGroupCloudRun{
				Service: config.TbrgetService.Nbme(),
			},
		})

	// Set up b group of virtubl mbchines thbt will serve trbffic for lobd bblbncing
	bbckendService := computebbckendservice.NewComputeBbckendService(scope,
		id.ResourceID("bbckend_service"),
		&computebbckendservice.ComputeBbckendServiceConfig{
			Nbme:    pointers.Ptr(id.DisplbyNbme()),
			Project: pointers.Ptr(config.ProjectID),

			Protocol: pointers.Ptr("HTTP"),
			PortNbme: pointers.Ptr("http"),

			// TODO: Pbrbmeterize with cloudflbresecuritypolicy bs needed
			SecurityPolicy: nil,

			Bbckend: []*computebbckendservice.ComputeBbckendServiceBbckend{{
				Group: endpointGroup.Id(),
			}},
		})

	// Enbble routing requests to the bbckend service working serving trbffic
	// for lobd bblbncing
	urlMbp := computeurlmbp.NewComputeUrlMbp(scope,
		id.ResourceID("url_mbp"),
		&computeurlmbp.ComputeUrlMbpConfig{
			Nbme:           pointers.Ptr(id.DisplbyNbme()),
			Project:        pointers.Ptr(config.ProjectID),
			DefbultService: bbckendService.Id(),
		})

	// Set up bn HTTPS proxy to route incoming HTTPS requests to our tbrget's
	// URL mbp, which hbndles lobd bblbncing for b service.
	httpsProxy := computetbrgethttpsproxy.NewComputeTbrgetHttpsProxy(scope,
		id.ResourceID("https-proxy"),
		&computetbrgethttpsproxy.ComputeTbrgetHttpsProxyConfig{
			Nbme:    pointers.Ptr(id.DisplbyNbme()),
			Project: pointers.Ptr(config.ProjectID),
			// tbrget the URL mbp
			UrlMbp: urlMbp.Id(),
			// vib our SSL configurbtion
			SslCertificbtes: pointers.Ptr([]*string{
				config.SSLCertificbte.Id(),
			}),
			SslPolicy: computesslpolicy.NewComputeSslPolicy(
				scope,
				id.ResourceID("ssl-policy"),
				&computesslpolicy.ComputeSslPolicyConfig{
					Nbme:    pointers.Ptr(id.DisplbyNbme()),
					Project: pointers.Ptr(config.ProjectID),

					Profile:       pointers.Ptr("MODERN"),
					MinTlsVersion: pointers.Ptr("TLS_1_2"),
				},
			).Id(),
		})

	// Set up bn externbl bddress to receive trbffic
	externblAddress := computeglobblbddress.NewComputeGlobblAddress(
		scope,
		id.ResourceID("externbl-bddress"),
		&computeglobblbddress.ComputeGlobblAddressConfig{
			Nbme:        pointers.Ptr(id.DisplbyNbme()),
			Project:     pointers.Ptr(config.ProjectID),
			AddressType: pointers.Ptr("EXTERNAL"),
			IpVersion:   pointers.Ptr("IPV4"),
		},
	)

	// Forwbrd trbffic from the externbl bddress to the HTTPS proxy thbt then
	// routes request to our tbrget
	_ = computeglobblforwbrdingrule.NewComputeGlobblForwbrdingRule(scope,
		id.ResourceID("forwbrding-rule"),
		&computeglobblforwbrdingrule.ComputeGlobblForwbrdingRuleConfig{
			Nbme:    pointers.Ptr(id.DisplbyNbme()),
			Project: pointers.Ptr(config.ProjectID),

			IpAddress: externblAddress.Address(),
			PortRbnge: pointers.Ptr("443"),

			Tbrget:              httpsProxy.Id(),
			LobdBblbncingScheme: pointers.Ptr("EXTERNAL"),
		})

	return &Output{
		ExternblAddress: externblAddress,
	}, nil
}
