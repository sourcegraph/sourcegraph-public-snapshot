package googleprovider

import (
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/project"
	google "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/provider"

	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/internal/stack"
	"github.com/sourcegraph/sourcegraph/internal/pointer"
)

// With modifies a new stack to use the Google Terraform provider with the given
// project. Every stack using GCP resources should be created with this option.
//
// All GCP resources created under a stack with this option should still explicitly
// configure ProjectID individually.
func With(p project.Project) stack.NewStackOption {
	return func(s stack.Stack) {
		_ = google.NewGoogleProvider(s.Stack, pointer.Value("google"), &google.GoogleProviderConfig{
			Project: p.ProjectId(),
		})
	}
}

// WithProjectID modifies a new stack to use the Google Terraform provider
// with the given project ID. This should only be used if a project.Project is
// not yet available.
//
// All GCP resources created under a stack with this option should still explicitly
// configure ProjectID individually.
func WithProjectID(projectID string) stack.NewStackOption {
	return func(s stack.Stack) {
		_ = google.NewGoogleProvider(s.Stack, pointer.Value("google"), &google.GoogleProviderConfig{
			Project: pointer.Value(projectID),
		})
	}
}
