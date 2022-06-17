package dependencies

import (
	"context"
	"fmt"
	"runtime"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/usershell"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
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
			{
				Name: "asdf",
				// TODO add the if Keegan check
				Check: checkAction(check.CommandOutputContains("asdf", "version")),
				Fix: func(ctx context.Context, cio check.IO, args CheckArgs) error {
					if err := usershell.Run(ctx, "git clone https://github.com/asdf-vm/asdf.git ~/.asdf --branch v0.9.0").StreamLines(cio.Verbose); err != nil {
						return err
					}
					return usershell.Run(ctx,
						`echo ". $HOME/.asdf/asdf.sh" >>`, usershell.ShellConfigPath(ctx),
					).Wait()
				},
			},
		},
	},
	{
		Name:    depsDocker,
		Enabled: disableInCI(), // Very wonky in CI
		Checks: []*dependency{
			{
				Name:  "Docker",
				Check: checkAction(check.InPath("docker")),
				Fix: aptGetInstall(
					"docker-ce docker-ce-cli",
					"curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -",
					fmt.Sprintf(`sudo add-apt-repository "deb [arch=%s] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable`, runtime.GOARCH)),
			},
			{
				Name: "Docker without sudo",
				Check: checkAction(check.Combine(
					check.InPath("docker"),
					// It's possible that the user that installed Docker this way needs sudo to run it, which is not
					// convenient. The following check diagnose that case.
					check.CommandOutputContains("docker ps", "CONTAINER")),
				),
				Fix: func(ctx context.Context, cio check.IO, args CheckArgs) error {
					if err := usershell.Command(ctx, "sudo groupadd docker || true").Run().StreamLines(cio.Verbose); err != nil {
						return err
					}
					if err := usershell.Command(ctx, "sudo usermod -aG docker $USER").Run().StreamLines(cio.Verbose); err != nil {
						return err
					}
					err := check.CommandOutputContains("docker ps", "CONTAINER")(ctx)
					if err != nil {
						cio.WriteAlertf(`You may need to restart your terminal for the permissions needed for Docker to take effect or you can run "newgrp docker" and restart the processe in this terminal.`)
					}
					return err
				},
			},
		},
	},
	categoryCloneRepositories(),
	{
		Name:      "Programming languages & tooling",
		DependsOn: []string{depsCloneRepo, depsBaseUtilities},
		Enabled:   enableOnlyInSourcegraphRepo(),
		Checks: []*check.Check[CheckArgs]{
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
					return cmdFixes(
						`grep -s "legacy_version_file = yes" ~/.asdfrc >/dev/null || echo 'legacy_version_file = yes' >> ~/.asdfrc`,
						"asdf install nodejs",
					)(ctx, cio, args)
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
		},
	},
}
