package main

import (
	"path/filepath"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/category"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var doctorCommand = &cli.Command{
	Name:  "doctor",
	Usage: "performs various diagnostics and generates a report",
	Description: `Runs a series of commands defined in sg-doctor.yaml.

	The output of the commands are stored in a report, which can then be given to a dev-infra team memeber for
	further diagnosis.
	`,
	Category: category.Util,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "outputFile",
			Aliases: []string{"o"},
			Usage:   "write the report to a file with this name",
		},
	},
	Action: doctorAction,
}

type Diagnostic struct {
	Name string
	Cmd string
}

type Diagnostics struct
	Diagnostic map[string]diagnostic `yaml: ""`
}
func doctorAction(cmd *cli.Context) error {
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return err
	}
	diagnosticsPath := filepath.Join(repoRoot, "sg-doctor.yaml")
	diags, err := loadDiagnostics(diagnosticsPath)
	if err != nil {
		return errors.Newf("failed to load diagnostics from %q", diagnosticsPath)
	}
	return nil
}


