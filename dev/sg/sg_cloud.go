package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/run"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/category"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/cloud"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var cloudCommand = &cli.Command{
	Name:  "cloud",
	Usage: "Install and work with Sourcegraph Cloud tools",
	Description: `Learn more about Sourcegraph Cloud:

- Product: https://sourcegraph.com/docs/cloud
- Handbook: https://handbook.sourcegraph.com/departments/cloud/
`,
	Category: category.Company,
	Subcommands: []*cli.Command{
		{
			Name:        "install",
			Usage:       "Install or upgrade local `mi2` CLI (for Cloud V2)",
			Description: "To learn more about Cloud V2, see https://handbook.sourcegraph.com/departments/cloud/technical-docs/v2.0/",
			Action: func(c *cli.Context) error {
				if err := installCloudCLI(c.Context); err != nil {
					return err
				}
				if err := checkGKEAuthPlugin(); err != nil {
					return err
				}
				return nil
			},
		},
		&cloud.EphemeralCommand,
	},
}

func checkGKEAuthPlugin() error {
	const executable = "gke-gcloud-auth-plugin"
	existingPath, err := exec.LookPath(executable)
	if err != nil {
		return errors.Wrapf(err, "gke-gcloud-auth-plugin not found on path, run `brew info google-cloud-sdk` for instructions OR \n"+
			"run `gcloud components install gke-gcloud-auth-plugin` to install it manually")
	}
	std.Out.WriteNoticef("Using gcloud auth plugin at %q", existingPath)
	return nil
}

func installCloudCLI(ctx context.Context) error {
	const executable = "mi2"

	// Ensure gh is installed
	ghPath, err := exec.LookPath("gh")
	if err != nil {
		return errors.Wrap(err, "GitHub CLI (https://cli.github.com/) is required for installation")
	}
	std.Out.Writef("Using GitHub CLI at %q", ghPath)

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
		if !strings.HasPrefix(existingPath, locationDir) {
			std.Out.WriteNoticef("Removing existing installation at of %q at %q",
				executable, existingPath)
			_ = os.Remove(existingPath)
		}
	}

	version, err := run.Cmd(ctx, ghPath, "version").Run().String()
	if err != nil {
		return errors.Wrap(err, "get gh version")
	}
	std.Out.WriteNoticef("Using GitHub CLI version %q", strings.Split(version, "\n")[0])

	start := time.Now()
	pending := std.Out.Pending(output.Styledf(output.StylePending, "Downloading %q to %q... (hang tight, this might take a while!)",
		executable, locationDir))

	const tempExecutable = "mi2_tmp"
	tempInstallPath := filepath.Join(locationDir, tempExecutable)
	finalInstallPath := filepath.Join(locationDir, executable)
	_ = os.Remove(tempInstallPath)
	// Get release
	if err := run.Cmd(ctx,
		ghPath, " release download -R github.com/sourcegraph/controller",
		"--pattern", fmt.Sprintf("mi2_%s_%s", runtime.GOOS, runtime.GOARCH),
		"--output", tempInstallPath).
		Run().Wait(); err != nil {
		pending.Close()
		return errors.Wrap(err, "download mi2")
	}
	pending.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess,
		"Download complete! (elapsed: %s)",
		time.Since(start).String()))

	// Move binary to final destination
	if err := os.Rename(tempInstallPath, finalInstallPath); err != nil {
		return errors.Wrap(err, "move mi2 to final path")
	}

	// Make binary executable
	if err := os.Chmod(finalInstallPath, 0755); err != nil {
		return errors.Wrap(err, "make mi2 executable")
	}

	std.Out.WriteSuccessf("%q successfully installed!", executable)
	return nil
}
