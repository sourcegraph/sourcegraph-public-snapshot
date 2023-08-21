package cloudflare

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/project"

	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/spec"
)

type Output struct {
}

type Config struct {
	Project project.Project

	Spec spec.EnvironmentDomainCloudflareSpec
}

func New(scope constructs.Construct, id string, config Config) (*Output, error) {

	// TODO

	return &Output{}, nil
}
