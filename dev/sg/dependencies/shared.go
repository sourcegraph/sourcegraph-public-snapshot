package dependencies

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/grafana/regexp"
	"github.com/sourcegraph/run"
	"go.bobheadxi.dev/streamline/pipeline"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/usershell"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func categoryCloneRepositories() category {
	return category{
		Name:      depsCloneRepo,
		DependsOn: []string{depsBaseUtilities},
		Checks: []*dependency{
			{
				Name: "SSH authentication with GitHub.com",
				Description: `Make sure that you can clone git repositories from GitHub via SSH.
See here on how to set that up:

https://docs.github.com/en/authentication/connecting-to-github-with-ssh`,
				Check: func(ctx context.Context, out *std.Output, args CheckArgs) error {
					return check.CommandOutputContains(
						"ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -T git@github.com",
						"successfully authenticated")(ctx)
				},
				// TODO we might be able to automate this fix
			},
			{
				Name:        "github.com/sourcegraph/sourcegraph",
				Description: `The 'sourcegraph' repository contains the Sourcegraph codebase and everything to run Sourcegraph locally.`,
				Check: func(ctx context.Context, out *std.Output, args CheckArgs) error {
					if _, err := root.RepositoryRoot(); err == nil {
						return nil
					}

					ok, err := pathExists("sourcegraph")
					if !ok || err != nil {
						return errors.New("'sg setup' is not run in sourcegraph and repository is also not found in current directory")
					}
					return nil
				},
				Fix: func(ctx context.Context, cio check.IO, args CheckArgs) error {
					cmd := run.Cmd(ctx, `git clone git@github.com:sourcegraph/sourcegraph.git`)
					return cmd.Run().StreamLines(cio.Write)
				},
			},
			{
				Name: "github.com/sourcegraph/dev-private",
				Description: `In order to run the local development environment as a Sourcegraph teammate,
you'll need to clone another repository: github.com/sourcegraph/dev-private.

It contains convenient preconfigured settings and code host connections.

It needs to be cloned into the same folder as sourcegraph/sourcegraph,
so they sit alongside each other, like this:

    /dir
    |-- dev-private
    +-- sourcegraph
`,
				Check: func(ctx context.Context, out *std.Output, args CheckArgs) error {
					ok, err := pathExists("dev-private")
					if ok && err == nil {
						return nil
					}
					wd, err := os.Getwd()
					if err != nil {
						return errors.Wrap(err, "failed to check for dev-private repository")
					}

					p := filepath.Join(wd, "..", "dev-private")
					ok, err = pathExists(p)
					if ok && err == nil {
						return nil
					}
					return errors.New("could not find dev-private repository either in current directory or one above")
				},
				Fix: func(ctx context.Context, cio check.IO, args CheckArgs) error {
					rootDir, err := root.RepositoryRoot()
					if err != nil {
						return errors.Wrap(err, "sourcegraph/sourcegraph should be cloned first")
					}

					return run.Cmd(ctx, `git clone git@github.com:sourcegraph/dev-private.git`).
						// Clone to parent
						Dir(filepath.Join(rootDir, "..")).
						Run().
						StreamLines(cio.Verbose)
				},
			},
		},
	}
}

