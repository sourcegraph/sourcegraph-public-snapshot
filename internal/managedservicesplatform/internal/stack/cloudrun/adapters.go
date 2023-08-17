package cloudrun

import (
	"strconv"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"

	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/cloudrunv2service"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/project"

	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/internal/resource/bigquery"
	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/internal/resource/redis"
	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/internal/pointer"
)

func makeContainerResourceLimits(r spec.EnvironmentInstancesResourcesSpec) *map[string]*string {
	return &map[string]*string{
		"cpu":    pointer.Value(strconv.Itoa(r.CPU)),
		"memory": pointer.Value(r.Memory),
	}
}

func makeContainerEnvVars(
	p project.Project,
	serviceEnvVarPrefix string,
	diagnosticsSecret string,
	env map[string]string,
	secretEnv map[string]string,

	// Optional resources
	redisInstance *redis.Output,
	bigqueryDataset *bigquery.Output,
) []*cloudrunv2service.CloudRunV2ServiceTemplateContainersEnv {
	vars := []*cloudrunv2service.CloudRunV2ServiceTemplateContainersEnv{
		{
			// Required to enable tracing etc.
			//
			// We don't use serviceEnvVarPrefix here because this is a
			// convention to indicate the environment's project.
			Name:  pointer.Value("GOOGLE_CLOUD_PROJECT"),
			Value: p.ProjectId(),
		},
		{
			Name:  pointer.Value(serviceEnvVarPrefix + "DIAGNOSTICS_SECRET"),
			Value: pointer.Value(diagnosticsSecret),
		},
	}
	if redisInstance != nil {
		vars = append(vars, &cloudrunv2service.CloudRunV2ServiceTemplateContainersEnv{
			// We don't use serviceEnvVarPrefix here because this is a
			// Sourcegraph-wide convention.
			Name:  pointer.Value("REDIS_ENDPOINT"),
			Value: pointer.Value(redisInstance.Endpoint),
		})
	}
	if bigqueryDataset != nil {
		vars = append(vars,
			&cloudrunv2service.CloudRunV2ServiceTemplateContainersEnv{
				Name:  pointer.Value(serviceEnvVarPrefix + "BIGQUERY_PROJECT_ID"),
				Value: pointer.Value(bigqueryDataset.ProjectID),
			}, &cloudrunv2service.CloudRunV2ServiceTemplateContainersEnv{
				Name:  pointer.Value(serviceEnvVarPrefix + "BIGQUERY_DATASET"),
				Value: pointer.Value(bigqueryDataset.Dataset),
			}, &cloudrunv2service.CloudRunV2ServiceTemplateContainersEnv{
				Name:  pointer.Value(serviceEnvVarPrefix + "BIGQUERY_TABLE"),
				Value: pointer.Value(bigqueryDataset.Table),
			})
	}

	envKeys := maps.Keys(env)
	slices.Sort(envKeys)
	for _, k := range envKeys {
		vars = append(vars, &cloudrunv2service.CloudRunV2ServiceTemplateContainersEnv{
			Name:  pointer.Value(k),
			Value: pointer.Value(env[k]),
		})
	}

	secretEnvKeys := maps.Keys(secretEnv)
	slices.Sort(secretEnvKeys)
	for _, k := range secretEnvKeys {
		vars = append(vars, &cloudrunv2service.CloudRunV2ServiceTemplateContainersEnv{
			Name: pointer.Value(k),
			ValueSource: &cloudrunv2service.CloudRunV2ServiceTemplateContainersEnvValueSource{
				SecretKeyRef: &cloudrunv2service.CloudRunV2ServiceTemplateContainersEnvValueSourceSecretKeyRef{
					Secret: pointer.Value(secretEnv[k]),
				},
			},
		})
	}

	return vars
}
