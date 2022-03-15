package main

import (
	"context"
	"flag"
	"os/exec"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/secrets"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var (
	updateFlags   = flag.NewFlagSet("sg update", flag.ExitOnError)
	updateToLocal = updateFlags.Bool("local", false, "Update to local copy of 'dev/sg'")
)

// autoUpdateSetting struct to store auto update setting in the secrets.
type autoUpdateSetting struct {
	Auto bool `json:"auto"`
}

// autoUpdateSecretsKey is the key under which this setting is filed in the secrets
var autoUpdateSecretsKey = "auto-update"

var updateCommand = &ffcli.Command{
	Name:       "update",
	FlagSet:    updateFlags,
	ShortUsage: "sg update",
	ShortHelp:  "Update sg.",
	LongHelp: `Update local sg installation with the latest changes. To see what's new, run:

  sg version changelog -next

Requires a local copy of the 'sourcegraph/sourcegraph' codebase.`,
	Subcommands: []*ffcli.Command{{
		Name:       "auto",
		ShortUsage: "sg update auto [enable|disable]",
		ShortHelp:  "Enable or disable sg auto-updates",
		Exec: func(ctx context.Context, args []string) error {
			if len(args) < 1 || (args[0] != "disable" && args[0] != "enable") {
				return flag.ErrHelp
			}
			sec := secrets.FromContext(ctx)
			if args[0] == "enable" {
				return sec.PutAndSave(autoUpdateSecretsKey, autoUpdateSetting{Auto: true})
			}
			return sec.PutAndSave(autoUpdateSecretsKey, autoUpdateSetting{Auto: false})
		},
	}},
	Exec: func(ctx context.Context, args []string) error {
		if *updateToLocal {
			stdout.Out.WriteLine(output.Line(output.EmojiHourglass, output.StylePending, "Upgrading to local copy of 'dev/sg'..."))
		} else {
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
		}

		// Run installation script
		cmd := exec.CommandContext(ctx, "./dev/sg/install.sh")
		if err := run.InteractiveInRoot(cmd); err != nil {
			return err
		}
		writeSuccessLinef("Update succeeded!")
		return nil
	},
}

func getAutoUpdateSetting(ctx context.Context) bool {
	sec := secrets.FromContext(ctx)
	var setting autoUpdateSetting
	_ = sec.Get(autoUpdateSecretsKey, &setting)
	return setting.Auto
}
