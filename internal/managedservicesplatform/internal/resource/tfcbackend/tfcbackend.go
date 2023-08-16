package tfcbackend

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type Config struct {
	Workspace string `validate:"required"`
}

// New configures the stack to use Terraform Cloud as remote state backend.
// Any top-level CDKTF stack should use this as remote state backend.
func New(scope constructs.Construct, config Config) cdktf.CloudBackend {
	return cdktf.NewCloudBackend(scope, &cdktf.CloudBackendConfig{
		Hostname:     jsii.String("app.terraform.io"),
		Organization: jsii.String("sourcegraph"),
		Workspaces:   cdktf.NewNamedCloudWorkspace(&config.Workspace),
	})
}
