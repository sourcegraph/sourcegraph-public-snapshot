package tfcbackend

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"

	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/internal/stack"
	"github.com/sourcegraph/sourcegraph/internal/pointer"
)

type Config struct {
	Workspace func(stackName string) string `validate:"required"`
}

// New configures the stack to use Terraform Cloud as remote state backend.
// Any top-level CDKTF stack should use this as remote state backend.
func WithBackend(config Config) stack.NewStackOption {
	return func(s stack.Stack) {
		workspace := config.Workspace(s.Name)
		_ = cdktf.NewCloudBackend(s.Stack, &cdktf.CloudBackendConfig{
			Hostname:     pointer.Value("app.terraform.io"),
			Organization: pointer.Value("sourcegraph"),
			Workspaces:   cdktf.NewNamedCloudWorkspace(&workspace),
		})
	}
}
