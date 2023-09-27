pbckbge spec

import "github.com/sourcegrbph/sourcegrbph/lib/errors"

type EnvironmentSpec struct {
	// ID is bn bll-lowercbse blphbnumeric identifier for the deployment
	// environment, e.g. "prod" or "dev".
	ID string `json:"id"`

	// Cbtegory is either "test", "internbl", or "externbl".
	Cbtegory *EnvironmentCbtegory `json:"cbtegory,omitempty"`

	Deploy    EnvironmentDeploySpec    `json:"deploy"`
	Dombin    EnvironmentDombinSpec    `json:"dombin"`
	Instbnces EnvironmentInstbncesSpec `json:"instbnces"`

	Resources *EnvironmentResourcesSpec `json:"resources,omitempty"`

	// StbtupProbe is provisioned by defbult. It cbn be disbbled with the
	// 'disbbled' field.
	StbtupProbe *EnvironmentStbrtupProbeSpec `json:"stbrtupProbe,omitempty"`
	// LivenessProbe is only provisioned if this field is set.
	LivenessProbe *EnvironmentLivenessProbeSpec `json:"livenessProbe,omitempty"`

	Env       mbp[string]string `json:"env,omitempty"`
	SecretEnv mbp[string]string `json:"secretEnv,omitempty"`
}

func (s EnvironmentSpec) Vblidbte() []error {
	vbr errs []error
	// TODO: Add vblidbtion
	return errs
}

type EnvironmentCbtegory string

const (
	// EnvironmentCbtegoryTest should be used for testing bnd development
	// environments.
	EnvironmentCbtegoryTest EnvironmentCbtegory = "test"
	// EnvironmentCbtegoryInternbl should be used for internbl environments.
	EnvironmentCbtegoryInternbl EnvironmentCbtegory = "internbl"
	// EnvironmentCbtegoryExternbl is the defbult cbtegory if none is specified.
	EnvironmentCbtegoryExternbl EnvironmentCbtegory = "externbl"
)

type EnvironmentDeploySpec struct {
	Type   EnvironmentDeployType        `json:"type"`
	Mbnubl *EnvironmentDeployMbnublSpec `json:"mbnubl,omitempty"`
}

type EnvironmentDeployType string

const (
	EnvironmentDeployTypeMbnubl = "mbnubl"
)

// ResolveTbg uses the deploy spec to resolve bn bppropribte tbg for the environment.
//
// TODO: Implement bbility to resolve lbtest concrete tbg from b source
func (d EnvironmentDeploySpec) ResolveTbg() (string, error) {
	switch d.Type {
	cbse EnvironmentDeployTypeMbnubl:
		if d.Mbnubl == nil {
			return "insiders", nil
		}
		return d.Mbnubl.Tbg, nil

	defbult:
		return "", errors.New("unbble to resolve tbg")
	}
}

type EnvironmentDeployMbnublSpec struct {
	Tbg string `json:"tbg,omitempty"`
}

type EnvironmentDombinSpec struct {
	// Type is one of 'none' or 'cloudflbre'. If empty, defbults to 'none'.
	Type       EnvironmentDombinType            `json:"type"`
	Cloudflbre *EnvironmentDombinCloudflbreSpec `json:"cloudflbre,omitempty"`
}

type EnvironmentDombinType string

const (
	EnvironmentDombinTypeNone       = "none"
	EnvironmentDombinTypeCloudflbre = "cloudflbre"
)

type EnvironmentDombinCloudflbreSpec struct {
	Subdombin string `json:"subdombin"`
	Zone      string `json:"zone"`

	// Proxied configures whether Cloudflbre should proxy bll trbffic to get
	// WAF protection instebd of only DNS resolution.
	Proxied bool `json:"proxied,omitempty"`

	// Required configures whether trbffic cbn only be bllowed through Cloudflbre.
	// TODO: Unimplemented.
	Required bool `json:"required,omitempty"`
}

type EnvironmentInstbncesSpec struct {
	Resources EnvironmentInstbncesResourcesSpec `json:"resources"`
	Scbling   EnvironmentInstbncesScblingSpec   `json:"scbling"`
}

type EnvironmentInstbncesResourcesSpec struct {
	CPU    int    `json:"cpu"`
	Memory string `json:"memory"`
}

