package dependencies

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/usershell"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
)

const (
	depsHomebrew      = "Homebrew"
	depsBaseUtilities = "Base utilities"
	depsDocker        = "Docker"
	depsCloneRepo     = "Clone repositories"
)

// Mac declares Mac dependencies.
var Mac = []category{
	{
		Name: depsHomebrew,
		Checks: []*dependency{
			{
				Name:        "brew",
				Check:       checkAction(check.InPath("brew")),
				Fix:         cmdFix(`eval $(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)`),
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
				Check: checkAction(check.Combine(check.InPath("git"), checkGitVersion(">= 2.38.1"))),
				Fix:   cmdFix(`brew install git`),
			},
			{
				Name:  "gnu-sed",
				Check: checkAction(check.InPath("gsed")),
				Fix:   cmdFix("brew install gnu-sed"),
			},
			{
				Name:  "findutils",
				Check: checkAction(check.InPath("gfind")),
				Fix:   cmdFix("brew install findutils"),
			},
			{
				Name:  "comby",
				Check: checkAction(check.InPath("comby")),
				Fix:   cmdFix("brew install comby"),
			},
			{
				Name:  "pcre",
				Check: checkAction(check.InPath("pcregrep")),
				Fix:   cmdFix(`brew install pcre`),
			},
			{
				Name:  "sqlite",
				Check: checkAction(check.InPath("sqlite3")),
				Fix:   cmdFix(`brew install sqlite`),
			},
			{
				Name:  "jq",
				Check: checkAction(check.InPath("jq")),
				Fix:   cmdFix(`brew install jq`),
			},
			{
				Name:  "bash",
				Check: checkAction(check.CommandOutputContains("bash --version", "version 5")),
				Fix:   cmdFix(`brew install bash`),
			},
			{
				Name: "rosetta",
				Check: checkAction(
					check.Any(
						// will return true on non-m1 macs
						check.CommandOutputContains("uname -m", "x86_64"),
						// oahd is the process running rosetta
						check.CommandExitCode("pgrep oahd", 0)),
				),
				Fix: cmdFix(`softwareupdate --install-rosetta --agree-to-license`),
			},
			{
				Name:        "certutil",
				Description: "Required for caddy certificates.",
				Check:       checkAction(check.InPath("certutil")),
				Fix:         cmdFix(`brew install nss`),
			},
			{
				Name:  "asdf",
				Check: checkAction(check.CommandOutputContains("asdf", "version")),
				Fix: func(ctx context.Context, cio check.IO, args CheckArgs) error {
					if err := usershell.Run(ctx, "brew install asdf").StreamLines(cio.Verbose); err != nil {
						return err
					}
					return usershell.Run(ctx,
						`echo ". ${HOMEBREW_PREFIX:-/usr/local}/opt/asdf/libexec/asdf.sh" >>`, usershell.ShellConfigPath(ctx),
					).Wait()
				},
			},
		},
	},
	{
		Name:      depsDocker,
		Enabled:   disableInCI(), // Very wonky in CI
		DependsOn: []string{depsHomebrew},
		Checks: []*dependency{
			{
				Name: "docker",
				Check: checkAction(check.Combine(
					check.WrapErrMessage(check.InPath("docker"),
						"if Docker is installed and the check fails, you might need to restart terminal and 'sg setup'"),
				)),
				Fix: func(ctx context.Context, cio check.IO, args CheckArgs) error {
					if err := usershell.Run(ctx, `brew install --cask docker`).StreamLines(cio.Verbose); err != nil {
						return err
					}

					cio.Write("Docker installed - attempting to start docker")

					return usershell.Cmd(ctx, "open --hide --background /Applications/Docker.app").Run()
				},
			},
		},
	},
	categoryCloneRepositories(),
	categoryProgrammingLanguagesAndTools(
		// src-cli is installed differently on Ubuntu and Mac
		&dependency{
			Name:  "src",
			Check: checkAction(check.Combine(check.InPath("src"), checkSrcCliVersion(">= 4.2.0"))),
			Fix:   cmdFix(`brew upgrade sourcegraph/src-cli/src-cli || brew install sourcegraph/src-cli/src-cli`),
		},
	),
	{
		Name:      "Postgres database",
		DependsOn: []string{depsHomebrew},
		Checks: []*dependency{
			{
				Name: "Install Postgres",
				Description: `psql, the PostgreSQL CLI client, needs to be available in your $PATH.

If you've installed PostgreSQL with Homebrew that should be the case.

If you used another method, make sure psql is available.`,
				Check: checkAction(check.InPath("psql")),
				Fix:   cmdFix("brew install postgresql"),
			},
			{
				Name: "Start Postgres",
				// In the eventuality of the user using a non standard configuration and having
				// set it up appropriately in its configuration, we can bypass the standard postgres
				// check and directly check for the sourcegraph database.
				//
				// Because only the latest error is returned, it's better to finish with the real check
				// for error message clarity.
				Check: func(ctx context.Context, out *std.Output, args CheckArgs) error {
					if err := checkSourcegraphDatabase(ctx, out, args); err == nil {
						return nil
					}
					return checkPostgresConnection(ctx)
				},
				Description: `Sourcegraph requires the PostgreSQL database to be running.

We recommend installing it with Homebrew and starting it as a system service.
If you know what you're doing, you can also install PostgreSQL another way.
For example: you can use https://postgresapp.com/

If you're not sure: use the recommended commands to install PostgreSQL.`,
				Fix: func(ctx context.Context, cio check.IO, args CheckArgs) error {
					err := usershell.Cmd(ctx, "brew services start postgresql").Run()
					if err != nil {
						return err
					}

					// Wait for startup
					time.Sleep(5 * time.Second)

					// Doesn't matter if this succeeds
					_ = usershell.Cmd(ctx, "createdb").Run()
					return nil
				},
			},
			{
				Name:        "Connection to 'sourcegraph' database",
				Check:       checkSourcegraphDatabase,
				Description: `Once PostgreSQL is installed and running, we need to set up Sourcegraph database itself and a specific user.`,
				Fix: cmdFixes(
					"createuser --superuser sourcegraph || true",
					`psql -c "ALTER USER sourcegraph WITH PASSWORD 'sourcegraph';"`,
					`createdb --owner=sourcegraph --encoding=UTF8 --template=template0 sourcegraph`,
				),
			},
		},
	},
	{
		Name:      "Redis database",
		DependsOn: []string{depsHomebrew},
		Checks: []*dependency{
			{
				Name: "Start Redis",
				Description: `Sourcegraph requires the Redis database to be running.
We recommend installing it with Homebrew and starting it as a system service.`,
				Check: checkAction(check.Retry(checkRedisConnection, 5, 500*time.Millisecond)),
				Fix: cmdFixes(
					"brew reinstall redis",
					"brew services start redis",
				),
			},
		},
	},
	{
		Name:      "sourcegraph.test development proxy",
		DependsOn: []string{depsBaseUtilities},
		Checks: []*dependency{
			{
				Name: "/etc/hosts contains sourcegraph.test",
				Description: `Sourcegraph should be reachable under https://sourcegraph.test:3443.
To do that, we need to add sourcegraph.test to the /etc/hosts file.`,
				Check: checkAction(check.FileContains("/etc/hosts", "sourcegraph.test")),
				Fix: func(ctx context.Context, cio check.IO, args CheckArgs) error {
					return root.Run(usershell.Command(ctx, `./dev/add_https_domain_to_hosts.sh`)).StreamLines(cio.Verbose)
				},
			},
			{
				Name: "Caddy root certificate is trusted by system",
				Description: `In order to use TLS to access your local Sourcegraph instance, you need to
trust the certificate created by Caddy, the proxy we use locally.

YOU NEED TO RESTART 'sg setup' AFTER RUNNING THIS COMMAND!`,
				Enabled: disableInCI(), // Can't seem to get this working
				Check:   checkAction(checkCaddyTrusted),
				Fix: func(ctx context.Context, cio check.IO, args CheckArgs) error {
					return root.Run(usershell.Command(ctx, `./dev/caddy.sh trust`)).StreamLines(cio.Verbose)
				},
			},
		},
	},
	categoryAdditionalSGConfiguration(),
	{
		Name:      "Cloud services",
		DependsOn: []string{depsHomebrew},
		Enabled:   enableForTeammatesOnly(),
		Checks: []*dependency{
			dependencyGcloud(),
		},
	},
}
