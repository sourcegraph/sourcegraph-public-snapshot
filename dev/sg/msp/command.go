// Package msp exports the 'sg msp' command for the Managed Services Platform.
package msp

import (
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/category"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
)

// Command is currently only implemented with the 'msp' build tag - see sg_msp.go
//
// The default implementation is hidden by default and offers some help text for
// for installing 'sg' with 'sg msp' enabled.
var Command = &cli.Command{
	Hidden:      true,
	Name:        "managed-services-platform",
	Aliases:     []string{"msp"},
	Usage:       "Generate and manage services deployed on the Sourcegraph Managed Services Platform",
	Description: `WARNING: This is currently still an experimental project. To learm more, see go/rfc-msp`,
	Category:    category.Company,
	Action: func(c *cli.Context) error {
		std.Out.WriteWarningf("'sg msp' is not available in this build of 'sg'.")
		std.Out.Write("To install a build of 'sg' that includes 'sg msp', run:")
		if err := std.Out.WriteCode("bash",
			"go build -tags=msp -o=./sg ./dev/sg && ./sg install -f -p=false"); err != nil {
			return err
		}
		return errors.New("command unimplemented")
	},
	Subcommands: nil,
}
