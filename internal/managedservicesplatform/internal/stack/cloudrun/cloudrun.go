package cloudrun

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/cloudrunv2service"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/project"

	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/internal/resource/bigquery"
	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/internal/resource/random"
	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/internal/resource/redis"
	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/internal/resource/serviceaccount"
	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/internal/stack"
	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/internal/stack/options/googleprovider"
	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/internal/stack/options/randomprovider"
	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/internal/pointer"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Output struct{}

type Variables struct {
	Project project.Project

	Service     spec.ServiceSpec
	Image       string
	Environment spec.EnvironmentSpec
}

const StackName = "cloudrun"

// Hardcoded variables.
var (
	// gcpRegion is currently hardcoded.
	gcpRegion = "us-central1"
	// serviceAccountRoles are granted to the service account for the Cloud Run service.
	serviceAccountRoles = []serviceaccount.Role{
		// Allow env vars to source from secrets
		{ID: "role_secret_accessor", Role: "roles/secretmanager.secretAccessor"},
		// Allow service to access private networks
		{ID: "role_compute_networkuser", Role: "roles/compute.networkUser"},
		// Allow service to emit observability
		{ID: "role_cloudtrace_agent", Role: "roles/cloudtrace.agent"},
		{ID: "role_monitoring_metricwriter", Role: "roles/monitoring.metricWriter"},
	}
	// servicePort is provided to the container as $PORT in Cloud Run:
	// https://cloud.google.com/run/docs/configuring/services/containers#configure-port
	servicePort = 9992
	// healthCheckEndpoint is the default healthcheck endpoint for all services.
	healthCheckEndpoint = "/-/healthz"
)

// Default values.
var (
	// defaultMaxInstances is the default Scaling.MaxCount
	defaultMaxInstances = 5
)

// makeServiceEnvVarPrefix returns the env var prefix for service-specific
// env vars that will be set on the Cloud Run service, i.e.
//
// - ${local.env_var_prefix}_BIGQUERY_PROJECT_ID
// - ${local.env_var_prefix}_BIGQUERY_DATASET
// - ${local.env_var_prefix}_BIGQUERY_TABLE
// - ${local.env_var_prefix}_DIAGNOSTICS_SECRET
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
// Note that some variables like GOOGLE_PROJECT_ID and REDIS_ENDPOINT do not
// get prefixed, and custom env vars configured on an environment are not prefixed
// either.
func makeServiceEnvVarPrefix(serviceID string) string {
	return strings.ToUpper(strings.ReplaceAll(serviceID, "-", "_")) + "_"
}

