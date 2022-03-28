package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var (
	updateFlags    = flag.NewFlagSet("sg update", flag.ExitOnError)
	updateToLocal  = updateFlags.Bool("local", false, "Update to local copy of 'dev/sg'")
	updateDownload = updateFlags.Bool("download", false, "Download a prebuilt binary of 'sg' instead of compiling it locally")
)

var updateCommand = &ffcli.Command{
	Name:       "update",
	FlagSet:    updateFlags,
	ShortUsage: "sg update",
	ShortHelp:  "Update sg.",
	LongHelp: `Update local sg installation with the latest changes. To see what's new, run:

  sg version changelog -next

Requires a local copy of the 'sourcegraph/sourcegraph' codebase.`,
	Exec: func(ctx context.Context, args []string) error {
		if *updateDownload {
			return updateToPrebuiltSG(ctx)
		}
		if *updateToLocal {
			return updateToLocalSG(ctx)
		}
		return updateSG(ctx)
	},
}

// updateSG fetches the latest sg changes on the main branch, build them and install the new sg version.
func updateSG(ctx context.Context) error {
	// Update from remote
	if _, err := run.GitCmd("fetch", "origin", "main"); err != nil {
		return err
	}

	// Make sure to switch back to previous working state
	var restoreFuncs []func() error
	defer func() {
		stdout.Out.WriteLine(output.Line(output.EmojiHourglass, output.StyleSuggestion, "Restoring workspace..."))
		var failed bool
		for i := len(restoreFuncs) - 1; i >= 0; i-- {
			if restoreErr := restoreFuncs[i](); restoreErr != nil {
				failed = true
				writeWarningLinef(restoreErr.Error())
			}
		}
		if !failed {
			stdout.Out.WriteLine(output.Line(output.EmojiInfo, output.StyleSuggestion, "Workspace restored!"))
		} else {
			writeFailureLinef("Failed to restore workspace")
		}
	}()

	// Stash workspace if dirty
	changes, err := run.TrimResult(run.GitCmd("status", "--porcelain"))
	if err != nil {
		return err
	}
	if len(changes) > 0 {
		stdout.Out.WriteLine(output.Line(output.EmojiHourglass, output.StyleSuggestion, "Stashing workspace..."))
		if _, err := run.GitCmd("stash", "--include-untracked"); err != nil {
			return err
		}
		restoreFuncs = append(restoreFuncs, func() (restoreErr error) {
			_, restoreErr = run.GitCmd("stash", "pop")
			return
		})
	}

	// Checkout main, which we will install from
	stdout.Out.WriteLine(output.Line(output.EmojiHourglass, output.StyleSuggestion, "Setting workspace up for update..."))
	if _, err := run.GitCmd("checkout", "origin/main"); err != nil {
		return err
	}
	restoreFuncs = append(restoreFuncs, func() (restoreErr error) {
		_, restoreErr = run.GitCmd("switch", "-")
		return
	})

	// For info, show what sg revision we are upgrading to
	commit, err := run.TrimResult(run.GitCmd("rev-parse", "HEAD"))
	if err != nil {
		return err
	}
	stdout.Out.WriteLine(output.Linef(output.EmojiHourglass, output.StylePending, "Updating to sg@%s...", commit))

	// Run installation script
	cmd := exec.CommandContext(ctx, "./dev/sg/install.sh")
	if err := run.InteractiveInRoot(cmd); err != nil {
		return err
	}
	writeSuccessLinef("Update succeeded!")
	return nil
}

// updateToLocalSG builds the currently checkout sg code and install the resulting sg binary.
func updateToLocalSG(ctx context.Context) error {
	stdout.Out.WriteLine(output.Line(output.EmojiHourglass, output.StylePending, "Upgrading to local copy of 'dev/sg'..."))

	// Run installation script
	cmd := exec.CommandContext(ctx, "./dev/sg/install.sh")
	if err := run.InteractiveInRoot(cmd); err != nil {
		return err
	}
	writeSuccessLinef("Update succeeded!")
	return nil
}

// updateToPrebuiltSG downloads the latest release of sg prebuilt binaries and install it.
func updateToPrebuiltSG(ctx context.Context) error {
	req, err := http.NewRequest("GET", "https://github.com/sourcegraph/sg/releases/latest", nil)
	if err != nil {
		return err
	}
	// We use the RountTripper to make an HTTP request without having to deal
	// with redirections.
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	location := resp.Header.Get("location")
	if location == "" {
		return errors.New("GitHub latest release: empty location")
	}
	location = strings.ReplaceAll(location, "/tag/", "/download/")
	downloadURL := fmt.Sprintf("%s/sg_%s_%s", location, runtime.GOOS, runtime.GOARCH)

	tmpDir, err := os.MkdirTemp("", "sg")
	if err != nil {
		return err
	}
	defer func() {
		os.RemoveAll(tmpDir)
	}()

	resp, err = http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	tmpSgPath := tmpDir + "/sg"
	f, err := os.Create(tmpSgPath)
	if err != nil {
		return err
	}

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return err
	}
	err = os.Chmod(tmpSgPath, 0755)
	if err != nil {
		return err
	}

	// Reliable way to find where to put sg, even if the user is using
	// asdf.
	cmd := exec.CommandContext(ctx, "go", "list", "-f", "{{.Target}}")
	out, err := cmd.CombinedOutput()
	sgPath := strings.TrimSpace(string(out))
	if err != nil {
		return err
	}

	return os.Rename(tmpSgPath, sgPath)
}
