package dependencies

import (
	"context"
	"os"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/usershell"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const (
	depsHomebrew      = "Homebrew"
	depsBaseUtilities = "Base utilities"
)

// Mac declares Mac dependencies.
var Mac = []category{
	{
		Name: depsHomebrew,
		Checks: []*dependency{
			{
				Name:        "brew",
				Check:       checkAction(check.InPath("brew")),
				Fix:         cmdAction(`eval $(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)`),
				Description: `We depend on having the Homebrew package manager available on macOS: https://brew.sh`,
			},
		},
	},
	{
		Name:      depsBaseUtilities,
		DependsOn: []string{depsHomebrew},
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
				Enabled: func(ctx context.Context, args CheckArgs) error {
					// Docker is quite funky in CI
					if os.Getenv("CI") == "true" {
						return errors.New("skipping Docker in CI")
					}
					return nil
				},
				Check: checkAction(check.Combine(
					check.WrapErrMessage(check.InPath("docker"),
						"if Docker is installed and the check fails, you might need to restart terminal and 'sg setup'"),
				)),
				Fix: func(ctx context.Context, cio check.IO, args CheckArgs) error {
					// TODO stream lines
					if err := usershell.Cmd(ctx, `brew install --cask docker`).Run(); err != nil {
						return err
					}

					cio.Verbose("Docker installed - attempting to start docker")

					return usershell.Cmd(ctx, "open --hide --background /Applications/Docker.app").Run()
				},
			},
			{
				Name:  "asdf",
				Check: checkAction(check.CommandOutputContains("asdf", "version")),
				Fix: func(ctx context.Context, cio check.IO, args CheckArgs) error {
					// Uses `&&` to avoid appending the shell config on failed installations attempts.
					cmd := `brew install asdf && echo ". ${HOMEBREW_PREFIX:-/usr/local}/opt/asdf/libexec/asdf.sh" >> ` + usershell.ShellConfigPath(ctx)
					return usershell.Cmd(ctx, cmd).Run()
				},
			},
		},
	},
	categoryCloneRepositories(),
	{
		Name:      "Programming languages & tooling",
		DependsOn: []string{depsHomebrew, depsBaseUtilities},
		Checks: []*check.Check[CheckArgs]{
			{
				Name:  "go",
				Check: checkGoVersion,
				Fix: cmdsAction(
					"asdf plugin-add golang https://github.com/kennyp/asdf-golang.git",
					"asdf install golang",
				),
			},
			{
				Name:  "yarn",
				Check: checkYarnVersion,
				Fix: cmdsAction(
					"brew install gpg",
					"asdf plugin-add yarn",
					"asdf install yarn",
				),
			},
			{
				Name:  "node",
				Check: checkNodeVersion,
				Fix: cmdsAction(
					"asdf plugin add nodejs https://github.com/asdf-vm/asdf-nodejs.git",
					`grep -s "legacy_version_file = yes" ~/.asdfrc >/dev/null || echo 'legacy_version_file = yes' >> ~/.asdfrc`,
					"asdf install nodejs",
				),
			},
			{
				Name:  "rust",
				Check: checkRustVersion,
				Fix: cmdsAction(
					"asdf plugin-add rust https://github.com/asdf-community/asdf-rust.git",
					"asdf install rust",
				),
			},
		},
	},
}