// categoryProgrammingLanguagesAndTools sets up programming languages and tooling using
// asdf, which is uniform across platforms. It takes an optional list of additonalChecks, useful
// when they depend on the plaftorm we're installing them on.
func categoryProgrammingLanguagesAndTools(additionalChecks ...*dependency) category {
	categories := category{
		Name:      "Programming languages & tooling",
		DependsOn: []string{depsCloneRepo, depsBaseUtilities},
		Enabled:   enableOnlyInSourcegraphRepo(),
		Checks: []*dependency{
			{
				Name:  "go",
				Check: checkGoVersion,
				Fix: func(ctx context.Context, cio check.IO, args CheckArgs) error {
					if err := forceASDFPluginAdd(ctx, "golang", "https://github.com/kennyp/asdf-golang.git"); err != nil {
						return err
					}
					return root.Run(usershell.Command(ctx, "asdf install golang")).StreamLines(cio.Verbose)
				},
			},
			{
				Name:  "python",
				Check: checkPythonVersion,
				Fix: func(ctx context.Context, cio check.IO, args CheckArgs) error {
					if err := forceASDFPluginAdd(ctx, "python", ""); err != nil {
						return err
					}
					return root.Run(usershell.Command(ctx, "asdf install python")).StreamLines(cio.Verbose)
				},
			},
			{
				Name:        "pnpm",
				Description: "Run `asdf plugin add pnpm && asdf install pnpm`",
				Check:       checkPnpmVersion,
				Fix: func(ctx context.Context, cio check.IO, args CheckArgs) error {
					if err := forceASDFPluginAdd(ctx, "pnpm", ""); err != nil {
						return err
					}
					return root.Run(usershell.Command(ctx, "asdf install pnpm")).StreamLines(cio.Verbose)
				},
			},
			{
				Name:  "node",
				Check: checkNodeVersion,
				Fix: func(ctx context.Context, cio check.IO, args CheckArgs) error {
					if err := forceASDFPluginAdd(ctx, "nodejs", "https://github.com/asdf-vm/asdf-nodejs.git"); err != nil {
						return err
					}
					return root.Run(usershell.Command(ctx, "asdf install nodejs")).StreamLines(cio.Verbose)
				},
			},
			{
				Name:  "rust",
				Check: checkRustVersion,
				Fix: func(ctx context.Context, cio check.IO, args CheckArgs) error {
					if err := forceASDFPluginAdd(ctx, "rust", "https://github.com/asdf-community/asdf-rust.git"); err != nil {
						return err
					}
					return root.Run(usershell.Command(ctx, "asdf install rust")).StreamLines(cio.Verbose)
				},
			},
			{
				Name:        "asdf reshim",
				Description: "Regenerate asdf shims",
				Check: func(ctx context.Context, out *std.Output, args CheckArgs) error {
					// If any of these fail with ErrNotInPath, we may need to regenerate
					// all our asdf shims.
					for _, c := range []check.CheckAction[CheckArgs]{
						checkGoVersion, checkPnpmVersion, checkNodeVersion, checkRustVersion, checkPythonVersion,
					} {
						if err := c(ctx, out, args); err != nil {
							return errors.Wrap(err, "we may need to regenerate asdf shims")
						}
					}
					return nil
				},
				Fix: cmdFixes(
					`rm -rf ~/.asdf/shims`,
					`asdf reshim`,
				),
			},
			{
				Name: "pre-commit.com is installed",
				Check: func(ctx context.Context, out *std.Output, args CheckArgs) error {
					if args.DisablePreCommits {
						return nil
					}

					repoRoot, err := root.RepositoryRoot()
					if err != nil {
						return err
					}
					return check.Combine(
						check.FileExists(filepath.Join(repoRoot, ".bin/pre-commit-3.3.2.pyz")),
						func(context.Context) error {
							return root.Run(usershell.Command(ctx, "cat .git/hooks/pre-commit | grep https://pre-commit.com")).Wait()
						},
					)(ctx)
				},
				Fix: func(ctx context.Context, cio check.IO, args CheckArgs) error {
					err := root.Run(usershell.Command(ctx, "mkdir -p .bin && curl -L --retry 3 --retry-max-time 120 https://github.com/pre-commit/pre-commit/releases/download/v3.3.2/pre-commit-3.3.2.pyz --output .bin/pre-commit-3.3.2.pyz --silent")).StreamLines(cio.Verbose)
					if err != nil {
						return errors.Wrap(err, "failed to download pre-commit release")
					}
					err = root.Run(usershell.Command(ctx, "python .bin/pre-commit-3.3.2.pyz install")).StreamLines(cio.Verbose)
					if err != nil {
						return errors.Wrap(err, "failed to install pre-commit")
					}
					return nil
				},
			},
		},
	}
	categories.Checks = append(categories.Checks, additionalChecks...)
	return categories
}

