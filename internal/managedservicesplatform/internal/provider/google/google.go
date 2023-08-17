package google

import (
	"github.com/aws/jsii-runtime-go"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/project"
	googleprovider "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/provider"

	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/internal/stack"
)

// StackWithProject modifies a new stack to use the Google Terraform provider
// with the given project.
func StackWithProject(p project.Project) stack.NewStackOption {
	return func(s stack.Stack) {
		_ = googleprovider.NewGoogleProvider(s.Stack, jsii.String("google"), &googleprovider.GoogleProviderConfig{
			Project: p.ProjectId(),
		})
	}
}

// StackWithProjectID modifies a new stack to use the Google Terraform provider
// with the given project ID.
func StackWithProjectID(projectID string) stack.NewStackOption {
	return func(s stack.Stack) {
		_ = googleprovider.NewGoogleProvider(s.Stack, jsii.String("google"), &googleprovider.GoogleProviderConfig{
			Project: jsii.String(projectID),
		})
	}
}
