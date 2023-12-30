package spec

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/imageupdater"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type EnvironmentSpec struct {
	// ID is an all-lowercase alphanumeric identifier for the deployment
	// environment, e.g. "prod" or "dev".
	ID string `yaml:"id"`

	// ProjectID is the generated Google Project ID for this environment,
	// provided by either 'sg msp init' or 'sg msp init -env'.
	//
	// The format is:
	//
	// 	$SERVICE_ID-$ENVIRONMENT_ID-$RANDOM_SUFFIX
	//
	// ❗ This value cannot be changed after your environment has been
	// initialized!
	ProjectID string `yaml:"projectID"`

	// Category is either "test", "internal", or "external". It informs the
	// which GCP project folder the environment lives in:
	//
	// - 'test': 'Engineering Projects' folder (liberal access)
	// - 'internal': 'Internal Services' folder (restricted access)
	// - 'external': 'Managed Services' folder (restricted access)
	//
	// It also informs what kind of notification channels are set up out-of-the-box:
	//
	// 1. 'test' services only generate Slack notifications.
	// 2. 'internal' and 'external' services generate Slack and Opsgenie notifications.
	//
	// Slack channels are expected to be named '#alerts-<service>-<environmentName>'.
	// Opsgenie teams are expected to correspond to service owners.
	//
	// Both Slack channels and Opsgenie teams are currently expected to be manually
	// configured.
	Category EnvironmentCategory `yaml:"category"`

	// Deploy specifies how to deploy revisions.
	Deploy EnvironmentDeploySpec `yaml:"deploy"`

	// EnvironmentServiceSpec carries service-specific configuration.
	*EnvironmentServiceSpec `yaml:",inline"`
	// EnvironmentJobSpec carries job-specific configuration.
	*EnvironmentJobSpec `yaml:",inline"`

	// Instances describes how machines running the service are deployed.
	Instances EnvironmentInstancesSpec `yaml:"instances"`

	// Env are key-value pairs of environment variables to set on the service.
	//
	// Values can be subsituted with supported runtime values with gotemplate, e.g., "{{ .ProjectID }}"
	// 	- ProjectID: The project ID of the service.
	//	- ServiceDnsName: The DNS name of the service.
	Env map[string]string `yaml:"env,omitempty"`

	// SecretEnv are key-value pairs of environment variables sourced from
	// secrets set on the service, where the value is the name of the secret
	// in the service's project to populate in the environment.
	//
	// To point to a secret in another project, use the format
	// 'projects/{project}/secrets/{secretName}' in the value. Access to the
	// target project will be automatically granted.
	SecretEnv map[string]string `yaml:"secretEnv,omitempty"`

	// Resources configures additional resources that a service may depend on.
	Resources *EnvironmentResourcesSpec `yaml:"resources,omitempty"`

	// AllowDestroys, if false, configures Terraform lifecycle guards against
	// deletion of potentially critical resources. This includes things like the
	// environment project and databases.
	// https://developer.hashicorp.com/terraform/tutorials/state/resource-lifecycle#prevent-resource-deletion
	//
	// To tear down an environment, or to apply a change that intentionally
	// causes deletion on guarded resources, set this to true and apply the
	// generated Terraform first.
	AllowDestroys *bool `yaml:"allowDestroys,omitempty"`
}

