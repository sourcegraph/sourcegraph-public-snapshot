package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/usershell"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// additional sg configuration setup is the same on all platforms.
var dependencyCategoryAdditionalSgConfiguration = dependencyCategory{
	name:       "Additional sg configuration",
	autoFixing: true,
	dependencies: []*dependency{
		{
			name: "autocompletions",
			check: func(ctx context.Context) error {
				if !usershell.IsSupportedShell(ctx) {
					return nil // dont do setup
				}
				sgHome, err := root.GetSGHomePath()
				if err != nil {
					return err
				}
				shell := usershell.ShellType(ctx)
				autocompletePath := usershell.AutocompleteScriptPath(sgHome, shell)
				if _, err := os.Stat(autocompletePath); err != nil {
					return errors.Wrapf(err, "autocomplete script for shell %s not found", shell)
				}

				shellConfig := usershell.ShellConfigPath(ctx)
				conf, err := os.ReadFile(shellConfig)
				if err != nil {
					return err
				}
				if !strings.Contains(string(conf), autocompletePath) {
					return errors.Newf("autocomplete script %s not found in shell config %s",
						autocompletePath, shellConfig)
				}
				return nil
			},
			instructionsCommandsBuilder: stringCommandBuilder(func(ctx context.Context) string {
				sgHome, err := root.GetSGHomePath()
				if err != nil {
					return fmt.Sprintf("echo %s && exit 1", err.Error())
				}

				var commands []string

				shell := usershell.ShellType(ctx)
				if shell == "" {
					return "echo 'Failed to detect shell type' && exit 1"
				}
				autocompleteScript := usershell.AutocompleteScripts[shell]
				autocompletePath := usershell.AutocompleteScriptPath(sgHome, shell)
				commands = append(commands,
					fmt.Sprintf(`echo "Writing autocomplete script to %s"`, autocompletePath),
					fmt.Sprintf(`echo '%s' > %s`, autocompleteScript, autocompletePath))

				shellConfig := usershell.ShellConfigPath(ctx)
				if shellConfig == "" {
					return "echo 'Failed to detect shell config path' && exit 1"
				}
				conf, err := os.ReadFile(shellConfig)
				if err != nil {
					return fmt.Sprintf("echo %s && exit 1", err.Error())
				}
				if !strings.Contains(string(conf), autocompletePath) {
					commands = append(commands,
						fmt.Sprintf(`echo "Adding configuration to %s"`, shellConfig),
						fmt.Sprintf(`echo "PROG=sg source %s" >> %s`,
							autocompletePath, shellConfig))
				}

				return strings.Join(commands, "\n")
			}),
		},
	},
}

// gcloud setup is the same on all platforms.
var dependencyGcloud = &dependency{
	name: "gcloud",
	check: check.Combine(
		check.InPath("gcloud"),
		// User should have logged in with a sourcegraph.com account
		check.CommandOutputContains("gcloud auth list", "@sourcegraph.com")),
	instructionsCommandsBuilder: stringCommandBuilder(func(ctx context.Context) string {
		var commands []string
		if err := check.InPath("gcloud")(ctx); err != nil {
			// This is the official interactive installer: https://cloud.google.com/sdk/docs/downloads-interactive
			commands = append(commands, "curl https://sdk.cloud.google.com | bash")
		}
		commands = append(commands,
			"gcloud auth login",
			"gcloud auth configure-docker")

		return strings.Join(commands, "\n")
	}),
	onlyTeammates:       true,
	instructionsComment: "NOTE: You can ignore this if you're not a Sourcegraph teammate.",
}

// check1password defines the 1password dependency check which is uniform across platforms.
func check1password() check.CheckFunc {
	return check.Combine(
		check.WrapErrMessage(check.InPath("op"), "The 1password CLI, 'op', is required"),
		check.CommandOutputContains("op account list", "team-sourcegraph.1password.com"))
}
