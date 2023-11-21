package spec

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/imageupdater"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type EnvironmentSpec struct {
	// ID is an all-lowercase alphanumeric identifier for the deployment
	// environment, e.g. "prod" or "dev".
	ID string `json:"id"`

	// Category is either "test", "internal", or "external".
	Category *EnvironmentCategory `json:"category,omitempty"`

	// Deploy specifies how to deploy revisions.
	Deploy EnvironmentDeploySpec `json:"deploy"`

	// EnvironmentServiceSpec carries service-specific configuration.
	*EnvironmentServiceSpec `json:",inline"`
	// EnvironmentJobSpec carries job-specific configuration.
	*EnvironmentJobSpec `json:",inline"`

	// Instances describes how machines running the service are deployed.
	Instances EnvironmentInstancesSpec `json:"instances"`

	// Env are key-value pairs of environment variables to set on the service.
	//
	// Values can be subsituted with supported runtime values with gotemplate, e.g., "{{ .ProjectID }}"
	// 	- ProjectID: The project ID of the service.
	//	- ServiceDnsName: The DNS name of the service.
	Env map[string]string `json:"env,omitempty"`

	// SecretEnv are key-value pairs of environment variables sourced from
	// secrets set on the service, where the value is the name of the secret
	// to populate in the environment.
	SecretEnv map[string]string `json:"secretEnv,omitempty"`

	// Resources configures additional resources that a service may depend on.
	Resources *EnvironmentResourcesSpec `json:"resources,omitempty"`
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

type EnvironmentServiceSpec struct {
	// Domain configures where the resource is externally accessible.
	//
	// Only supported for services of 'kind: service'.
	Domain *EnvironmentServiceDomainSpec `json:"domain,omitempty"`
	// StatupProbe is provisioned by default. It can be disabled with the
	// 'disabled' field.
	//
	// Only supported for services of 'kind: service'.
	StatupProbe *EnvironmentServiceStartupProbeSpec `json:"startupProbe,omitempty"`
	// LivenessProbe is only provisioned if this field is set.
	//
	// Only supported for services of 'kind: service'.
	LivenessProbe *EnvironmentServiceLivenessProbeSpec `json:"livenessProbe,omitempty"`
}

type EnvironmentServiceDomainSpec struct {
	// Type is one of 'none' or 'cloudflare'. If empty, defaults to 'none'.
	Type       EnvironmentDomainType            `json:"type"`
	Cloudflare *EnvironmentDomainCloudflareSpec `json:"cloudflare,omitempty"`
}

// GetDNSName generates the DNS name for the environment. If nil or not configured,
// am empty string is returned.
func (s *EnvironmentServiceDomainSpec) GetDNSName() string {
	if s == nil {
		return ""
	}
	if s.Cloudflare != nil {
		return fmt.Sprintf("%s.%s", s.Cloudflare.Subdomain, s.Cloudflare.Zone)
	}
	return ""
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
	// Scaling specifies the scaling behavior of the service.
	//
	// Currently only used for services of 'kind: service'.
	Scaling *EnvironmentInstancesScalingSpec `json:"scaling,omitempty"`
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

type EnvironmentServiceLivenessProbeSpec struct {
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

type EnvironmentServiceStartupProbeSpec struct {
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

type EnvironmentJobSpec struct {
	// Schedule configures a cron schedule for the service.
	//
	// Only supported for services of 'kind: job'.
	Schedule *EnvironmentJobScheduleSpec `json:"schedule,omitempty"`
}

type EnvironmentJobScheduleSpec struct {
	// Cron is a cron schedule in the form of "* * * * *".
	Cron string `json:"cron"`
	// Deadline of each attempt, in seconds.
	Deadline *int `json:"deadline,omitempty"`
}

type EnvironmentResourcesSpec struct {
	// Redis, if provided, provisions a Redis instance backed by Cloud Memorystore.
	// Details for using this Redis instance is automatically provided in
	// environment variables:
	//
	//  - REDIS_ENDPOINT
	//
	// Sourcegraph Redis libraries (i.e. internal/redispool) will automatically
	// use the given configuration.
	Redis *EnvironmentResourceRedisSpec `json:"redis,omitempty"`
	// PostgreSQL, if provided, provisions a PostgreSQL database instance backed
	// by CloudSQL.
	//
	// TODO
	PostgreSQL *EnvironmentResourcePostgreSQLSpec `json:"postgreSQL,omitempty"`
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
	if s.PostgreSQL != nil {
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

type EnvironmentResourcePostgreSQLSpec struct {
	// Databases to provision - required.
	Databases []string `json:"databases"`
	// Defaults to 1.
	CPU *int `json:"cpu,omitempty"`
	// Defaults to 1.
	MemoryGB *int `json:"memoryGB,omitempty"`

	// DisableDeletionProtection disables the Cloud SQL deletion protection that
	// we enable by default. Only set to true if tearing down the database.
	DisableDeletionProtection bool `json:"disableDeletionProtection,omitempty"`
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