func (s EnvironmentSpec) Validate() []error {
	var errs []error

	if s.ProjectID == "" {
		errs = append(errs, errors.New("projectID is required"))
	}
	if len(s.ProjectID) > 30 {
		errs = append(errs, errors.New("projectID must be less than 30 characters"))
	}
	if !strings.Contains(s.ProjectID, fmt.Sprintf("-%s-", s.ID)) {
		errs = append(errs, errors.Newf("projectID %q must contain environment ID: expecting format '$SERVICE_ID-$ENVIRONMENT_ID-$RANDOM_SUFFIX'",
			s.ProjectID))
	}
	if err := s.Category.Validate(); err != nil {
		return append(errs, errors.Wrap(err, "category"))
	}

	errs = append(errs, s.Deploy.Validate()...)
	errs = append(errs, s.Resources.Validate()...)
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

func (c EnvironmentCategory) Validate() error {
	switch c {
	case EnvironmentCategoryTest,
		EnvironmentCategoryInternal,
		EnvironmentCategoryExternal:
	case "":
		return errors.New("no category provided")
	default:
		return errors.Newf("invalid category %q", c)
	}
	return nil
}

type EnvironmentDeploySpec struct {
	Type         EnvironmentDeployType                  `yaml:"type"`
	Manual       *EnvironmentDeployManualSpec           `yaml:"manual,omitempty"`
	Subscription *EnvironmentDeployTypeSubscriptionSpec `yaml:"subscription,omitempty"`
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
	Tag string `yaml:"tag,omitempty"`
}

type EnvironmentDeployTypeSubscriptionSpec struct {
	// Tag is the tag to subscribe to.
	Tag string `yaml:"tag,omitempty"`
	// TODO: In the future, we may support subscribing by semver constraints.
}

type EnvironmentServiceSpec struct {
	// Domain configures where the resource is externally accessible.
	//
	// Only supported for services of 'kind: service'.
	Domain *EnvironmentServiceDomainSpec `yaml:"domain,omitempty"`
	// StatupProbe is provisioned by default. It can be disabled with the
	// 'disabled' field. Probes are made to the MSP-standard '/-/healthz'
	// endpoint.
	//
	// Only supported for services of 'kind: service'.
	StatupProbe *EnvironmentServiceStartupProbeSpec `yaml:"startupProbe,omitempty"`
	// LivenessProbe is only provisioned if this field is set. Probes are made
	// to the MSP-standard '/-/healthz' endpoint.
	//
	// Only supported for services of 'kind: service'.
	LivenessProbe *EnvironmentServiceLivenessProbeSpec `yaml:"livenessProbe,omitempty"`
	// Authentication configures access to the service. By default, the service
	// is publically available, and the service should handle any required
	// authentication by itself. Set this field to an empty value to not
	// configure any access to the service at all.
	//
	// More complex strategies, such as granting access to specific groups,
	// should add custom resources in the MSP IAM module defining
	// 'google_cloud_run_v2_service_iam_member' for 'roles/run.invoker' on
	// 'local.cloud_run_resource_name' and 'local.cloud_run_location' as 'name'
	// and 'location' respectively:
	// https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/cloud_run_v2_service_iam
	//
	// Only supported for services of 'kind: service'.
	Authentication *EnvironmentServiceAuthenticationSpec `yaml:"authentication,omitempty"`
}

type EnvironmentServiceDomainSpec struct {
	// Type is one of 'none' or 'cloudflare'. If empty, defaults to 'none'.
	Type       EnvironmentDomainType            `yaml:"type"`
	Cloudflare *EnvironmentDomainCloudflareSpec `yaml:"cloudflare,omitempty"`
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
	Subdomain string `yaml:"subdomain"`
	Zone      string `yaml:"zone"`

	// Proxied configures whether Cloudflare should proxy all traffic to get
	// WAF protection instead of only DNS resolution.
	Proxied bool `yaml:"proxied,omitempty"`

	// Required configures whether traffic can only be allowed through Cloudflare.
	// TODO: Unimplemented.
	Required bool `yaml:"required,omitempty"`
}

type EnvironmentInstancesSpec struct {
	Resources EnvironmentInstancesResourcesSpec `yaml:"resources"`
	// Scaling specifies the scaling behavior of the service.
	//
	// Currently only used for services of 'kind: service'.
	Scaling *EnvironmentInstancesScalingSpec `yaml:"scaling,omitempty"`
}

type EnvironmentInstancesResourcesSpec struct {
	CPU    int    `yaml:"cpu"`
	Memory string `yaml:"memory"`
}

type EnvironmentInstancesScalingSpec struct {
	// MaxRequestConcurrency is the maximum number of concurrent requests that
	// each instance is allowed to serve. Before this concurrency is reached,
	// Cloud Run will begin scaling up additional instances, up to MaxCount.
	//
	// If not provided, the defualt is defaultMaxConcurrentRequests
	MaxRequestConcurrency *int `yaml:"maxRequestConcurrency,omitempty"`
	// MinCount is the minimum number of instances that will be running at all
	// times. Set this to >0 to avoid service warm-up delays.
	MinCount int `yaml:"minCount"`
	// MaxCount is the maximum number of instances that Cloud Run is allowed to
	// scale up to.
	//
	// If not provided, the default is 5.
	MaxCount *int `yaml:"maxCount,omitempty"`
}

type EnvironmentServiceLivenessProbeSpec struct {
	// Timeout configures the period of time after which the probe times out,
	// in seconds.
	//
	// Defaults to 1 second.
	Timeout *int `yaml:"timeout,omitempty"`
	// Interval configures the interval, in seconds, at which to
	// probe the deployed service.
	//
	// Defaults to 1 second.
	Interval *int `yaml:"interval,omitempty"`
}

type EnvironmentServiceAuthenticationSpec struct {
	// Sourcegraph enables access to everyone in the sourcegraph.com GSuite
	// domain.
	Sourcegraph *bool `yaml:"sourcegraph,omitempty"`
}

type EnvironmentServiceStartupProbeSpec struct {
	// Disabled configures whether the startup probe should be disabled.
	// We recommend disabling it when creating a service, and re-enabling it
	// once the service is healthy.
	//
	// This prevents the first Terraform apply from failing if your healthcheck
	// is comprehensive.
	Disabled *bool `yaml:"disabled,omitempty"`

	// Timeout configures the period of time after which the probe times out,
	// in seconds.
	//
	// Defaults to 1 second.
	Timeout *int `yaml:"timeout,omitempty"`
	// Interval configures the interval, in seconds, at which to
	// probe the deployed service.
	//
	// Defaults to 1 second.
	Interval *int `yaml:"interval,omitempty"`
}

type EnvironmentJobSpec struct {
	// Schedule configures a cron schedule for the service.
	//
	// Only supported for services of 'kind: job'.
	Schedule *EnvironmentJobScheduleSpec `yaml:"schedule,omitempty"`
}

type EnvironmentJobScheduleSpec struct {
	// Cron is a cron schedule in the form of "* * * * *".
	Cron string `yaml:"cron"`
	// Deadline of each attempt, in seconds.
	Deadline *int `yaml:"deadline,omitempty"`
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
	Redis *EnvironmentResourceRedisSpec `yaml:"redis,omitempty"`
	// PostgreSQL, if provided, provisions a PostgreSQL database instance backed
	// by Cloud SQL.
	//
	// To connect to the database, use
	// (lib/managedservicesplatform/service.Contract).GetPostgreSQLDB().
	PostgreSQL *EnvironmentResourcePostgreSQLSpec `yaml:"postgreSQL,omitempty"`
	// BigQueryDataset, if provided, provisions a dataset for the service to write
	// to. Details for writing to the dataset are automatically provided in
	// environment variables:
	//
	//  - BIGQUERY_PROJECT
	//  - BIGQUERY_DATASET
	//
	// Only one dataset can be provisioned using MSP per MSP service, but the
	// dataset may contain more than one table.
	BigQueryDataset *EnvironmentResourceBigQueryDatasetSpec `yaml:"bigQueryDataset,omitempty"`
}

func (s *EnvironmentResourcesSpec) Validate() []error {
	if s == nil {
		return nil
	}
	var errs []error
	errs = append(errs, s.PostgreSQL.Validate()...)
	errs = append(errs, s.BigQueryDataset.Validate()...)
	return errs
}

type EnvironmentResourceRedisSpec struct {
	// Defaults to STANDARD_HA.
	Tier *string `yaml:"tier,omitempty"`
	// Defaults to 1.
	MemoryGB *int `yaml:"memoryGB,omitempty"`
}

type EnvironmentResourcePostgreSQLSpec struct {
	// Databases to provision - required.
	Databases []string `yaml:"databases"`
	// Defaults to 1. Must be 1, or an even number between 2 and 96.
	CPU *int `yaml:"cpu,omitempty"`
	// Defaults to 4 (to meet CloudSQL minimum). You must request 0.9 to 6.5 GB
	// per vCPU.
	MemoryGB *int `yaml:"memoryGB,omitempty"`
}

func (s *EnvironmentResourcePostgreSQLSpec) Validate() []error {
	if s == nil {
		return nil
	}
	var errs []error
	if s.CPU != nil {
		if *s.CPU < 1 {
			errs = append(errs, errors.New("postgreSQL.cpu must be >= 1"))
		}
		if *s.CPU > 1 && *s.CPU%2 != 0 {
			errs = append(errs, errors.New("postgreSQL.cpu must be 1 or a multiple of 2"))
		}
		if *s.CPU > 96 {
			errs = append(errs, errors.New("postgreSQL.cpu must be <= 96"))
		}
	}
	if s.MemoryGB != nil {
		cpu := pointers.Deref(s.CPU, 1)
		if *s.MemoryGB < 4 {
			errs = append(errs, errors.New("postgreSQL.memoryGB must be >= 4"))
		}
		if *s.MemoryGB < cpu {
			errs = append(errs, errors.New("postgreSQL.memoryGB must be >= postgreSQL.cpu"))
		}
		if *s.MemoryGB > 6*cpu {
			errs = append(errs, errors.New("postgreSQL.memoryGB must be <= 6*postgreSQL.cpu"))
		}
	}
	return errs
}

type EnvironmentResourceBigQueryDatasetSpec struct {
	// Tables are the IDs of tables to create within the service's BigQuery
	// dataset. Required.
	//
	// For EACH table, a BigQuery JSON schema MUST be provided alongside the
	// service specification file, in `${tableID}.bigquerytable.json`. Learn
	// more about BigQuery table schemas here:
	// https://cloud.google.com/bigquery/docs/schemas#specifying_a_json_schema_file
	Tables []string `yaml:"tables"`
	// rawSchemaFiles are the `${tableID}.bigquerytable.json` files adjacent
	// to the service specification.
	// Loaded by (EnvironmentResourceBigQueryTableSpec).LoadSchemas().
	rawSchemaFiles map[string][]byte

	// DatasetID, if provided, configures a custom dataset ID to place all tables
	// into. By default, we use the service ID as the dataset ID.
	//
	// Dataset IDs must be alphanumeric (plus underscores).
	DatasetID *string `yaml:"datasetID,omitempty"`
	// ProjectID can be used to specify a separate project ID from the service's
	// project for BigQuery resources. If not provided, resources are created
	// within the service's project.
	ProjectID *string `yaml:"projectID,omitempty"`
	// Location defaults to "US". Do not configure unless you know what you are
	// doing, as BigQuery locations are not the same as standard GCP regions.
	Location *string `yaml:"region,omitempty"`
}

func (s *EnvironmentResourceBigQueryDatasetSpec) Validate() []error {
	if s == nil {
		return nil
	}
	var errs []error
	if len(s.Tables) == 0 {
		errs = append(errs, errors.New("bigQueryDataset.tables must be non-empty"))
	}
	return errs
}

// LoadSchemas populates rawSchemaFiles by convention, looking for
// `bigquery.${tableID}.schema.json` files in dir.
func (s *EnvironmentResourceBigQueryDatasetSpec) LoadSchemas(dir string) error {
	s.rawSchemaFiles = make(map[string][]byte, len(s.Tables))

	// Make sure all tables have a schema file
	for _, table := range s.Tables {
		// Open by convention
		schema, err := os.ReadFile(filepath.Join(dir, fmt.Sprintf("%s.bigquerytable.json", table)))
		if err != nil {
			return errors.Wrapf(err, "read schema for BigQuery table %s", table)
		}

		// Parse and marshal for consistent formatting. Note that the table
		// must be a JSON array.
		var schemaData []any
		if err := json.Unmarshal(schema, &schemaData); err != nil {
			return errors.Wrapf(err, "parse schema for BigQuery table %s", table)
		}
		s.rawSchemaFiles[table], err = json.MarshalIndent(schemaData, "", "  ")
		if err != nil {
			return errors.Wrapf(err, "marshal schema for BigQuery table %s", table)
		}
	}

	return nil
}

// GetSchema returns the schema for the given tableID as loaded by LoadSchemas().
// LoadSchemas will ensure that each table has a corresponding schema file.
func (s *EnvironmentResourceBigQueryDatasetSpec) GetSchema(tableID string) []byte {
	return s.rawSchemaFiles[tableID]
}
