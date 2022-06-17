package dependencies

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
)

func aptGetInstall(pkg string, preinstall ...string) check.FixAction[CheckArgs] {
	commands := []string{
		`sudo apt-get update`,
	}
	commands = append(commands, preinstall...)
	commands = append(commands, fmt.Sprintf("sudo apt-get install -y %s", pkg))
	return cmdFixes(commands...)
}

// Ubuntu declares Ubuntu dependencies.
var Ubuntu = []category{
	{
		Name: depsBaseUtilities,
		Checks: []*dependency{
			{
				Name:  "gcc",
				Check: checkAction(check.InPath("gcc")),
				Fix:   aptGetInstall("build-essential"),
			},
			{
				Name:  "git",
				Check: checkAction(check.Combine(check.InPath("git"), checkGitVersion(">= 3.34.1"))),
				Fix:   aptGetInstall("git", "sudo add-apt-repository ppa:git-core/ppa"),
			}, {
				Name:  "pcre",
				Check: checkAction(check.HasUbuntuLibrary("libpcre3-dev")),
				Fix:   aptGetInstall("libpcre3-dev"),
			},
			{
				Name:  "sqlite",
				Check: checkAction(check.HasUbuntuLibrary("libsqlite3-dev")),
				Fix:   aptGetInstall("libsqlite3-dev"),
			},
			{
				Name:  "libev",
				Check: checkAction(check.HasUbuntuLibrary("libev-dev")),
				Fix:   aptGetInstall("libev-dev"),
			},
			{
				Name:  "pkg-config",
				Check: checkAction(check.InPath("pkg-config")),
				Fix:   aptGetInstall("pkg-config"),
			},
			{
				Name:  "jq",
				Check: checkAction(check.InPath("jq")),
				Fix:   aptGetInstall("jq"),
			},
			{
				Name:  "curl",
				Check: checkAction(check.InPath("curl")),
				Fix:   aptGetInstall("curl"),
			},
			// Comby will fail systematically on linux/arm64 as there aren't binaries available for that platform.
			{
				Name:  "comby",
				Check: checkAction(check.InPath("comby")),
				Fix:   cmdFix(`bash <(curl -sL get-comby.netlify.app)`),
			},
			{
				Name:  "bash",
				Check: checkAction(check.CommandOutputContains("bash --version", "version 5")),
				Fix:   aptGetInstall("bash"),
			},
		},
	},
}
