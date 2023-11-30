// Package msp exports the 'sg msp' command for the Managed Services Platform.
package msp

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/category"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
)

const commandDescription = `WARNING: This is currently still an experimental project.
To learm more, refer to go/rfc-msp and go/msp (https://handbook.sourcegraph.com/departments/engineering/teams/core-services/managed-services/platform)`

const buildCommand = "go build -tags=msp -o=./sg ./dev/sg && ./sg install -f -p=false"

// Command is currently only implemented with the 'msp' build tag - see sg_msp.go
//
// The default implementation is hidden by default and offers some help text for
// for installing 'sg' with 'sg msp' enabled.
var Command = &cli.Command{
	Name:    "managed-services-platform",
	Aliases: []string{"msp"},
	Usage:   "EXPERIMENTAL: Generate and manage services deployed on the Sourcegraph Managed Services Platform",
	Description: fmt.Sprintf(`%s

MSP commands are currently build-flagged to avoid increasing 'sg' binary sizes. To install a build of 'sg' that includes 'sg msp', run:

	%s

MSP commands should then be available under 'sg msp --help'.`, commandDescription, buildCommand),
	UsageText: `
# Create a service specification
sg msp init $SERVICE

# Provision Terraform Cloud workspaces
sg msp tfc sync $SERVICE $ENVIRONMENT

# Generate Terraform manifests
sg msp generate $SERVICE $ENVIRONMENT
`,
	Category: category.Company,
	Action: func(c *cli.Context) error {
		std.Out.WriteWarningf("'sg msp' is not available in this build of 'sg'.")
		std.Out.Write("To install a build of 'sg' that includes 'sg msp', run:")
		if err := std.Out.WriteCode("bash", buildCommand); err != nil {
			return err
		}
		return errors.New("command unimplemented")
	},
	Subcommands: nil,
}
