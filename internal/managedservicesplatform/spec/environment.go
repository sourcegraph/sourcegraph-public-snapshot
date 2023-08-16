package spec

import "github.com/sourcegraph/sourcegraph/lib/errors"

type EnvironmentSpec struct {
	Name      string                   `json:"name"`
	Deploy    EnvironmentDeploySpec    `json:"deploy"`
	Domain    EnvironmentDomainSpec    `json:"domain"`
	Instances EnvironmentInstancesSpec `json:"instances"`
	Resources EnvironmentResourcesSpec `json:"resources"`
	Env       map[string]string        `json:"env"`
	SecretEnv map[string]string        `json:"secretEnv"`
}

type EnvironmentDeploySpec struct {
	Type   string                       `json:"type"` // TODO typed string
	Manual *EnvironmentDeployManualSpec `json:"manual"`
}

func (d EnvironmentDeploySpec) ResolveTag() (string, error) {
	switch d.Type {
	case "manual":
		if d.Manual == nil {
			return "insiders", nil
		}
		return d.Manual.Tag, nil

	default:
		return "", errors.New("unable to resolve tag")
	}
}

type EnvironmentDeployManualSpec struct {
	Tag string `json:"tag"`
}

type EnvironmentDomainSpec struct {
	Type       string                           `json:"type"`
	Cloudflare *EnvironmentDomainCloudflareSpec `json:"cloudflare"`
}

type EnvironmentDomainCloudflareSpec struct {
	Subdomain string `json:"subdomain"`
	Zone      string `json:"zone"`

	Required bool `json:"required"`
}

type EnvironmentInstancesSpec struct {
	Resources EnvironmentInstancesResourcesSpec `json:"resources"`
	Scaling   EnvironmentInstancesScalingSpec   `json:"scaling"`
}

type EnvironmentInstancesResourcesSpec struct {
	CPU    int    `json:"cpu"`
	Memory string `json:"memory"`
}

type EnvironmentInstancesScalingSpec struct {
	MaxRequestConcurrency int `json:"max_request_concurrency"`
	MinCount              int `json:"min_count"`
	MaxCount              int `json:"max_count"`
}

type EnvironmentHealthcheckSpec struct {
	// In seconds
	LivenessProbeInterval int `json:"liveness_probe_interval"`
}

type EnvironmentResourcesSpec struct {
	Redis    *EnvironmentResourceRedisSpec    `json:"redis"`
	BigQuery *EnvironmentResourceBigQuerySpec `json:"big_query"`
}

type EnvironmentResourceRedisSpec struct {
	Tier   string `json:"tier"`
	Memory string `json:"memory"`
}

type EnvironmentResourceBigQuerySpec struct {
	ProjectID string `json:"project_id"`
	Region    string `json:"region"`
}
