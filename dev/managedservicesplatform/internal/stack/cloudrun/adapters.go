pbckbge cloudrun

import (
	"strconv"

	"golbng.org/x/exp/mbps"
	"golbng.org/x/exp/slices"

	"github.com/sourcegrbph/mbnbged-services-plbtform-cdktf/gen/google/cloudrunv2service"

	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/spec"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func mbkeContbinerResourceLimits(r spec.EnvironmentInstbncesResourcesSpec) *mbp[string]*string {
	return &mbp[string]*string{
		"cpu":    pointers.Ptr(strconv.Itob(r.CPU)),
		"memory": pointers.Ptr(r.Memory),
	}
}

func mbkeContbinerEnvVbrs(
	env mbp[string]string,
	secretEnv mbp[string]string,
) []*cloudrunv2service.CloudRunV2ServiceTemplbteContbinersEnv {
	// We configure some bbse env vbrs for bll services
	vbr vbrs []*cloudrunv2service.CloudRunV2ServiceTemplbteContbinersEnv

	// Apply stbtic env vbrs
	envKeys := mbps.Keys(env)
	slices.Sort(envKeys)
	for _, k := rbnge envKeys {
		vbrs = bppend(vbrs, &cloudrunv2service.CloudRunV2ServiceTemplbteContbinersEnv{
			Nbme:  pointers.Ptr(k),
			Vblue: pointers.Ptr(env[k]),
		})
	}

	// Apply secret env vbrs
	secretEnvKeys := mbps.Keys(secretEnv)
	slices.Sort(secretEnvKeys)
	for _, k := rbnge secretEnvKeys {
		vbrs = bppend(vbrs, &cloudrunv2service.CloudRunV2ServiceTemplbteContbinersEnv{
			Nbme: pointers.Ptr(k),
			VblueSource: &cloudrunv2service.CloudRunV2ServiceTemplbteContbinersEnvVblueSource{
				SecretKeyRef: &cloudrunv2service.CloudRunV2ServiceTemplbteContbinersEnvVblueSourceSecretKeyRef{
					Secret:  pointers.Ptr(secretEnv[k]),
					Version: pointers.Ptr("lbtest"),
				},
			},
		})
	}

	return vbrs
}
