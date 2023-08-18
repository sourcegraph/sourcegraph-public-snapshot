package googleprovider

import (
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/project"
	google "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/provider"

	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/internal/stack"
	"github.com/sourcegraph/sourcegraph/internal/pointer"
)

// With modifies a new stack to use the Google Terraform provider with the given
// project.
func With(p project.Project) stack.NewStackOption {
	return func(s stack.Stack) {
		_ = google.NewGoogleProvider(s.Stack, pointer.Value("google"), &google.GoogleProviderConfig{
			Project: p.ProjectId(),
		})
	}
}

// WithProjectID modifies a new stack to use the Google Terraform provider
// with the given project ID.
func WithProjectID(projectID string) stack.NewStackOption {
	return func(s stack.Stack) {
		_ = google.NewGoogleProvider(s.Stack, pointer.Value("google"), &google.GoogleProviderConfig{
			Project: pointer.Value(projectID),
		})
	}
}
