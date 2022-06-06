package dependencies

import (
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
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
}
