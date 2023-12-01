package cloudrun

import (
	"bytes"
	"html/template"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"

	"github.com/hashicorp/terraform-cdk-go/cdktf"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/googlesecretsmanager"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/bigquery"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/cloudsql"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/gsmsecret"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/postgresqlroles"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/privatenetwork"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/random"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/redis"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/serviceaccount"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/tfvar"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack/cloudrun/internal/builder"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack/cloudrun/internal/builder/job"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack/cloudrun/internal/builder/service"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack/options/cloudflareprovider"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack/options/dynamicvariables"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack/options/googleprovider"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack/options/randomprovider"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type CrossStackOutput struct{}

type Variables struct {
	ProjectID                      string
	CloudRunWorkloadServiceAccount *serviceaccount.Output

	Service     spec.ServiceSpec
	Image       string
	Environment spec.EnvironmentSpec

	StableGenerate bool
}

const StackName = "cloudrun"

// Hardcoded variables.
var (
	// gcpRegion is currently hardcoded.
	gcpRegion = "us-central1"
)

// makeServiceEnvVarPrefix returns the env var prefix for service-specific
// env vars that will be set on the Cloud Run service, i.e.
//
// - ${local.env_var_prefix}_BIGQUERY_PROJECT_ID
// - ${local.env_var_prefix}_BIGQUERY_DATASET
// - ${local.env_var_prefix}_BIGQUERY_TABLE
//
// The prefix is an all-uppercase underscore-delimited version of the service ID,
// for example:
//
//	cody-gateway
//
// The prefix for various env vars will be:
//
//	CODY_GATEWAY_
//
// Note that some variables conforming to conventions like DIAGNOSTICS_SECRET,
// GOOGLE_PROJECT_ID, and REDIS_ENDPOINT do not get prefixed, and custom env
// vars configured on an environment are not automatically prefixed either.
func makeServiceEnvVarPrefix(serviceID string) string {
	return strings.ToUpper(strings.ReplaceAll(serviceID, "-", "_")) + "_"
}

const tfVarKeyResolvedImageTag = "resolved_image_tag"

// NewStack instantiates the MSP cloudrun stack, which is currently a pretty
// monolithic stack that encompasses all the core components of an MSP service,
// including networking and dependencies like Redis.
func NewStack(stacks *stack.Set, vars Variables) (crossStackOutput *CrossStackOutput, _ error) {
	stack, locals, err := stacks.New(StackName,
		googleprovider.With(vars.ProjectID),
		cloudflareprovider.With(gsmsecret.DataConfig{
			Secret:    googlesecretsmanager.SecretCloudflareAPIToken,
			ProjectID: googlesecretsmanager.ProjectID,
		}),
		randomprovider.With(),
		dynamicvariables.With(vars.StableGenerate, func() (stack.TFVars, error) {
			resolvedImageTag, err := vars.Environment.Deploy.ResolveTag(vars.Image)
			return stack.TFVars{tfVarKeyResolvedImageTag: resolvedImageTag}, err
		}))
	if err != nil {
		return nil, err
	}

	// Set up a service-specific env var prefix to avoid conflicts where relevant
	serviceEnvVarPrefix := pointers.Deref(
		vars.Service.EnvVarPrefix,
		makeServiceEnvVarPrefix(vars.Service.ID))

	diagnosticsSecret := random.New(stack, resourceid.New("diagnostics-secret"), random.Config{
		ByteLength: 8,
	})

	id := resourceid.New("cloudrun")

	// Set up configuration for the Cloud Run resources
	var cloudRunBuilder builder.Builder
	switch pointers.Deref(vars.Service.Kind, spec.ServiceKindService) {
	case spec.ServiceKindService:
		cloudRunBuilder = service.NewBuilder()
	case spec.ServiceKindJob:
		cloudRunBuilder = job.NewBuilder()
	}

	// Required to enable tracing etc.
	//
	// We don't use serviceEnvVarPrefix here because this is a
	// convention to indicate the environment's project.
	cloudRunBuilder.AddEnv("GOOGLE_CLOUD_PROJECT", vars.ProjectID)

	// Set up secret that service should accept for diagnostics
	// endpoints.
	//
	// We don't use serviceEnvVarPrefix here because this is a
	// convention across MSP services.
	cloudRunBuilder.AddEnv("DIAGNOSTICS_SECRET", diagnosticsSecret.HexValue)

	// Add the domain as an environment variable.
	dnsName := pointers.DerefZero(vars.Environment.EnvironmentServiceSpec).Domain.GetDNSName()
	if dnsName != "" {
		cloudRunBuilder.AddEnv("EXTERNAL_DNS_NAME", dnsName)
	}

	// Add user-configured env vars
	if err := addContainerEnvVars(cloudRunBuilder, vars.Environment.Env, vars.Environment.SecretEnv, envVariablesData{
		ProjectID:      vars.ProjectID,
		ServiceDnsName: dnsName,
	}); err != nil {
		return nil, errors.Wrap(err, "add user env vars")
	}

	// Load image tag from tfvars.
	imageTag := tfvar.New(stack, id, tfvar.Config{
		VariableKey: tfVarKeyResolvedImageTag,
		Description: "Resolved image tag to deploy",
	})

	// privateNetworkEnabled indicates if privateNetwork has been instantiated
	// before.
	var privateNetworkEnabled bool
	// privateNetwork is only instantiated if used, and is only instantiated
	// once. If called, it always returns a non-nil value.
	privateNetwork := sync.OnceValue(func() *privatenetwork.Output {
		privateNetworkEnabled = true
		return privatenetwork.New(stack, privatenetwork.Config{
			ProjectID: vars.ProjectID,
			ServiceID: vars.Service.ID,
			Region:    gcpRegion,
		})
	})

	// Add MSP env var indicating that the service is running in a Managed
	// Services Platform environment.
	cloudRunBuilder.AddEnv("MSP", "true")

	// For SSL_CERT_DIR, configure right before final build
	sslCertDirs := []string{"/etc/ssl/certs"}

	// redisInstance is only created and non-nil if Redis is configured for the
	// environment.
	if vars.Environment.Resources != nil && vars.Environment.Resources.Redis != nil {
		redisInstance, err := redis.New(stack,
			resourceid.New("redis"),
			redis.Config{
				ProjectID: vars.ProjectID,
				Region:    gcpRegion,
				Spec:      *vars.Environment.Resources.Redis,
				Network:   privateNetwork().Network,
			})
		if err != nil {
			return nil, errors.Wrap(err, "failed to render Redis instance")
		}

		// We don't use serviceEnvVarPrefix here because this is a
		// Sourcegraph-wide convention.
		cloudRunBuilder.AddEnv("REDIS_ENDPOINT", redisInstance.Endpoint)

		// Mount the custom cert and add it to SSL_CERT_DIR
		caCertVolumeName := "redis-ca-cert"
		cloudRunBuilder.AddSecretVolume(
			caCertVolumeName,
			"redis-ca-cert.pem",
			builder.SecretRef{
				Name:    redisInstance.Certificate.ID,
				Version: redisInstance.Certificate.Version,
			},
			292, // 0444 read-only
		)
		cloudRunBuilder.AddVolumeMount(caCertVolumeName, "/etc/ssl/custom-certs")
		sslCertDirs = append(sslCertDirs, "/etc/ssl/custom-certs")
	}

	if vars.Environment.Resources != nil && vars.Environment.Resources.PostgreSQL != nil {
		sqlInstance, err := cloudsql.New(stack, resourceid.New("postgresql"), cloudsql.Config{
			ProjectID: vars.ProjectID,
			Region:    gcpRegion,
			Spec:      *vars.Environment.Resources.PostgreSQL,
			Network:   privateNetwork().Network,

			WorkloadIdentity: *vars.CloudRunWorkloadServiceAccount,

			// ServiceNetworkingConnection is required for Cloud SQL to connect
			// to the private network, so we must wait for it to be provisioned.
			// See https://cloud.google.com/sql/docs/mysql/private-ip#network_requirements
			DependsOn: []cdktf.ITerraformDependable{
				privateNetwork().ServiceNetworkingConnection,
			},
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to render Cloud SQL instance")
		}

		// Add parameters required for authentication
		cloudRunBuilder.AddEnv("PGINSTANCE", *sqlInstance.Instance.ConnectionName())
		cloudRunBuilder.AddEnv("PGUSER", *sqlInstance.WorkloadUser.Name())
		// NOTE: https://pkg.go.dev/cloud.google.com/go/cloudsqlconn#section-readme
		// magically handles certs for us, so we don't need to mount certs in
		// Cloud Run.

		// Apply additional runtime configuration
		pgRoles, err := postgresqlroles.New(stack, id.Group("postgresqlroles"), sqlInstance)
		if err != nil {
			return nil, errors.Wrap(err, "failed to render Cloud SQL PostgreSQL roles")
		}

		// We need the workload superuser role to be granted before Cloud Run
		// can correctly use the database instance
		cloudRunBuilder.AddDependency(pgRoles.WorkloadSuperuserGrant)
	}

	// bigqueryDataset is only created and non-nil if BigQuery is configured for
	// the environment.
	if vars.Environment.Resources != nil && vars.Environment.Resources.BigQueryTable != nil {
		bigqueryDataset, err := bigquery.New(stack, resourceid.New("bigquery"), bigquery.Config{
			DefaultProjectID:       vars.ProjectID,
			ServiceID:              vars.Service.ID,
			WorkloadServiceAccount: vars.CloudRunWorkloadServiceAccount,
			Spec:                   *vars.Environment.Resources.BigQueryTable,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to render BigQuery dataset")
		}

		// Add parameters required for writing to the correct BigQuery table
		cloudRunBuilder.AddEnv(serviceEnvVarPrefix+"BIGQUERY_PROJECT_ID", bigqueryDataset.ProjectID)
		cloudRunBuilder.AddEnv(serviceEnvVarPrefix+"BIGQUERY_DATASET", bigqueryDataset.DatasetID)
		cloudRunBuilder.AddEnv(serviceEnvVarPrefix+"BIGQUERY_TABLE", bigqueryDataset.TableID)
	}

	// Finalize output of builder
	cloudRunBuilder.AddEnv("SSL_CERT_DIR", strings.Join(sslCertDirs, ":"))
	if _, err := cloudRunBuilder.Build(stack, builder.Variables{
		Service:           vars.Service,
		Image:             vars.Image,
		ResolvedImageTag:  *imageTag.StringValue,
		Environment:       vars.Environment,
		GCPProjectID:      vars.ProjectID,
		GCPRegion:         gcpRegion,
		ServiceAccount:    vars.CloudRunWorkloadServiceAccount,
		DiagnosticsSecret: diagnosticsSecret,
		ResourceLimits:    makeContainerResourceLimits(vars.Environment.Instances.Resources),
		PrivateNetwork: func() *privatenetwork.Output {
			if privateNetworkEnabled {
				return privateNetwork()
			}
			return nil
		}(),
	}); err != nil {
		return nil, errors.Wrapf(err, "build Cloud Run resource kind %q", cloudRunBuilder.Kind())
	}

	// Collect outputs
	locals.Add("image_tag", imageTag.StringValue,
		"Resolved tag of service image to deploy")
	return &CrossStackOutput{}, nil
}

type envVariablesData struct {
	ProjectID      string
	ServiceDnsName string
}

func addContainerEnvVars(
	b builder.Builder,
	env map[string]string,
	secretEnv map[string]string,
	varsData envVariablesData,
) error {
	// Apply static env vars
	envKeys := maps.Keys(env)
	slices.Sort(envKeys)
	for _, k := range envKeys {
		tmpl, err := template.New("").Parse(env[k])
		if err != nil {
			return errors.Wrapf(err, "parse env var template: %q", env[k])
		}
		var buf bytes.Buffer
		if err = tmpl.Execute(&buf, varsData); err != nil {
			return errors.Wrapf(err, "execute template: %q", env[k])
		}

		b.AddEnv(k, buf.String())
	}

	// Apply secret env vars
	secretEnvKeys := maps.Keys(secretEnv)
	slices.Sort(secretEnvKeys)
	for _, k := range secretEnvKeys {
		b.AddSecretEnv(k, builder.SecretRef{
			Name:    secretEnv[k],
			Version: "latest",
		})
	}

	return nil
}

func makeContainerResourceLimits(r spec.EnvironmentInstancesResourcesSpec) map[string]*string {
	return map[string]*string{
		"cpu":    pointers.Ptr(strconv.Itoa(r.CPU)),
		"memory": pointers.Ptr(r.Memory),
	}
}
