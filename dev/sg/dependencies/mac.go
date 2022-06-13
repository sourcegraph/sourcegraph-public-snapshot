package dependencies

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/usershell"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
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
				Check: checkAction(check.Combine(check.InPath("git"), checkGitVersion(">= 2.34.1"))),
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
				Name:    "docker",
				Enabled: disableInCI(), // Very wonky in CI
				Check: checkAction(check.Combine(
					check.WrapErrMessage(check.InPath("docker"),
						"if Docker is installed and the check fails, you might need to restart terminal and 'sg setup'"),
				)),
				Fix: func(ctx context.Context, cio check.IO, args CheckArgs) error {
					if err := usershell.Run(ctx, `brew install --cask docker`).StreamLines(cio.Verbose); err != nil {
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
					if err := usershell.Run(ctx, "brew install asdf").StreamLines(cio.Verbose); err != nil {
						return err
					}
					return usershell.Run(ctx,
						`echo ". ${HOMEBREW_PREFIX:-/usr/local}/opt/asdf/libexec/asdf.sh" >>`, usershell.ShellConfigPath(ctx),
					).Wait()
				},
			},
			{
				Name:        "gpg",
				Description: "Required for yarn installation.",
				Check:       checkAction(check.InPath("gpg")),
				Fix:         cmdFix("brew install gpg"),
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
			{
				Name:  "1password",
				Check: checkAction(check1password()),
				Fix: func(ctx context.Context, cio check.IO, args CheckArgs) error {
					if err := usershell.Run(ctx, "brew install --cask 1password/tap/1password-cli").StreamLines(cio.Verbose); err != nil {
						return err
					}
					if cio.Input == nil {
						return errors.New("interactive input required")
					}

					accounts, err := usershell.Run(ctx, "op account list").String()
					if err == nil {
						if strings.Contains(accounts, "team-sourcegraph.1password.com") {
							// Account already added, we just need to sign in again.
							// However, we also can't seem to automate this login without
							// going through the same setup as the initial flow because
							// the 'op signin' command refuses to accept any input, so we
							// just tell the user to log in separately and exit.
							cio.WriteNoticef("Unfortunately I can't easily fix this for you :( You should run the following command:")
							loginCmd := "eval $(op signin --account=team-sourcegraph.1password.com)"
							cio.WriteMarkdown("```sh\n" + loginCmd + "\n```")
							return errors.Newf("Cannot be fixed by automatically")
						}
					}

					cio.Promptf("Enter secret key:")
					var key string
					if _, err := fmt.Fscan(cio.Input, &key); err != nil {
						return err
					}
					cio.Promptf("Enter account email:")
					var email string
					if _, err := fmt.Fscan(cio.Input, &email); err != nil {
						return err
					}
					cio.Promptf("Enter account password:")
					var password string
					if _, err := fmt.Fscan(cio.Input, &password); err != nil {
						return err
					}

					// 1password does some weird things, and it doesn't seem to want to
					// accept piped input, so we just echo the input we want inside the
					// eval command.
					return usershell.Command(ctx, "eval", fmt.Sprintf(`$(echo "$OP_PASSWORD" | %s)`,
						`op account add --signin --address team-sourcegraph.1password.com --email "$OP_EMAIL"`)).
						Env(map[string]string{
							"OP_SECRET_KEY": key,
							"OP_PASSWORD":   password,
							"OP_EMAIL":      email,
						}).
						Run().
						StreamLines(cio.Verbose)
				},
			},
		},
	},
}
