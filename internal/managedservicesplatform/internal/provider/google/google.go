package google

import (
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/project"
	googleprovider "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/provider"

	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/internal/stack"
	"github.com/sourcegraph/sourcegraph/internal/pointer"
)

// StackWithProject modifies a new stack to use the Google Terraform provider
// with the given project.
func StackWithProject(p project.Project) stack.NewStackOption {
	return func(s stack.Stack) {
		_ = googleprovider.NewGoogleProvider(s.Stack, pointer.Value("google"), &googleprovider.GoogleProviderConfig{
			Project: p.ProjectId(),
		})
	}
}

// StackWithProjectID modifies a new stack to use the Google Terraform provider
// with the given project ID.
func StackWithProjectID(projectID string) stack.NewStackOption {
	return func(s stack.Stack) {
		_ = googleprovider.NewGoogleProvider(s.Stack, pointer.Value("google"), &googleprovider.GoogleProviderConfig{
			Project: pointer.Value(projectID),
		})
	}
}
