package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/urfave/cli/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/category"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
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
		&cli.PathFlag{
			Name:    "outputFile",
			Aliases: []string{"o"},
			Value:   fmt.Sprintf("sg-doctor-report-%s.md", time.Now().Format("2006-01-02-150405")),
			Usage:   "write the report to a file with this name",
		},
	},
	Action: runDoctorDiagnostics,
}

type Diagnostic struct {
	Name string `yaml:"name"`
	Cmd  string `yaml:"cmd"`
}

type Diagnostics struct {
	Diagnostic map[string][]Diagnostic `yaml:"diagnostics"`
}

type diagnosticRunner struct {
	diagnostics *Diagnostics
	reporter    *std.Output
}

type DiagnosticResult struct {
	Diagnostic *Diagnostic
	Output     string
	Err        error
}
type DiagnosticReport map[string][]*DiagnosticResult

func (r DiagnosticReport) Add(group string, result *DiagnosticResult) {
	if v, ok := r[group]; !ok {
		r[group] = []*DiagnosticResult{result}
	} else {
		r[group] = append(v, result)
	}
}

func runDoctorDiagnostics(cmd *cli.Context) error {
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return err
	}
	diagnosticsPath := filepath.Join(repoRoot, "sg-doctor.yaml")
	diagnostics, err := loadDiagnostics(diagnosticsPath)
	if err != nil {
		return errors.Newf("failed to load diagnostics from %q: %v", diagnosticsPath, err)
	}

	runner := &diagnosticRunner{
		diagnostics,
		std.Out,
	}
	report := runner.Run(cmd.Context)
	err = writeDiagnosticReport(report, cmd.Path("outputFile"))
	if err != nil {
		return errors.Wrap(err, "failed to write diagnostic report")
	}

	std.Out.WriteLine(output.Emoji("📋", "Diagnostic report written to "+cmd.Path("outputFile")))

	return nil
}

func (d *diagnosticRunner) Run(ctx context.Context) DiagnosticReport {
	env := os.Environ()
	report := make(DiagnosticReport)

	d.reporter.WriteLine(output.Emoji("🥼", "Gathering diagnostics"))
	for group, diagnostics := range d.diagnostics.Diagnostic {
		blk := d.reporter.Block(output.Emojif("💊", "Running %s diagnostics", group))
		for _, diagnostic := range diagnostics {
			pending := blk.Pending(output.Styledf(output.StylePending, "Running %q", diagnostic.Name))
			out, err := run.BashInRoot(ctx, diagnostic.Cmd, env)
			diag := diagnostic
			report.Add(group, &DiagnosticResult{
				&diag,
				out,
				err,
			})
			pending.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Diagnostic %q complete", diagnostic.Name))
		}
		blk.Close()
	}

	d.reporter.WriteLine(output.Emoji("💉", "Gathering of diagnostics complete!"))

	return report
}

func writeDiagnosticReport(report DiagnosticReport, dst string) error {
	fd, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer fd.Close()

	fmt.Fprintf(fd, "# Diagnostic Report\n\n")
	// General information
	fmt.Fprintf(fd, "sg commit: %s\n", BuildCommit)
	fmt.Fprintf(fd, "generated on: %s\n\n", time.Now())
	// Write out the report
	titleCaser := cases.Title(language.English)
	for group, result := range report {
		fmt.Fprintf(fd, "## %s diagnostics\n\n", titleCaser.String(group))
		for _, item := range result {
			cmdLine := fmt.Sprintf("Command: `%s`", item.Diagnostic.Cmd)
			outputSection := fmt.Sprintf("Output: \n```\n%s\n```\n", item.Output)
			errSection := fmt.Sprintf("Error: \n```\n%v\n```\n", item.Err)
			fmt.Fprintf(fd, "### %s\n\n%s\n\n%s\n%s", titleCaser.String(item.Diagnostic.Name), cmdLine, outputSection, errSection)
		}
	}

	return nil
}

func loadDiagnostics(path string) (*Diagnostics, error) {
	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	var diags Diagnostics
	dec := yaml.NewDecoder(fd)

	err = dec.Decode(&diags)
	if err != nil {
		return nil, err
	}

	return &diags, nil
}
