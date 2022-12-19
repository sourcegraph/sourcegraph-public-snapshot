package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/sourcegraph/run"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var cloudCommand = &cli.Command{
	Name:  "cloud",
	Usage: "Install and work with Sourcegraph Cloud tools",
	Description: `Learn more about Sourcegraph Cloud:

- Product: https://docs.sourcegraph.com/cloud
- Handbook: https://handbook.sourcegraph.com/departments/cloud/
`,
	Category: CategoryCompany,
	Subcommands: []*cli.Command{
		{
			Name:        "install",
			Usage:       "Install or upgrade local `mi2` CLI (for Cloud V2)",
			Description: "To learn more about Cloud V2, see https://handbook.sourcegraph.com/departments/cloud/technical-docs/v2.0/",
			Action: func(c *cli.Context) error {
				const executable = "mi2"

				// Use the same directory as sg, since we add that to path
				homeDir, err := os.UserHomeDir()
				if err != nil {
					return err
				}
				locationDir, err := sgInstallDir(homeDir)
				if err != nil {
					return err
				}

				// Remove existing install if there is one
				if existingPath, err := exec.LookPath(executable); err == nil {
					// If this mi2 installation is installed elsewhere, remove it to
					// avoid conflicts
					if !filepath.HasPrefix(existingPath, locationDir) {
						std.Out.WriteNoticef("Removing existing installation at of %q at %q",
							executable, existingPath)
						_ = os.Remove(existingPath)
					}
				}

				// Ensure gh is installed
				if _, err := exec.LookPath("gh"); err != nil {
					return errors.Wrap(err, "GitHub CLI (https://cli.github.com/) is required for installation")
				}

				start := time.Now()
				pending := std.Out.Pending(output.Styledf(output.StylePending, "Downloading %q to %q... (hang tight, this might take a while!)",
					executable, locationDir))

				// Get release
				const tempExecutable = "mi2_tmp"
				_ = os.Remove(filepath.Join(locationDir, tempExecutable))
				if err := run.Cmd(c.Context,
					"gh release download -R github.com/sourcegraph/controller",
					"--pattern", fmt.Sprintf("mi2_%s_%s", runtime.GOOS, runtime.GOARCH),
					"--output", tempExecutable).
					Dir(locationDir).
					Run().Wait(); err != nil {
					pending.Close()
					return errors.Wrap(err, "download mi2")
				}
				pending.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess,
					"Download complete! (elapsed: %s)",
					time.Since(start).String()))

				// Move binary to final destination
				if err := run.Cmd(c.Context, "mv", tempExecutable, executable).
					Dir(locationDir).
					Run().Wait(); err != nil {
					return errors.Wrap(err, "move mi2 to final path")
				}

				// Make binary executable
				if err := run.Cmd(c.Context, "chmod +x", executable).
					Dir(locationDir).
					Run().Wait(); err != nil {
					return errors.Wrap(err, "make mi2 executable")
				}

				std.Out.WriteSuccessf("%q successfully installed!", executable)
				return nil
			},
		},
	},
}
