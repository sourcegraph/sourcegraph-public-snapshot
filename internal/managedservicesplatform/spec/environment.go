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
	Type   EnvironmentDeployType        `json:"type"`
	Manual *EnvironmentDeployManualSpec `json:"manual"`
}

type EnvironmentDeployType string

const (
	EnvironmentDeployTypeManual = "manual"
)

// ResolveTag uses the deploy spec to resolve an appropriate tag for the environment.
//
// TODO
func (d EnvironmentDeploySpec) ResolveTag() (string, error) {
	switch d.Type {
	case EnvironmentDeployTypeManual:
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
	// MaxRequestConcurrency is the maximum number of concurrent requests that
	// each instance is allowed to serve. Before this concurrency is reached,
	// Cloud Run will begin scaling up additional instances, up to MaxCount.
	MaxRequestConcurrency int `json:"max_request_concurrency"`
	// MinCount is the minimum number of instances that will be running at all
	// times. Set this to >0 to avoid service warm-up delays.
	MinCount int `json:"min_count"`
	// MaxCount is the maximum number of instances that Cloud Run is allowed to
	// scale up to.
	//
	// If not provided, the default is 5.
	MaxCount *int `json:"max_count"`
}

type EnvironmentHealthcheckSpec struct {
	// LivenessProbeInterval configures the interval, in seconds, at which to
	// probe the deployed service.
	LivenessProbeInterval int `json:"liveness_probe_interval"`
}

type EnvironmentResourcesSpec struct {
	Redis    *EnvironmentResourceRedisSpec    `json:"redis"`
	BigQuery *EnvironmentResourceBigQuerySpec `json:"big_query"`
}

type EnvironmentResourceRedisSpec struct {
	Tier     string `json:"tier"`
	MemoryGB int    `json:"memoryGB"`
}

type EnvironmentResourceBigQuerySpec struct {
	ProjectID string `json:"project_id"`
	Region    string `json:"region"`
}
