package tfcbackend

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type Config struct {
	Workspace func(stackName string) string `validate:"required"`
}

const metadataKey = "tfc-workspace"

// With configures the stack to use Terraform Cloud as remote state backend.
// Any top-level CDKTF stack should use this as remote state backend.
func With(config Config) stack.NewStackOption {
	return func(s stack.Stack) error {
		workspace := config.Workspace(s.Name)
		_ = cdktf.NewCloudBackend(s.Stack, &cdktf.CloudBackendConfig{
			Hostname:     pointers.Ptr("app.terraform.io"),
			Organization: pointers.Ptr("sourcegraph"),
			Workspaces:   cdktf.NewNamedCloudWorkspace(&workspace),
		})
		s.Metadata[metadataKey] = workspace
		return nil
	}
}
