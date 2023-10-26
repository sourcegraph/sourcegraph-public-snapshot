package cloudrun

import (
	"bytes"
	"strconv"
	"text/template"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"

	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/cloudrunv2service"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func makeContainerResourceLimits(r spec.EnvironmentInstancesResourcesSpec) *map[string]*string {
	return &map[string]*string{
		"cpu":    pointers.Ptr(strconv.Itoa(r.CPU)),
		"memory": pointers.Ptr(r.Memory),
	}
}

type envVariablesData struct {
	ProjectID      string
	ServiceDnsName string
}

func makeContainerEnvVars(
	env map[string]string,
	secretEnv map[string]string,
	varsData envVariablesData,
) ([]*cloudrunv2service.CloudRunV2ServiceTemplateContainersEnv, error) {
	// We configure some base env vars for all services
	var vars []*cloudrunv2service.CloudRunV2ServiceTemplateContainersEnv

	// Apply static env vars
	envKeys := maps.Keys(env)
	slices.Sort(envKeys)
	for _, k := range envKeys {
		tmpl, err := template.New("").Parse(env[k])
		if err != nil {
			return nil, errors.Wrapf(err, "parse env var template: %q", env[k])
		}
		var buf bytes.Buffer
		if err = tmpl.Execute(&buf, varsData); err != nil {
			return nil, errors.Wrapf(err, "execute template: %q", env[k])
		}
		vars = append(vars, &cloudrunv2service.CloudRunV2ServiceTemplateContainersEnv{
			Name:  pointers.Ptr(k),
			Value: pointers.Ptr(buf.String()),
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
					Secret:  pointers.Ptr(secretEnv[k]),
					Version: pointers.Ptr("latest"),
				},
			},
		})
	}

	return vars, nil
}