// NewStack instantiates the MSP cloudrun stack, which is currently a pretty
// monolithic stack that encompasses all the core components of an MSP service,
// including networking and dependencies like Redis.
func NewStack(stacks *stack.Set, vars Variables) (*Output, error) {
	stack := stacks.New(StackName,
		googleprovider.With(vars.Project),
		randomprovider.With())

	// TODO
	serviceImageTag, err := vars.Environment.Deploy.ResolveTag()
	if err != nil {
		return nil, err
	}

	// Set up service account for the Cloud Run instance
	serviceAccount, err := serviceaccount.New(stack, "cloudrun-sa", serviceaccount.Config{
		Project:   vars.Project,
		AccountID: vars.Service.ID,
		DisplayName: fmt.Sprintf("%s Service Account",
			pointer.IfNil(vars.Service.Name, vars.Service.ID)),
		Roles: serviceAccountRoles,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to render Cloud Run service account")
	}

	// Set up networking for the Cloud Run service
	networking := newCloudRunNetwork(stack, cloudRunNetworkConfig{
		ProjectID: *vars.Project.Id(),
		ServiceID: vars.Service.ID,
		Region:    gcpRegion,
	})

	// redisInstance is only created and non-nil if Redis is configured for the
	// environment.
	var redisInstance *redis.Output
	if vars.Environment.Resources.Redis != nil {
		redisInstance, err = redis.New(stack, "redis", redis.Config{
			Project: vars.Project,
			Network: networking.network,
			Region:  gcpRegion,
			Spec:    *vars.Environment.Resources.Redis,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to render Redis instance")
		}
	}

	// bigqueryDataset is only created and non-nil if BigQuery is configured for
	// the environment.
	var bigqueryDataset *bigquery.Output
	if vars.Environment.Resources.BigQueryTable != nil {
		bigqueryDataset, err = bigquery.New(stack, "bigquery", bigquery.Config{
			DefaultProject: vars.Project,
			Spec:           *vars.Environment.Resources.BigQueryTable,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to render BigQuery dataset")
		}
	}

	// Create secret for healthcheck endpoints
	diagnosticsSecret := random.New(stack, "diagnostics-secret", random.Config{
		ByteLength: 8,
	})

	// Set up the core Cloud Run service
	_ = cloudrunv2service.NewCloudRunV2Service(stack, pointer.Value("cloudrun"),
		&cloudrunv2service.CloudRunV2ServiceConfig{
			Name:     pointer.Value(vars.Service.ID),
			Location: pointer.Value(gcpRegion),
			//  Disallows direct traffic from public internet, we have a LB set up for that.
			Ingress: pointer.Value("INGRESS_TRAFFIC_INTERNAL_LOAD_BALANCER"),

			Template: &cloudrunv2service.CloudRunV2ServiceTemplate{
				ServiceAccount: pointer.Value(serviceAccount.Email),

				// So set a high limit that matches our default Cloudflare zone's
				// timeout:
				//
				//   export CF_API_TOKEN=$(gcloud secrets versions access latest --secret CLOUDFLARE_API_TOKEN --project sourcegraph-secrets)
				//   curl -H "Authorization: Bearer $CF_API_TOKEN" https://api.cloudflare.com/client/v4/zones | jq '.result[]  | select(.name == "sourcegraph.com") | .id'
				//   curl -H "Authorization: Bearer $CF_API_TOKEN" https://api.cloudflare.com/client/v4/zones/$CF_ZONE_ID/settings | jq '.result[] | select(.id == "proxy_read_timeout")'
				//
				// Result should be something like:
				//
				//   {
				//     "id": "proxy_read_timeout",
				//     "value": "300",
				//     "modified_on": "2022-02-08T23:10:35.772888Z",
				//     "editable": true
				//   }
				//
				// The service should implement tighter timeouts on its own if desired.
				Timeout: pointer.Value("300s"),

				MaxInstanceRequestConcurrency: pointer.Float64(vars.Environment.Instances.Scaling.MaxRequestConcurrency),
				Scaling: &cloudrunv2service.CloudRunV2ServiceTemplateScaling{
					MinInstanceCount: pointer.Float64(vars.Environment.Instances.Scaling.MinCount),
					MaxInstanceCount: pointer.Float64(
						pointer.IfNil(vars.Environment.Instances.Scaling.MaxCount, defaultMaxInstances)),
				},

				Containers: []*cloudrunv2service.CloudRunV2ServiceTemplateContainers{{
					Name:  pointer.Value(vars.Service.ID),
					Image: pointer.Value(fmt.Sprintf("%s:%s", vars.Image, serviceImageTag)),

					Resources: &cloudrunv2service.CloudRunV2ServiceTemplateContainersResources{
						Limits: makeContainerResourceLimits(vars.Environment.Instances.Resources),
					},

					Ports: []*cloudrunv2service.CloudRunV2ServiceTemplateContainersPorts{{
						// ContainerPort is provided to the container as $PORT in Cloud Run
						ContainerPort: pointer.Float64(servicePort),
					}},

					Env: makeContainerEnvVars(
						vars.Project,
						makeServiceEnvVarPrefix(vars.Service.ID),
						diagnosticsSecret.HexValue,
						vars.Environment.Env,
						vars.Environment.SecretEnv,
						// Additional optional components
						redisInstance,
						bigqueryDataset,
					),

					// Do healtchecks with authorization based on MSP convention.
					StartupProbe: &cloudrunv2service.CloudRunV2ServiceTemplateContainersStartupProbe{
						HttpGet: &cloudrunv2service.CloudRunV2ServiceTemplateContainersStartupProbeHttpGet{
							Path: pointer.Value(healthCheckEndpoint),
							HttpHeaders: []*cloudrunv2service.CloudRunV2ServiceTemplateContainersStartupProbeHttpGetHttpHeaders{{
								Name:  pointer.Value("Bearer"),
								Value: pointer.Value(fmt.Sprintf("Authorization %s", diagnosticsSecret.HexValue)), // TODO
							}},
						},
					},

					// LivenessProbe: &cloudrunv2service.CloudRunV2ServiceTemplateContainersLivenessProbe{
					// 	// TODO
					// },

					// VolumeMounts: &cloudrunv2service.CloudRunV2ServiceTemplateContainersVolumeMounts{
					// 	// TODO
					// },
				}},

				// Volumes: &cloudrunv2service.CloudRunV2ServiceTemplateVolumes{
				// 	// TODO
				// },

				// VpcAccess: &cloudrunv2service.CloudRunV2ServiceTemplateVpcAccess{
				// 	// TODO
				// },
			},
		})

	return &Output{}, nil
}
