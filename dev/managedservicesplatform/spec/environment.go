package spec

import (
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/imageupdater"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type EnvironmentSpec struct {
	// ID is an all-lowercase alphanumeric identifier for the deployment
	// environment, e.g. "prod" or "dev".
	ID string `json:"id"`

	// Category is either "test", "internal", or "external".
	Category *EnvironmentCategory `json:"category,omitempty"`

	Deploy    EnvironmentDeploySpec    `json:"deploy"`
	Domain    EnvironmentDomainSpec    `json:"domain"`
	Instances EnvironmentInstancesSpec `json:"instances"`

	Resources *EnvironmentResourcesSpec `json:"resources,omitempty"`

	// StatupProbe is provisioned by default. It can be disabled with the
	// 'disabled' field.
	StatupProbe *EnvironmentStartupProbeSpec `json:"startupProbe,omitempty"`
	// LivenessProbe is only provisioned if this field is set.
	LivenessProbe *EnvironmentLivenessProbeSpec `json:"livenessProbe,omitempty"`

	// Env is key-value pairs of environment variables to set on the service.
	//
	// Values can be subsituted with supported runtime values with gotemplate, e.g., "{{ .ProjectID }}"
	// 	- ProjectID: The project ID of the service.
	//	- ServiceDnsName: The DNS name of the service.
	Env       map[string]string `json:"env,omitempty"`
	SecretEnv map[string]string `json:"secretEnv,omitempty"`
}

func (s EnvironmentSpec) Validate() []error {
	var errs []error
	errs = append(errs, s.Deploy.Validate()...)
	return errs
}

type EnvironmentCategory string

const (
	// EnvironmentCategoryTest should be used for testing and development
	// environments.
	EnvironmentCategoryTest EnvironmentCategory = "test"
	// EnvironmentCategoryInternal should be used for internal environments.
	EnvironmentCategoryInternal EnvironmentCategory = "internal"
	// EnvironmentCategoryExternal is the default category if none is specified.
	EnvironmentCategoryExternal EnvironmentCategory = "external"
)

type EnvironmentDeploySpec struct {
	Type         EnvironmentDeployType                  `json:"type"`
	Manual       *EnvironmentDeployManualSpec           `json:"manual,omitempty"`
	Subscription *EnvironmentDeployTypeSubscriptionSpec `json:"subscription,omitempty"`
}

func (s EnvironmentDeploySpec) Validate() []error {
	var errs []error
	if s.Type == EnvironmentDeployTypeSubscription {
		if s.Subscription == nil {
			errs = append(errs, errors.New("no subscription specified when deploy type is subscription"))
		}
		if s.Subscription.Tag == "" {
			errs = append(errs, errors.New("no tag in image subscription specified"))
		}
	}
	return errs
}

type EnvironmentDeployType string

const (
	EnvironmentDeployTypeManual       = "manual"
	EnvironmentDeployTypeSubscription = "subscription"
)

// ResolveTag uses the deploy spec to resolve an appropriate tag for the environment.
//
// TODO: Implement ability to resolve latest concrete tag from a source
func (d EnvironmentDeploySpec) ResolveTag(repo string) (string, error) {
	switch d.Type {
	case EnvironmentDeployTypeManual:
		if d.Manual == nil {
			return "insiders", nil
		}
		return d.Manual.Tag, nil
	case EnvironmentDeployTypeSubscription:
		// we already validated in Validate(), hence it's fine to assume this won't panic
		updater, err := imageupdater.New()
		if err != nil {
			return "", errors.Wrapf(err, "create image updater")
		}
		tagAndDigest, err := updater.ResolveTagAndDigest(repo, d.Subscription.Tag)
		if err != nil {
			return "", errors.Wrapf(err, "resolve digest for tag %q", "insiders")
		}
		return tagAndDigest, nil
	default:
		return "", errors.New("unable to resolve tag")
	}
}

type EnvironmentDeployManualSpec struct {
	// Tag is the tag to deploy. If empty, defaults to "insiders".
	Tag string `json:"tag,omitempty"`
}

type EnvironmentDeployTypeSubscriptionSpec struct {
	// Tag is the tag to subscribe to.
	Tag string `json:"tag,omitempty"`
	// TODO: In the future, we may support subscribing by semver constraints.
}

type EnvironmentDomainSpec struct {
	// Type is one of 'none' or 'cloudflare'. If empty, defaults to 'none'.
	Type       EnvironmentDomainType            `json:"type"`
	Cloudflare *EnvironmentDomainCloudflareSpec `json:"cloudflare,omitempty"`
}

