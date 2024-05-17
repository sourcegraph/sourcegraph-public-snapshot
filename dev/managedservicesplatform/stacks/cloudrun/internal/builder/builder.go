package builder

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/privatenetwork"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/random"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/serviceaccount"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/stacks/cloudrun/cloudrunresource"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Variables struct {
	Service     spec.ServiceSpec
	Environment spec.EnvironmentSpec

	// Image and ResolvedImageTag are used to declare the full image reference
	// to deploy.
	Image    string
	ImageTag string
	// GCPProjectID for all resources.
	GCPProjectID string
	// GCPRegion for all resources.
	GCPRegion string
	// ServiceAccount for the Cloud Run resource
	ServiceAccount *serviceaccount.Output
	// DiagnosticsSecret is the secret for healthcheck or diagnostics endpoints
	DiagnosticsSecret *random.Output
	// PrivateNetwork is configured if required as an internal network for the
	// Cloud Run resource to talk to other GCP resources.
	PrivateNetwork *privatenetwork.Output
	// ResourceLimits is a map of resource limits for the Cloud Run resource.
	ResourceLimits map[string]*string
}

// Name returns the name to use for the Cloud Run resource.
func (v *Variables) Name() (string, error) {
	name := cloudrunresource.NewName(v.Service.ID, v.Environment.ID, v.GCPRegion)
	// Extra guard against long names, just in case - an apply to change the
	// name that fails during apply could cause extended downtime.
	if len(name) > 63 {
		return name, errors.Newf("evaluated Cloud Run name %q is too long, maximum length is 63 characters")
	}
	return name, nil
}

type SecretRef struct {
	Name    string
	Version string
}

type Resource interface {
	cdktf.TerraformResource

	// Name of the Cloud Run resource
	Name() *string
	// Location of the Cloud Run resource
	Location() *string
}

// Builder configures and creates Cloud Run Services or Jobs.
type Builder interface {
	// Kind indicates what this Builder implementation creates.
	Kind() spec.ServiceKind

	// AddEnv adds an environment variable to the resource, and should only be
	// used before Build.
	AddEnv(key, value string)
	// AddSecretEnv adds an environment variable to the resource, and should only
	// be used before Build.
	AddSecretEnv(key string, secret SecretRef)
	// AddVolumeMount adds a volume mount to the resource, and should only be
	// used before Build.
	AddVolumeMount(name, mountPath string)
	// AddVolumeMount adds a volume mount sourced from a secret to the resource,
	// and should only be used before Build.
	AddSecretVolume(name, mountPath string, secret SecretRef, mode int)
	// AddDependency adds an explicit resource dependency that must be available
	// before the Cloud Run resource is created.
	AddDependency(cdktf.ITerraformDependable)

	// Build finalizes the resource.
	Build(cdktf.TerraformStack, Variables) (Resource, error)
}

const (
	// ServicePort is provided to the container as $PORT in Cloud Run:
	// https://cloud.google.com/run/docs/configuring/services/containers#configure-port
	ServicePort = 9992
	// HealthCheckEndpoint is the default healthcheck endpoint for all services.
	HealthCheckEndpoint = "/-/healthz"
	// DefaultMaxInstances is the default Scaling.MaxCount
	DefaultMaxInstances = 5
	// DefaultMaxConcurrentRequests is the default Scaling.MaxRequestConcurrency
	// It is set very high to prefer fewer instances, as Go services can generally
	// handle very high load without issue.
	DefaultMaxConcurrentRequests = 1000
)
