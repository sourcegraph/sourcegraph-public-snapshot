package cloudrun

import (
	"strconv"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"

	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/cloudrunv2service"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func makeContainerResourceLimits(r spec.EnvironmentInstancesResourcesSpec) *map[string]*string {
	return &map[string]*string{
		"cpu":    pointers.Ptr(strconv.Itoa(r.CPU)),
		"memory": pointers.Ptr(r.Memory),
	}
}

func makeContainerEnvVars(
	env map[string]string,
	secretEnv map[string]string,
) []*cloudrunv2service.CloudRunV2ServiceTemplateContainersEnv {
	// We configure some base env vars for all services
	var vars []*cloudrunv2service.CloudRunV2ServiceTemplateContainersEnv

	// Apply static env vars
	envKeys := maps.Keys(env)
	slices.Sort(envKeys)
	for _, k := range envKeys {
		vars = append(vars, &cloudrunv2service.CloudRunV2ServiceTemplateContainersEnv{
			Name:  pointers.Ptr(k),
			Value: pointers.Ptr(env[k]),
		})
	}

	// Apply secret env vars
	secretEnvKeys := maps.Keys(secretEnv)
	slices.Sort(secretEnvKeys)
	for _, k := range secretEnvKeys {
		vars = append(vars, &cloudrunv2service.CloudRunV2ServiceTemplateContainersEnv{
			Name: pointers.Ptr(k),
			ValueSource: &cloudrunv2service.CloudRunV2ServiceTemplateContainersEnvValueSource{
				SecretKeyRef: &cloudrunv2service.CloudRunV2ServiceTemplateContainersEnvValueSourceSecretKeyRef{
					Secret: pointers.Ptr(secretEnv[k]),
				},
			},
		})
	}

	return vars
}
