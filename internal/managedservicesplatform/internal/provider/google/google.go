package google

import (
	"github.com/aws/jsii-runtime-go"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/project"
	googleprovider "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/provider"

	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/internal/stack"
)

func WithProject(p project.Project) stack.NewStackOption {
	return func(s cdktf.TerraformStack) {
		_ = googleprovider.NewGoogleProvider(s, jsii.String("google"), &googleprovider.GoogleProviderConfig{
			Project: p.ProjectId(),
		})
	}
}

func WithProjectID(projectID string) stack.NewStackOption {
	return func(s cdktf.TerraformStack) {
		_ = googleprovider.NewGoogleProvider(s, jsii.String("google"), &googleprovider.GoogleProviderConfig{
			Project: jsii.String(projectID),
		})
	}
}