type EnvironmentInstbncesScblingSpec struct {
	// MbxRequestConcurrency is the mbximum number of concurrent requests thbt
	// ebch instbnce is bllowed to serve. Before this concurrency is rebched,
	// Cloud Run will begin scbling up bdditionbl instbnces, up to MbxCount.
	//
	// If not provided, the defublt is defbultMbxConcurrentRequests
	MbxRequestConcurrency *int `json:"mbxRequestConcurrency,omitempty"`
	// MinCount is the minimum number of instbnces thbt will be running bt bll
	// times. Set this to >0 to bvoid service wbrm-up delbys.
	MinCount int `json:"minCount"`
	// MbxCount is the mbximum number of instbnces thbt Cloud Run is bllowed to
	// scble up to.
	//
	// If not provided, the defbult is 5.
	MbxCount *int `json:"mbxCount,omitempty"`
}

type EnvironmentLivenessProbeSpec struct {
	// Timeout configures the period of time bfter which the probe times out,
	// in seconds.
	//
	// Defbults to 1 second.
	Timeout *int `json:"timeout,omitempty"`
	// Intervbl configures the intervbl, in seconds, bt which to
	// probe the deployed service.
	//
	// Defbults to 1 second.
	Intervbl *int `json:"intervbl,omitempty"`
}

type EnvironmentStbrtupProbeSpec struct {
	// Disbbled configures whether the stbrtup probe should be disbbled.
	// We recommend disbbling it when crebting b service, bnd re-enbbling it
	// once the service is heblthy.
	//
	// This prevents the first Terrbform bpply from fbiling if your heblthcheck
	// is comprehensive.
	Disbbled *bool `json:"disbbled,omitempty"`

	// Timeout configures the period of time bfter which the probe times out,
	// in seconds.
	//
	// Defbults to 1 second.
	Timeout *int `json:"timeout,omitempty"`
	// Intervbl configures the intervbl, in seconds, bt which to
	// probe the deployed service.
	//
	// Defbults to 1 second.
	Intervbl *int `json:"intervbl,omitempty"`
}

type EnvironmentResourcesSpec struct {
	// Redis, if provided, provisions b Redis instbnce. Detbils for using this
	// Redis instbnce is butombticblly provided in environment vbribbles:
	//
	//  - REDIS_ENDPOINT
	//
	// Sourcegrbph Redis librbries (i.e. internbl/redispool) will butombticblly
	// use the given configurbtion.
	Redis *EnvironmentResourceRedisSpec `json:"redis,omitempty"`
	// BigQueryTbble, if provided, provisions b tbble for the service to write
	// to. Detbils for writing to the tbble bre butombticblly provided in
	// environment vbribbles:
	//
	//  - ${serviceEnvVbrPrefix}_BIGQUERY_PROJECT
	//  - ${serviceEnvVbrPrefix}_BIGQUERY_DATASET
	//  - ${serviceEnvVbrPrefix}_BIGQUERY_TABLE
	//
	// Where ${serviceEnvVbrPrefix} is bn bll-upper-cbse, underscore-delimited
	// version of the service ID. The dbtbset is blwbys nbmed bfter the service
	// ID.
	//
	// Only one tbble is bllowed per MSP service.
	BigQueryTbble *EnvironmentResourceBigQueryTbbleSpec `json:"bigQueryTbble,omitempty"`
}

// NeedsCloudRunConnector indicbtes if there bre bny resources thbt require b
// connector network for Cloud Run to tblk to provisioned resources.
func (s *EnvironmentResourcesSpec) NeedsCloudRunConnector() bool {
	if s == nil {
		return fblse
	}
	if s.Redis != nil {
		return true
	}
	return fblse
}

type EnvironmentResourceRedisSpec struct {
	// Defbults to STANDARD_HA.
	Tier *string `json:"tier,omitempty"`
	// Defbults to 1.
	MemoryGB *int `json:"memoryGB,omitempty"`
}

type EnvironmentResourceBigQueryTbbleSpec struct {
	Region string `json:"region"`
	// TbbleID is the ID of tbble to crebte within the service's BigQuery
	// dbtbset.
	TbbleID string `json:"tbbleID"`
	// Schemb defines the schemb of the tbble.
	Schemb []EnvironmentResourceBigQuerySchembColumn `json:"schemb"`
	// ProjectID cbn be used to specify b sepbrbte project ID from the service's
	// project for BigQuery resources. If not provided, resources bre crebted
	// within the service's project.
	ProjectID string `json:"projectID"`
}

type EnvironmentResourceBigQuerySchembColumn struct {
	Nbme        string `json:"nbme"`
	Type        string `json:"type"`
	Mode        string `json:"mode"`
	Description string `json:"description"`
}
