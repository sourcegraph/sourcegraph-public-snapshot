package dependencies

import (
	"context"
	"os"
	"path/filepath"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// MacOS declares MacOS dependencies.
var MacOS = []category{
	{
		Name: "Homebrew",
		Checks: []*dependency{
			{
				Name:        "brew",
				Check:       checkAction(check.InPath("brew")),
				Fix:         cmdAction(`eval $(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)`),
				Description: `We depend on having the Homebrew package manager available on macOS.`,
			},
		},
	},
	{
		Name:      "Base utilities (git, docker, ...)",
		DependsOn: []string{"Homebrew"},
		Checks: []*dependency{
			{
				Name:  "git",
				Check: checkAction(check.Combine(check.InPath("git"), checkGitVersion(">= 2.34.1"))),
				Fix:   cmdAction(`brew install git`),
			},
			{
				Name:  "gnu-sed",
				Check: checkAction(check.InPath("gsed")),
				Fix:   cmdAction("brew install gnu-sed"),
			},
			{
				Name:  "findutils",
				Check: checkAction(check.InPath("gfind")),
				Fix:   cmdAction("brew install findutils"),
			},
			{
				Name:  "comby",
				Check: checkAction(check.InPath("comby")),
				Fix:   cmdAction("brew install comby"),
			},
			{
				Name:  "pcre",
				Check: checkAction(check.InPath("pcregrep")),
				Fix:   cmdAction(`brew install pcre`),
			},
			{
				Name:  "sqlite",
				Check: checkAction(check.InPath("sqlite3")),
				Fix:   cmdAction(`brew install sqlite`),
			},
			{
				Name:  "jq",
				Check: checkAction(check.InPath("jq")),
				Fix:   cmdAction(`brew install jq`),
			},
			{
				Name:  "bash",
				Check: checkAction(check.CommandOutputContains("bash --version", "version 5")),
				Fix:   cmdAction(`brew install bash`)},
			{
				Name: "rosetta",
				Check: checkAction(
					check.Any(
						// will return true on non-m1 macs
						check.CommandOutputContains("uname -m", "x86_64"),
						// oahd is the process running rosetta
						check.CommandExitCode("pgrep oahd", 0)),
				),
				Fix: cmdAction(`softwareupdate --install-rosetta --agree-to-license`),
			},
			{
				Name: "docker",
				Check: checkAction(
					check.WrapErrMessage(check.InPath("docker"),
						"if Docker is installed and the check fails, you might need to start Docker.app and restart terminal and 'sg setup'"),
				),
				Fix: cmdAction(`brew install --cask docker`),
			},
		},
	},
	{
		Name: "Clone repositories",
		Checks: []*dependency{
			{
				Name: "SSH authentication with GitHub.com",
				Check: checkAction(check.CommandOutputContains(
					"ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -T git@github.com",
					"successfully authenticated")),
				// 				instructionsComment: `` +
				// 					`Make sure that you can clone git repositories from GitHub via SSH.
				// See here on how to set that up:

				// https://docs.github.com/en/authentication/connecting-to-github-with-ssh
				// `,
			},
			{
				Name:        "github.com/sourcegraph/sourcegraph",
				Description: `The 'sourcegraph' repository contains the Sourcegraph codebase and everything to run Sourcegraph locally.`,
				Check: func(ctx context.Context, cio check.IO, args CheckArgs) error {
					if args.InRepo {
						return nil
					}

					ok, err := pathExists("sourcegraph")
					if !ok || err != nil {
						return errors.New("'sg setup' is not run in sourcegraph and repository is also not found in current directory")
					}
					return nil
				},
				Fix: cmdAction(`git clone git@github.com:sourcegraph/sourcegraph.git`),
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

NOTE: You can ignore this if you're not a Sourcegraph teammate.
`,
				Enabled: func(ctx context.Context, args CheckArgs) error {
					if !args.Teammate {
						return errors.New("Disabled if not a Sourcegraph teammate")
					}
					return nil
				},
				Check: func(ctx context.Context, cio check.IO, args CheckArgs) error {
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
				Fix: cmdAction(`git clone git@github.com:sourcegraph/dev-private.git`),
			},
		},
	},
}
