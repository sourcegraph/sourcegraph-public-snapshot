package dependencies

import (
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
)

var MacOS = []category{
	{
		Name: "Install homebrew",
		Checks: []*dependency{
			{
				Name:        "brew",
				Check:       checkAction(check.InPath("brew")),
				Fix:         commandAction(`eval $(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)`),
				Description: `We depend on having the Homebrew package manager available on macOS.`,
			},
		},
	},
	{
		Name: "Install base utilities (git, docker, ...)",
		Checks: []*dependency{
			{
				Name:  "git",
				Check: checkAction(check.Combine(check.InPath("git"), checkGitVersion(">= 2.34.1"))),
				Fix:   commandAction(`brew install git`),
			},
			{
				Name:  "gnu-sed",
				Check: checkAction(check.InPath("gsed")),
				Fix:   commandAction("brew install gnu-sed"),
			},
			{
				Name:  "findutils",
				Check: checkAction(check.InPath("gfind")),
				Fix:   commandAction("brew install findutils"),
			},
			{
				Name:  "comby",
				Check: checkAction(check.InPath("comby")),
				Fix:   commandAction("brew install comby"),
			},
			{
				Name:  "pcre",
				Check: checkAction(check.InPath("pcregrep")),
				Fix:   commandAction(`brew install pcre`),
			},
			{
				Name:  "sqlite",
				Check: checkAction(check.InPath("sqlite3")),
				Fix:   commandAction(`brew install sqlite`),
			},
			{
				Name:  "jq",
				Check: checkAction(check.InPath("jq")),
				Fix:   commandAction(`brew install jq`),
			},
			{
				Name:  "bash",
				Check: checkAction(check.CommandOutputContains("bash --version", "version 5")),
				Fix:   commandAction(`brew install bash`)},
			{
				Name: "rosetta",
				Check: checkAction(
					check.Any(
						// will return true on non-m1 macs
						check.CommandOutputContains("uname -m", "x86_64"),
						// oahd is the process running rosetta
						check.CommandExitCode("pgrep oahd", 0)),
				),
				Fix: commandAction(`softwareupdate --install-rosetta --agree-to-license`),
			},
			{
				Name: "docker",
				Check: checkAction(
					check.WrapErrMessage(check.InPath("docker"),
						"if Docker is installed and the check fails, you might need to start Docker.app and restart terminal and 'sg setup'"),
				),
				Fix: commandAction(`brew install --cask docker`),
			},
		},
	},
}
