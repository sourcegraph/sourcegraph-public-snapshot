package dependencies

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/usershell"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Windows declares Windows dependencies.
var Windows = []category{
	{
		Name: depsBaseUtilities,
		Checks: []*dependency{
			{
				Name:  "git",
				Check: checkAction(check.Combine(check.InPath("git"), checkGitVersion(">= 2.38.1"))),
			},
			{
				Name:  "sqlite",
				Check: checkAction(check.InPath("sqlite3")),
				Fix: func(ctx context.Context, cio check.IO, args CheckArgs) error {
					cio.Write("Run 'choco install sqlite.shell' in an elevated PowerShell terminal")
					return nil
				},
			},
			{
				Name:  "jq",
				Check: checkAction(check.InPath("jq")),
				Fix: func(ctx context.Context, cio check.IO, args CheckArgs) error {
					cio.Write("Run 'choco install jq' in an elevated PowerShell terminal")
					return nil
				},
			},
			{
				Name:  "curl",
				Check: checkAction(check.InPath("curl")),
			},
			{
				Name:  "bash",
				Check: checkAction(check.CommandOutputContains("bash --version", "version 5")),
			},
			{
				// Bazelisk is a wrapper for Bazel written in Go. It automatically picks a good version of Bazel given your current working directory
				// Bazelisk replaces the bazel binary in your path
				Name:  "bazelisk",
				Check: checkAction(check.Combine(check.InPath("bazel"), check.CommandOutputContains("bazel version", "Bazelisk version"))),
				Fix: func(ctx context.Context, cio check.IO, args CheckArgs) error {
					if err := check.InPath("bazel")(ctx); err == nil {
						cio.WriteAlertf("There already exists a bazel binary in your path and it is not managed by Bazlisk. Please remove it as Bazelisk replaces the bazel binary")
						return errors.New("bazel binary already exists - please uninstall it with your package manager ex. `apt remove bazel`")
					}

					cio.Write("Run 'choco install bazelisk' in an elevated PowerShell terminal")
					return nil
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
					cio.Write("Download from https://www.docker.com/ and run")
					return nil
				},
			},
		},
	},
	categoryCloneRepositories(),
	categoryProgrammingLanguagesAndTools(
		// src-cli is installed differently on Ubuntu and Mac
		&dependency{
			Name:  "src",
			Check: checkAction(check.Combine(check.InPath("src"), checkSrcCliVersion(">= 4.0.2"))),
			Fix:   cmdFix(`sudo curl -L https://sourcegraph.com/.api/src-cli/src_linux_amd64 -o /usr/local/bin/src && sudo chmod +x /usr/local/bin/src`),
		},
	),
	{
		Name:      "Postgres database",
		DependsOn: []string{depsBaseUtilities},
		Checks: []*dependency{
			{
				Name:  "Install Postgres",
				Check: checkAction(check.Combine(check.InPath("psql"))),
				Fix: func(ctx context.Context, cio check.IO, args CheckArgs) error {
					cio.Write("Download from https://www.postgresql.org/download/windows/ and run")
					return nil
				},
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
				Description: `Sourcegraph requires the PostgreSQL database to be running.`,
				Fix: func(ctx context.Context, cio check.IO, args CheckArgs) error {
					cio.Write("Download from https://www.postgresql.org/download/windows/ and run")
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
		Name: "Redis database",
		Checks: []*dependency{
			{
				Name:  "Install Redis",
				Check: checkAction(check.InPath("redis-cli")),
				Fix:   aptGetInstall("redis-server"),
			},
			{
				Name:        "Start Redis",
				Description: `Sourcegraph requires the Redis database to be running.`,
				Check:       checkAction(check.Retry(checkRedisConnection, 5, 500*time.Millisecond)),
				Fix:         cmdFix("sudo systemctl enable --now redis-server.service"),
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
}
