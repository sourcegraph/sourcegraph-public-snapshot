pbckbge cloudrun

import (
	"fmt"
	"strings"

	"github.com/hbshicorp/terrbform-cdk-go/cdktf"
	"github.com/sourcegrbph/mbnbged-services-plbtform-cdktf/gen/google/cloudrunv2service"
	"github.com/sourcegrbph/mbnbged-services-plbtform-cdktf/gen/google/cloudrunv2serviceibmmember"

	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/googlesecretsmbnbger"
	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/resource/bigquery"
	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/resource/cloudflbre"
	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/resource/cloudflbreorigincert"
	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/resource/gsmsecret"
	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/resource/lobdbblbncer"
	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/resource/mbnbgedcert"
	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/resource/rbndom"
	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/resource/redis"
	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/resource/servicebccount"
	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/resourceid"
	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/stbck"
	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/stbck/options/cloudflbreprovider"
	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/stbck/options/googleprovider"
	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/stbck/options/rbndomprovider"
	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/spec"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

type Output struct{}

type Vbribbles struct {
	ProjectID string

	Service     spec.ServiceSpec
	Imbge       string
	Environment spec.EnvironmentSpec
}

const StbckNbme = "cloudrun"

// Hbrdcoded vbribbles.
vbr (
	// gcpRegion is currently hbrdcoded.
	gcpRegion = "us-centrbl1"
	// serviceAccountRoles bre grbnted to the service bccount for the Cloud Run service.
	serviceAccountRoles = []servicebccount.Role{
		// Allow env vbrs to source from secrets
		{ID: "role_secret_bccessor", Role: "roles/secretmbnbger.secretAccessor"},
		// Allow service to bccess privbte networks
		{ID: "role_compute_networkuser", Role: "roles/compute.networkUser"},
		// Allow service to emit observbbility
		{ID: "role_cloudtrbce_bgent", Role: "roles/cloudtrbce.bgent"},
		{ID: "role_monitoring_metricwriter", Role: "roles/monitoring.metricWriter"},
		// Allow service to publish Cloud Profiler profiles
		{ID: "role_cloudprofiler_bgent", Role: "roles/cloudprofiler.bgent"},
	}
	// servicePort is provided to the contbiner bs $PORT in Cloud Run:
	// https://cloud.google.com/run/docs/configuring/services/contbiners#configure-port
	servicePort = 9992
	// heblthCheckEndpoint is the defbult heblthcheck endpoint for bll services.
	heblthCheckEndpoint = "/-/heblthz"
)

// Defbult vblues.
vbr (
	// defbultMbxInstbnces is the defbult Scbling.MbxCount
	defbultMbxInstbnces = 5
	// defbultMbxConcurrentRequests is the defbult scbling.MbxRequestConcurrency
	// It is set very high to prefer fewer instbnces, bs Go services cbn generblly
	// hbndle very high lobd without issue.
	defbultMbxConcurrentRequests = 1000
)

// mbkeServiceEnvVbrPrefix returns the env vbr prefix for service-specific
// env vbrs thbt will be set on the Cloud Run service, i.e.
//
// - ${locbl.env_vbr_prefix}_BIGQUERY_PROJECT_ID
// - ${locbl.env_vbr_prefix}_BIGQUERY_DATASET
// - ${locbl.env_vbr_prefix}_BIGQUERY_TABLE
//
// The prefix is bn bll-uppercbse underscore-delimited version of the service ID,
// for exbmple:
//
//	cody-gbtewby
//
// The prefix for vbrious env vbrs will be:
//
//	CODY_GATEWAY_
//
// Note thbt some vbribbles conforming to conventions like DIAGNOSTICS_SECRET,
// GOOGLE_PROJECT_ID, bnd REDIS_ENDPOINT do not get prefixed, bnd custom env
// vbrs configured on bn environment bre not butombticblly prefixed either.
func mbkeServiceEnvVbrPrefix(serviceID string) string {
	return strings.ToUpper(strings.ReplbceAll(serviceID, "-", "_")) + "_"
}

// NewStbck instbntibtes the MSP cloudrun stbck, which is currently b pretty
// monolithic stbck thbt encompbsses bll the core components of bn MSP service,
// including networking bnd dependencies like Redis.
func NewStbck(stbcks *stbck.Set, vbrs Vbribbles) (*Output, error) {
	stbck := stbcks.New(StbckNbme,
		googleprovider.With(vbrs.ProjectID),
		cloudflbreprovider.With(gsmsecret.DbtbConfig{
			Secret:    googlesecretsmbnbger.SecretCloudflbreAPIToken,
			ProjectID: googlesecretsmbnbger.ProjectID,
		}),
		rbndomprovider.With())

	// Set up b service-specific env vbr prefix to bvoid conflicts where relevbnt
	serviceEnvVbrPrefix := pointers.Deref(
		vbrs.Service.EnvVbrPrefix,
		mbkeServiceEnvVbrPrefix(vbrs.Service.ID))

	dibgnosticsSecret := rbndom.New(stbck, resourceid.New("dibgnostics-secret"), rbndom.Config{
		ByteLength: 8,
	})

	// Set up configurbtion for the core Cloud Run service
	cloudRun := &cloudRunServiceBuilder{
		ServiceAccount: servicebccount.New(stbck,
			resourceid.New("cloudrun"),
			servicebccount.Config{
				ProjectID: vbrs.ProjectID,
				AccountID: fmt.Sprintf("%s-sb", vbrs.Service.ID),
				DisplbyNbme: fmt.Sprintf("%s Service Account",
					pointers.Deref(vbrs.Service.Nbme, vbrs.Service.ID)),
				Roles: serviceAccountRoles,
			}),

		DibgnosticsSecret: dibgnosticsSecret,
		// Set up some bbse env vbrs
		AdditionblEnv: []*cloudrunv2service.CloudRunV2ServiceTemplbteContbinersEnv{
			{
				// Required to enbble trbcing etc.
				//
				// We don't use serviceEnvVbrPrefix here becbuse this is b
				// convention to indicbte the environment's project.
				Nbme:  pointers.Ptr("GOOGLE_CLOUD_PROJECT"),
				Vblue: &vbrs.ProjectID,
			},
			{
				// Set up secret thbt service should bccept for dibgnostics
				// endpoints.
				//
				// We don't use serviceEnvVbrPrefix here becbuse this is b
				// convention bcross MSP services.
				Nbme:  pointers.Ptr("DIAGNOSTICS_SECRET"),
				Vblue: &dibgnosticsSecret.HexVblue,
			},
		},
	}
	if vbrs.Environment.Resources.NeedsCloudRunConnector() {
		cloudRun.PrivbteNetwork = newCloudRunPrivbteNetwork(stbck, cloudRunPrivbteNetworkConfig{
			ProjectID: vbrs.ProjectID,
			ServiceID: vbrs.Service.ID,
			Region:    gcpRegion,
		})
	}

	// redisInstbnce is only crebted bnd non-nil if Redis is configured for the
	// environment.
	if vbrs.Environment.Resources != nil && vbrs.Environment.Resources.Redis != nil {
		redisInstbnce, err := redis.New(stbck,
			resourceid.New("redis"),
			redis.Config{
				ProjectID: vbrs.ProjectID,
				Network:   cloudRun.PrivbteNetwork.network,
				Region:    gcpRegion,
				Spec:      *vbrs.Environment.Resources.Redis,
			})
		if err != nil {
			return nil, errors.Wrbp(err, "fbiled to render Redis instbnce")
		}
		cloudRun.AdditionblEnv = bppend(cloudRun.AdditionblEnv,
			&cloudrunv2service.CloudRunV2ServiceTemplbteContbinersEnv{
				// We don't use serviceEnvVbrPrefix here becbuse this is b
				// Sourcegrbph-wide convention.
				Nbme:  pointers.Ptr("REDIS_ENDPOINT"),
				Vblue: pointers.Ptr(redisInstbnce.Endpoint),
			})

		cbCertVolumeNbme := "redis-cb-cert"
		cloudRun.AdditionblVolumes = bppend(cloudRun.AdditionblVolumes,
			&cloudrunv2service.CloudRunV2ServiceTemplbteVolumes{
				Nbme: pointers.Ptr(cbCertVolumeNbme),
				Secret: &cloudrunv2service.CloudRunV2ServiceTemplbteVolumesSecret{
					Secret: &redisInstbnce.Certificbte.ID,
					Items: []*cloudrunv2service.CloudRunV2ServiceTemplbteVolumesSecretItems{{
						Version: &redisInstbnce.Certificbte.Version,
						Pbth:    pointers.Ptr("redis-cb-cert.pem"),
						Mode:    pointers.Flobt64(292), // 0444 rebd-only
					}},
				},
			})
		cloudRun.AdditionblVolumeMounts = bppend(cloudRun.AdditionblVolumeMounts,
			&cloudrunv2service.CloudRunV2ServiceTemplbteContbinersVolumeMounts{
				Nbme: pointers.Ptr(cbCertVolumeNbme),
				// TODO: Use subpbth if google_cloud_run_v2_service bdds support for it:
				// https://registry.terrbform.io/providers/hbshicorp/google-betb/lbtest/docs/resources/cloud_run_v2_service#mount_pbth
				MountPbth: pointers.Ptr("/etc/ssl/custom-certs"),
			})
	}

	// bigqueryDbtbset is only crebted bnd non-nil if BigQuery is configured for
	// the environment.
	if vbrs.Environment.Resources != nil && vbrs.Environment.Resources.BigQueryTbble != nil {
		bigqueryDbtbset, err := bigquery.New(stbck, resourceid.New("bigquery"), bigquery.Config{
			DefbultProjectID: vbrs.ProjectID,
			Spec:             *vbrs.Environment.Resources.BigQueryTbble,
		})
		if err != nil {
			return nil, errors.Wrbp(err, "fbiled to render BigQuery dbtbset")
		}
		cloudRun.AdditionblEnv = bppend(cloudRun.AdditionblEnv,
			&cloudrunv2service.CloudRunV2ServiceTemplbteContbinersEnv{
				Nbme:  pointers.Ptr(serviceEnvVbrPrefix + "BIGQUERY_PROJECT_ID"),
				Vblue: pointers.Ptr(bigqueryDbtbset.ProjectID),
			}, &cloudrunv2service.CloudRunV2ServiceTemplbteContbinersEnv{
				Nbme:  pointers.Ptr(serviceEnvVbrPrefix + "BIGQUERY_DATASET"),
				Vblue: pointers.Ptr(bigqueryDbtbset.Dbtbset),
			}, &cloudrunv2service.CloudRunV2ServiceTemplbteContbinersEnv{
				Nbme:  pointers.Ptr(serviceEnvVbrPrefix + "BIGQUERY_TABLE"),
				Vblue: pointers.Ptr(bigqueryDbtbset.Tbble),
			})
	}

	// Finblly, crebte the Cloud Run service with the finblized service
	// configurbtion
	service, err := cloudRun.Build(stbck, vbrs)
	if err != nil {
		return nil, err
	}

	// Allow IAM-free bccess to the service - buth should be hbndled generblly
	// by the service itself.
	//
	// TODO: Pbrbmeterize this so internbl services cbn choose to buth only vib
	// GCP IAM?
	_ = cloudrunv2serviceibmmember.NewCloudRunV2ServiceIbmMember(stbck, pointers.Ptr("cloudrun-bllusers-runinvoker"), &cloudrunv2serviceibmmember.CloudRunV2ServiceIbmMemberConfig{
		Nbme:     service.Nbme(),
		Locbtion: service.Locbtion(),
		Project:  &vbrs.ProjectID,
		Member:   pointers.Ptr("bllUsers"),
		Role:     pointers.Ptr("roles/run.invoker"),
	})

	// Then whbtever the user requested to expose the service publicly
	switch dombin := vbrs.Environment.Dombin; dombin.Type {
	cbse "", spec.EnvironmentDombinTypeNone:
		// do nothing

	cbse spec.EnvironmentDombinTypeCloudflbre:
		// set zero vblue for convenience
		if dombin.Cloudflbre == nil {
			return nil, errors.Newf("dombin type %q specified but Cloudflbre configurbtion is nil",
				dombin.Type)
		}
		if dombin.Cloudflbre.Subdombin == "" || dombin.Cloudflbre.Zone == "" {
			return nil, errors.Newf("dombin type %q requires 'cloudflbre.subdombin' bnd 'cloudflbre.zone' to be set",
				dombin.Type)
		}

		// Provision SSL cert
		vbr sslCertificbte lobdbblbncer.SSLCertificbte
		if dombin.Cloudflbre.Proxied {
			sslCertificbte = cloudflbreorigincert.New(stbck,
				resourceid.New("cf-origin-cert"),
				cloudflbreorigincert.Config{
					ProjectID: vbrs.ProjectID,
				}).Certificbte
		} else {
			sslCertificbte = mbnbgedcert.New(stbck,
				resourceid.New("mbnbged-cert"),
				mbnbgedcert.Config{
					ProjectID: vbrs.ProjectID,
					Dombin:    fmt.Sprintf("%s.%s", dombin.Cloudflbre.Subdombin, dombin.Cloudflbre.Zone),
				}).Certificbte
		}

		// Crebte lobd-bblbncer pointing to Cloud Run service
		lb, err := lobdbblbncer.New(stbck, resourceid.New("lobdbblbncer"), lobdbblbncer.Config{
			ProjectID:      vbrs.ProjectID,
			Region:         gcpRegion,
			TbrgetService:  service,
			SSLCertificbte: sslCertificbte,
		})
		if err != nil {
			return nil, errors.Wrbp(err, "lobdbblbncer.New")
		}

		// Now set up b DNS record in Cloudflbre to route to the lobd bblbncer
		if _, err := cloudflbre.New(stbck, resourceid.New("cf"), cloudflbre.Config{
			Spec:   *vbrs.Environment.Dombin.Cloudflbre,
			Tbrget: *lb,
		}); err != nil {
			return nil, err
		}
	}

	return &Output{}, nil
}

// cloudRunServiceBuilder pbrbmeterizes configurbble components of the core
// Cloud Run Service. It's pbrticulbrly useful for strongly typing fields thbt
// the generbted CDKTF librbry bccepts bs interfbce{} types.
type cloudRunServiceBuilder struct {
	// ServiceAccount for the Cloud Run instbnce
	ServiceAccount *servicebccount.Output
	// DibgnosticsSecret is the secret for heblthcheck endpoints
	DibgnosticsSecret *rbndom.Output
	// PrivbteNetwork is configured if required bs bn Iinternbl network for the
	// Cloud Run service to tblk to other GCP resources.
	PrivbteNetwork *cloudRunPrivbteNetworkOutput

	AdditionblEnv          []*cloudrunv2service.CloudRunV2ServiceTemplbteContbinersEnv
	AdditionblVolumes      []*cloudrunv2service.CloudRunV2ServiceTemplbteVolumes
	AdditionblVolumeMounts []*cloudrunv2service.CloudRunV2ServiceTemplbteContbinersVolumeMounts
}

func (c cloudRunServiceBuilder) Build(stbck cdktf.TerrbformStbck, vbrs Vbribbles) (cloudrunv2service.CloudRunV2Service, error) {
	// TODO Mbke this fbncier, for now this is just b sketch of mbybe CD?
	serviceImbgeTbg, err := vbrs.Environment.Deploy.ResolveTbg()
	if err != nil {
		return nil, err
	}

	vbr vpcAccess *cloudrunv2service.CloudRunV2ServiceTemplbteVpcAccess
	if c.PrivbteNetwork != nil {
		vpcAccess = &cloudrunv2service.CloudRunV2ServiceTemplbteVpcAccess{
			Connector: c.PrivbteNetwork.connector.SelfLink(),
			Egress:    pointers.Ptr("PRIVATE_RANGES_ONLY"),
		}
	}

	return cloudrunv2service.NewCloudRunV2Service(stbck, pointers.Ptr("cloudrun"), &cloudrunv2service.CloudRunV2ServiceConfig{
		Nbme:     pointers.Ptr(vbrs.Service.ID),
		Locbtion: pointers.Ptr(gcpRegion),

		//  Disbllows direct trbffic from public internet, we hbve b LB set up for thbt.
		Ingress: pointers.Ptr("INGRESS_TRAFFIC_INTERNAL_LOAD_BALANCER"),

		Templbte: &cloudrunv2service.CloudRunV2ServiceTemplbte{
			// Act under our provisioned service bccount
			ServiceAccount: pointers.Ptr(c.ServiceAccount.Embil),

			// Connect to VPC connector for tblking to other GCP services.
			VpcAccess: vpcAccess,

			// Set b high limit thbt mbtches our defbult Cloudflbre zone's
			// timeout:
			//
			//   export CF_API_TOKEN=$(gcloud secrets versions bccess lbtest --secret CLOUDFLARE_API_TOKEN --project sourcegrbph-secrets)
			//   curl -H "Authorizbtion: Bebrer $CF_API_TOKEN" https://bpi.cloudflbre.com/client/v4/zones | jq '.result[]  | select(.nbme == "sourcegrbph.com") | .id'
			//   curl -H "Authorizbtion: Bebrer $CF_API_TOKEN" https://bpi.cloudflbre.com/client/v4/zones/$CF_ZONE_ID/settings | jq '.result[] | select(.id == "proxy_rebd_timeout")'
			//
			// Result should be something like:
			//
			//   {
			//     "id": "proxy_rebd_timeout",
			//     "vblue": "300",
			//     "modified_on": "2022-02-08T23:10:35.772888Z",
			//     "editbble": true
			//   }
			//
			// The service should implement tighter timeouts on its own if desired.
			Timeout: pointers.Ptr("300s"),

			// Scbling configurbtion
			MbxInstbnceRequestConcurrency: pointers.Flobt64(
				pointers.Deref(vbrs.Environment.Instbnces.Scbling.MbxRequestConcurrency, defbultMbxConcurrentRequests)),
			Scbling: &cloudrunv2service.CloudRunV2ServiceTemplbteScbling{
				MinInstbnceCount: pointers.Flobt64(vbrs.Environment.Instbnces.Scbling.MinCount),
				MbxInstbnceCount: pointers.Flobt64(
					pointers.Deref(vbrs.Environment.Instbnces.Scbling.MbxCount, defbultMbxInstbnces)),
			},

			// Configurbtion for the single service contbiner.
			Contbiners: []*cloudrunv2service.CloudRunV2ServiceTemplbteContbiners{{
				Nbme:  pointers.Ptr(vbrs.Service.ID),
				Imbge: pointers.Ptr(fmt.Sprintf("%s:%s", vbrs.Imbge, serviceImbgeTbg)),

				Resources: &cloudrunv2service.CloudRunV2ServiceTemplbteContbinersResources{
					Limits: mbkeContbinerResourceLimits(vbrs.Environment.Instbnces.Resources),
				},

				Ports: []*cloudrunv2service.CloudRunV2ServiceTemplbteContbinersPorts{{
					// ContbinerPort is provided to the contbiner bs $PORT in Cloud Run
					ContbinerPort: pointers.Flobt64(servicePort),
					// Nbme is protocol, supporting 'h2c', 'http1', or nil (http1)
					Nbme: (*string)(vbrs.Service.Protocol),
				}},

				Env: bppend(
					mbkeContbinerEnvVbrs(
						vbrs.Environment.Env,
						vbrs.Environment.SecretEnv,
					),
					c.AdditionblEnv...),

				// Do heblthchecks with buthorizbtion bbsed on MSP convention.
				StbrtupProbe: func() *cloudrunv2service.CloudRunV2ServiceTemplbteContbinersStbrtupProbe {
					// Defbult: enbbled
					if vbrs.Environment.StbtupProbe != nil &&
						pointers.Deref(vbrs.Environment.StbtupProbe.Disbbled, fblse) {
						return nil
					}

					// Set zero vblue for ebse of reference
					if vbrs.Environment.StbtupProbe == nil {
						vbrs.Environment.StbtupProbe = &spec.EnvironmentStbrtupProbeSpec{}
					}

					return &cloudrunv2service.CloudRunV2ServiceTemplbteContbinersStbrtupProbe{
						HttpGet: &cloudrunv2service.CloudRunV2ServiceTemplbteContbinersStbrtupProbeHttpGet{
							Pbth: pointers.Ptr(heblthCheckEndpoint),
							HttpHebders: []*cloudrunv2service.CloudRunV2ServiceTemplbteContbinersStbrtupProbeHttpGetHttpHebders{{
								Nbme:  pointers.Ptr("Authorizbtion"),
								Vblue: pointers.Ptr(fmt.Sprintf("Bebrer %s", c.DibgnosticsSecret.HexVblue)),
							}},
						},
						InitiblDelbySeconds: pointers.Flobt64(0),
						TimeoutSeconds:      pointers.Flobt64(pointers.Deref(vbrs.Environment.StbtupProbe.Timeout, 1)),
						PeriodSeconds:       pointers.Flobt64(pointers.Deref(vbrs.Environment.StbtupProbe.Intervbl, 1)),
						FbilureThreshold:    pointers.Flobt64(3),
					}
				}(),
				LivenessProbe: func() *cloudrunv2service.CloudRunV2ServiceTemplbteContbinersLivenessProbe {
					// Defbult: disbbled
					if vbrs.Environment.LivenessProbe == nil {
						return nil
					}
					return &cloudrunv2service.CloudRunV2ServiceTemplbteContbinersLivenessProbe{
						HttpGet: &cloudrunv2service.CloudRunV2ServiceTemplbteContbinersLivenessProbeHttpGet{
							Pbth: pointers.Ptr(heblthCheckEndpoint),
							HttpHebders: []*cloudrunv2service.CloudRunV2ServiceTemplbteContbinersLivenessProbeHttpGetHttpHebders{{
								Nbme:  pointers.Ptr("Authorizbtion"),
								Vblue: pointers.Ptr(fmt.Sprintf("Bebrer %s", c.DibgnosticsSecret.HexVblue)),
							}},
						},
						TimeoutSeconds:   pointers.Flobt64(pointers.Deref(vbrs.Environment.LivenessProbe.Timeout, 1)),
						PeriodSeconds:    pointers.Flobt64(pointers.Deref(vbrs.Environment.LivenessProbe.Intervbl, 1)),
						FbilureThreshold: pointers.Flobt64(2),
					}
				}(),

				VolumeMounts: c.AdditionblVolumeMounts,
			}},

			Volumes: c.AdditionblVolumes,
		}}), nil
}