type EnvironmentDomainType string

const (
	EnvironmentDomainTypeNone       = "none"
	EnvironmentDomainTypeCloudflare = "cloudflare"
)

type EnvironmentDomainCloudflareSpec struct {
	Subdomain string `json:"subdomain"`
	Zone      string `json:"zone"`

	// Proxied configures whether Cloudflare should proxy all traffic to get
	// WAF protection instead of only DNS resolution.
	Proxied bool `json:"proxied,omitempty"`

	// Required configures whether traffic can only be allowed through Cloudflare.
	// TODO: Unimplemented.
	Required bool `json:"required,omitempty"`
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
	//
	// If not provided, the defualt is defaultMaxConcurrentRequests
	MaxRequestConcurrency *int `json:"maxRequestConcurrency,omitempty"`
	// MinCount is the minimum number of instances that will be running at all
	// times. Set this to >0 to avoid service warm-up delays.
	MinCount int `json:"minCount"`
	// MaxCount is the maximum number of instances that Cloud Run is allowed to
	// scale up to.
	//
	// If not provided, the default is 5.
	MaxCount *int `json:"maxCount,omitempty"`
}

type EnvironmentLivenessProbeSpec struct {
	// Timeout configures the period of time after which the probe times out,
	// in seconds.
	//
	// Defaults to 1 second.
	Timeout *int `json:"timeout,omitempty"`
	// Interval configures the interval, in seconds, at which to
	// probe the deployed service.
	//
	// Defaults to 1 second.
	Interval *int `json:"interval,omitempty"`
}

type EnvironmentStartupProbeSpec struct {
	// Disabled configures whether the startup probe should be disabled.
	// We recommend disabling it when creating a service, and re-enabling it
	// once the service is healthy.
	//
	// This prevents the first Terraform apply from failing if your healthcheck
	// is comprehensive.
	Disabled *bool `json:"disabled,omitempty"`

	// Timeout configures the period of time after which the probe times out,
	// in seconds.
	//
	// Defaults to 1 second.
	Timeout *int `json:"timeout,omitempty"`
	// Interval configures the interval, in seconds, at which to
	// probe the deployed service.
	//
	// Defaults to 1 second.
	Interval *int `json:"interval,omitempty"`
}

type EnvironmentResourcesSpec struct {
	// Redis, if provided, provisions a Redis instance. Details for using this
	// Redis instance is automatically provided in environment variables:
	//
	//  - REDIS_ENDPOINT
	//
	// Sourcegraph Redis libraries (i.e. internal/redispool) will automatically
	// use the given configuration.
	Redis *EnvironmentResourceRedisSpec `json:"redis,omitempty"`
	// BigQueryTable, if provided, provisions a table for the service to write
	// to. Details for writing to the table are automatically provided in
	// environment variables:
	//
	//  - ${serviceEnvVarPrefix}_BIGQUERY_PROJECT
	//  - ${serviceEnvVarPrefix}_BIGQUERY_DATASET
	//  - ${serviceEnvVarPrefix}_BIGQUERY_TABLE
	//
	// Where ${serviceEnvVarPrefix} is an all-upper-case, underscore-delimited
	// version of the service ID. The dataset is always named after the service
	// ID.
	//
	// Only one table is allowed per MSP service.
	BigQueryTable *EnvironmentResourceBigQueryTableSpec `json:"bigQueryTable,omitempty"`
}

// NeedsCloudRunConnector indicates if there are any resources that require a
// connector network for Cloud Run to talk to provisioned resources.
func (s *EnvironmentResourcesSpec) NeedsCloudRunConnector() bool {
	if s == nil {
		return false
	}
	if s.Redis != nil {
		return true
	}
	return false
}

type EnvironmentResourceRedisSpec struct {
	// Defaults to STANDARD_HA.
	Tier *string `json:"tier,omitempty"`
	// Defaults to 1.
	MemoryGB *int `json:"memoryGB,omitempty"`
}

type EnvironmentResourceBigQueryTableSpec struct {
	Region string `json:"region"`
	// TableID is the ID of table to create within the service's BigQuery
	// dataset.
	TableID string `json:"tableID"`
	// Schema defines the schema of the table.
	Schema []EnvironmentResourceBigQuerySchemaColumn `json:"schema"`
	// ProjectID can be used to specify a separate project ID from the service's
	// project for BigQuery resources. If not provided, resources are created
	// within the service's project.
	ProjectID string `json:"projectID"`
}

type EnvironmentResourceBigQuerySchemaColumn struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Mode        string `json:"mode"`
	Description string `json:"description"`
}