func categoryAdditionalSGConfiguration() category {
	return category{
		Name: "Additional sg configuration",
		Checks: []*dependency{
			{
				Name: "Autocompletions",
				Check: func(ctx context.Context, out *std.Output, args CheckArgs) error {
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
				Fix: func(ctx context.Context, cio check.IO, args CheckArgs) error {
					sgHome, err := root.GetSGHomePath()
					if err != nil {
						return err
					}
					// sgHome needs to have appropriate permissions
					if err := os.Chmod(sgHome, os.ModePerm); err != nil {
						return errors.Wrap(err, "failed to chmod sg home")
					}

					shell := usershell.ShellType(ctx)
					if shell == "" {
						return errors.New("failed to detect shell type")
					}

					// Generate the completion script itself
					autocompleteScript := usershell.AutocompleteScripts[shell]
					autocompletePath := usershell.AutocompleteScriptPath(sgHome, shell)
					_ = os.Remove(autocompletePath) // forcibly remove old version first
					if err := os.WriteFile(autocompletePath, []byte(autocompleteScript), os.ModePerm); err != nil {
						return errors.Wrap(err, "generatng autocomplete script")
					}

					// Add the completion script to shell
					shellConfig := usershell.ShellConfigPath(ctx)
					if shellConfig == "" {
						return errors.New("Failed to detect shell config path")
					}
					conf, err := os.ReadFile(shellConfig)
					if err != nil {
						return err
					}

					// Compinit needs to be initialized
					if shell == usershell.ZshShell && !strings.Contains(string(conf), "compinit") {
						cio.Verbosef("Adding compinit to %s", shellConfig)
						if err := usershell.Run(ctx,
							"echo", run.Arg(`autoload -Uz compinit && compinit`), ">>", shellConfig,
						).Wait(); err != nil {
							return err
						}
					}

					if !strings.Contains(string(conf), autocompletePath) {
						cio.Verbosef("Adding configuration to %s", shellConfig)
						if err := usershell.Run(ctx,
							"echo", run.Arg(`PROG=sg source `+autocompletePath), ">>", shellConfig,
						).Wait(); err != nil {
							return err
						}
					}

					return nil
				},
			},
		},
	}
}

var gcloudSourceRegexp = regexp.MustCompile(`(Source \[)(?P<path>[^\]]*)(\] in your profile)`)

func dependencyGcloud() *dependency {
	return &dependency{
		Name: "gcloud",
		Check: checkAction(
			check.Combine(
				check.InPath("gcloud"),
				check.FileExists("~/.config/gcloud/application_default_credentials.json"),
				// User should have logged in with a sourcegraph.com account
				check.CommandOutputContains("gcloud auth list", "@sourcegraph.com"),
			),
		),
		Fix: func(ctx context.Context, cio check.IO, args CheckArgs) error {
			if cio.Input == nil {
				return errors.New("interactive input required to fix this check")
			}

			if err := check.InPath("gcloud")(ctx); err != nil {
				var pathsToSource []string

				// This is the official interactive installer: https://cloud.google.com/sdk/docs/downloads-interactive
				if err := usershell.Command(ctx,
					"curl https://sdk.cloud.google.com | bash -s -- --disable-prompts").
					Input(cio.Input).
					Run().
					Pipeline(pipeline.Map(func(line []byte) []byte {
						// Listen for gcloud telling us to source paths
						if matches := gcloudSourceRegexp.FindSubmatch(line); len(matches) > 0 {
							shouldSource := matches[gcloudSourceRegexp.SubexpIndex("path")]
							if len(shouldSource) > 0 {
								pathsToSource = append(pathsToSource, string(shouldSource))
							}
						}
						return line
					})).
					StreamLines(cio.Write); err != nil {
					return err
				}

				// If gcloud tells us to source some stuff, try to do it
				if len(pathsToSource) > 0 {
					shellConfig := usershell.ShellConfigPath(ctx)
					if shellConfig == "" {
						return errors.New("Failed to detect shell config path")
					}
					conf, err := os.ReadFile(shellConfig)
					if err != nil {
						return err
					}
					for _, p := range pathsToSource {
						if !bytes.Contains(conf, []byte(p)) {
							source := fmt.Sprintf("source %s", p)
							cio.Verbosef("Adding %q to %s", source, shellConfig)
							if err := usershell.Run(ctx,
								"echo", run.Arg(source), ">>", shellConfig,
							).Wait(); err != nil {
								return errors.Wrapf(err, "adding %q", source)
							}
						}
					}
				}
			}

			if err := usershell.Command(ctx, "gcloud auth application-default login").Input(cio.Input).Run().StreamLines(cio.Write); err != nil {
				return err
			}

			return usershell.Command(ctx, "gcloud auth configure-docker").Run().Wait()
		},
	}
}
