package dependencies

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/run"

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
					if args.Teammate {
						return check.CommandOutputContains(
							"ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -T git@github.com",
							"successfully authenticated")(ctx)
					}
					// otherwise, we don't need auth set up at all, since everything is OSS
					return nil
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
					var cmd *run.Command
					if args.Teammate {
						cmd = run.Cmd(ctx, `git clone git@github.com:sourcegraph/sourcegraph.git`)
					} else {
						cmd = run.Cmd(ctx, `git clone https://github.com/sourcegraph/sourcegraph.git`)
					}
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

NOTE: You can ignore this if you're not a Sourcegraph teammate.`,
				Enabled: enableForTeammatesOnly(),
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
// asdf, which is uniform across platforms.
func categoryProgrammingLanguagesAndTools() category {
	return category{
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
				Name:  "yarn",
				Check: checkYarnVersion,
				Fix: func(ctx context.Context, cio check.IO, args CheckArgs) error {
					if err := forceASDFPluginAdd(ctx, "yarn", ""); err != nil {
						return err
					}
					return root.Run(usershell.Command(ctx, "asdf install yarn")).StreamLines(cio.Verbose)
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
						checkGoVersion, checkYarnVersion, checkNodeVersion, checkRustVersion,
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
		},
	}
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

					shell := usershell.ShellType(ctx)
					if shell == "" {
						return errors.New("failed to detect shell type")
					}
					autocompleteScript := usershell.AutocompleteScripts[shell]
					autocompletePath := usershell.AutocompleteScriptPath(sgHome, shell)

					cio.Verbosef("Writing autocomplete script to %s", autocompletePath)
					if err := usershell.Run(ctx,
						"echo", run.Arg(autocompleteScript), ">", autocompletePath,
					).Wait(); err != nil {
						return err
					}

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

func dependencyGcloud() *dependency {
	return &dependency{
		Name: "gcloud",
		Check: checkAction(
			check.Combine(
				check.InPath("gcloud"),
				// User should have logged in with a sourcegraph.com account
				check.CommandOutputContains("gcloud auth list", "@sourcegraph.com")),
		),
		Fix: func(ctx context.Context, cio check.IO, args CheckArgs) error {
			if cio.Input == nil {
				return errors.New("interactive input required to fix this check")
			}

			if err := check.InPath("gcloud")(ctx); err != nil {
				// This is the official interactive installer: https://cloud.google.com/sdk/docs/downloads-interactive
				if err := run.Cmd(ctx, "curl https://sdk.cloud.google.com | bash -s -- --disable-prompts").
					Input(cio.Input).
					Run().StreamLines(cio.Write); err != nil {
					return err
				}
			}

			if err := run.Cmd(ctx, "gcloud auth login").Input(cio.Input).Run().StreamLines(cio.Write); err != nil {
				return err
			}

			return run.Cmd(ctx, "gcloud auth configure-docker").Run().Wait()
		},
	}
}

func opLoginFix() check.FixAction[CheckArgs] {
	return func(ctx context.Context, cio check.IO, args CheckArgs) error {
		if cio.Input == nil {
			return errors.New("interactive input required")
		}

		key, err := cio.Output.PromptPasswordf(cio.Input, "Enter secret key:")
		if err != nil {
			return err
		}

		email, err := cio.Output.PromptPasswordf(cio.Input, "Enter account email:")
		if err != nil {
			return err
		}

		password, err := cio.Output.PromptPasswordf(cio.Input, "Enter account password:")
		if err != nil {
			return err
		}

		return usershell.Command(ctx,
			"op account add --signin --address team-sourcegraph.1password.com --email", email).
			Env(map[string]string{"OP_SECRET_KEY": key}).
			Input(strings.NewReader(password)).
			Run().
			StreamLines(cio.Verbose)
	}
}
