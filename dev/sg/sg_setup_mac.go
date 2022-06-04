package main

import (
	"context"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/usershell"
)

var macosDependencyBrew = &dependency{
	name:  "brew",
	check: check.InPath("brew"),
	instructionsComment: `We depend on having the Homebrew package manager available on macOS.

Follow the instructions at https://brew.sh to install it, then rerun 'sg setup'.`,
}

var macOSDependencies = []dependencyCategory{
	{
		name: "Install homebrew",
		dependencies: []*dependency{
			macosDependencyBrew,
		},
		autoFixing: false,
	},
	{
		name: "Install base utilities (git, docker, ...)",
		dependencies: []*dependency{
			{name: "git", check: getCheck("git"), instructionsCommands: `brew install git`},
			{name: "gnu-sed", check: check.InPath("gsed"), instructionsCommands: "brew install gnu-sed"},
			{name: "findutils", check: check.InPath("gfind"), instructionsCommands: "brew install findutils"},
			{name: "comby", check: check.InPath("comby"), instructionsCommands: "brew install comby"},
			{name: "pcre", check: check.InPath("pcregrep"), instructionsCommands: `brew install pcre`},
			{name: "sqlite", check: check.InPath("sqlite3"), instructionsCommands: `brew install sqlite`},
			{name: "jq", check: check.InPath("jq"), instructionsCommands: `brew install jq`},
			{name: "bash", check: check.CommandOutputContains("bash --version", "version 5"), instructionsCommands: `brew install bash`},
			{
				name: "rosetta",
				check: check.Any(
					check.CommandOutputContains("uname -m", "x86_64"), // will return true on non-m1 macs
					check.CommandExitCode("pgrep oahd", 0)),           // oahd is the process running rosetta
				instructionsCommands: `softwareupdate --install-rosetta --agree-to-license`,
			},
			{
				name:                 "docker",
				check:                getCheck("docker-installed"),
				instructionsCommands: `brew install --cask docker`,
			},
		},
		autoFixing:             true,
		autoFixingDependencies: []*dependency{macosDependencyBrew},
	},
	{
		name: "Clone repositories",
		dependencies: []*dependency{
			{
				name:  "SSH authentication with GitHub.com",
				check: check.CommandOutputContains("ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -T git@github.com", "successfully authenticated"),
				instructionsComment: `` +
					`Make sure that you can clone git repositories from GitHub via SSH.
See here on how to set that up:

https://docs.github.com/en/authentication/connecting-to-github-with-ssh
`,
			},
			{
				name:                 "github.com/sourcegraph/sourcegraph",
				check:                checkInMainRepoOrRepoInDirectory,
				instructionsCommands: `git clone git@github.com:sourcegraph/sourcegraph.git`,
				instructionsComment: `` +
					`The 'sourcegraph' repository contains the Sourcegraph codebase and everything to run Sourcegraph locally.`,
			},
			{
				name:                 "github.com/sourcegraph/dev-private",
				check:                checkDevPrivateInParentOrInCurrentDirectory,
				instructionsCommands: `git clone git@github.com:sourcegraph/dev-private.git`,
				instructionsComment: `` +
					`In order to run the local development environment as a Sourcegraph teammate,
you'll need to clone another repository: github.com/sourcegraph/dev-private.

It contains convenient preconfigured settings and code host connections.

It needs to be cloned into the same folder as sourcegraph/sourcegraph,
so they sit alongside each other, like this:

   /dir
   |-- dev-private
   +-- sourcegraph

NOTE: You can ignore this if you're not a Sourcegraph teammate.
`,
				onlyTeammates: true,
			},
		},
		autoFixing: true,
	},
	{
		name:               "Programming languages & tooling",
		requiresRepository: true,
		autoFixing:         true,
		// autoFixingDependencies are only accounted for it the user asks to fix the category. Otherwise, they'll never be
		// checked nor print an error, because the only thing that matters to run Sourcegraph are the final dependencies
		// defined in the dependencies field itself.
		autoFixingDependencies: []*dependency{
			{
				name:  "asdf",
				check: getCheck("asdf"),
				instructionsCommandsBuilder: stringCommandBuilder(func(ctx context.Context) string {
					// Uses `&&` to avoid appending the shell config on failed installations attempts.
					return `brew install asdf && echo ". ${HOMEBREW_PREFIX:-/usr/local}/opt/asdf/libexec/asdf.sh" >> ` + usershell.ShellConfigPath(ctx)
				}),
			},
		},
		dependencies: []*dependency{
			{
				name:  "go",
				check: getCheck("go"),
				instructionsComment: `` +
					`Souregraph requires Go to be installed.

Check the .tool-versions file for which version.

We *highly recommend* using the asdf version manager to install and manage
programming languages and tools. Find out how to install asdf here:

	https://asdf-vm.com/guide/getting-started.html

Once you have asdf, execute the commands below.`,
				instructionsCommands: `
asdf plugin-add golang https://github.com/kennyp/asdf-golang.git
asdf install golang
`,
			},
			{
				name:  "yarn",
				check: getCheck("yarn"),
				instructionsComment: `` +
					`Souregraph requires Yarn to be installed.

Check the .tool-versions file for which version.

We *highly recommend* using the asdf version manager to install and manage
programming languages and tools. Find out how to install asdf here:

	https://asdf-vm.com/guide/getting-started.html

Once you have asdf, execute the commands below.`,
				instructionsCommands: `
brew install gpg
asdf plugin-add yarn
asdf install yarn
`,
			},
			{
				name:  "node",
				check: getCheck("node"),
				instructionsComment: `` +
					`Souregraph requires Node.JS to be installed.

Check the .tool-versions file for which version.

We *highly recommend* using the asdf version manager to install and manage
programming languages and tools. Find out how to install asdf here:

	https://asdf-vm.com/guide/getting-started.html

Once you have asdf, execute the commands below.`,
				instructionsCommands: `
asdf plugin add nodejs https://github.com/asdf-vm/asdf-nodejs.git
grep -s "legacy_version_file = yes" ~/.asdfrc >/dev/null || echo 'legacy_version_file = yes' >> ~/.asdfrc
asdf install nodejs
`,
			},
			{
				name:  "rust",
				check: getCheck("rust"),
				instructionsComment: `` +
					`Souregraph requires Rust to be installed.

Check the .tool-versions file for which version.

We *highly recommend* using the asdf version manager to install and manage
programming languages and tools. Find out how to install asdf here:

	https://asdf-vm.com/guide/getting-started.html

Once you have asdf, execute the commands below.`,
				instructionsCommands: `
asdf plugin-add rust https://github.com/asdf-community/asdf-rust.git
asdf install rust
`,
			},
		},
	},
	{
		name:               "Set up PostgreSQL database",
		requiresRepository: true,
		autoFixing:         true,
		dependencies: []*dependency{
			{
				name: "Install Postgresql",
				// In the eventuality of the user using a non standard configuration and having
				// set it up appropriately in its configuration, we can bypass the standard postgres
				// check and directly check for the sourcegraph database.
				//
				// Because only the latest error is returned, it's better to finish with the real check
				// for error message clarity.
				check: getCheck("postgres"),
				instructionsComment: `` +
					`Sourcegraph requires the PostgreSQL database to be running.

We recommend installing it with Homebrew and starting it as a system service.
If you know what you're doing, you can also install PostgreSQL another way.
For example: you can use https://postgresapp.com/

If you're not sure: use the recommended commands to install PostgreSQL.`,
				instructionsCommands: `brew reinstall postgresql && brew services start postgresql
sleep 5
createdb || true
`,
			},
			{
				name:  "Connection to 'sourcegraph' database",
				check: getCheck("sourcegraph-database"),
				instructionsComment: `` +
					`Once PostgreSQL is installed and running, we need to set up Sourcegraph database itself and a
specific user.`,
				instructionsCommands: `createuser --superuser sourcegraph || true
psql -c "ALTER USER sourcegraph WITH PASSWORD 'sourcegraph';"
createdb --owner=sourcegraph --encoding=UTF8 --template=template0 sourcegraph
`,
			},
			{
				name:  "psql",
				check: getCheck("psql"),
				instructionsComment: `` +
					`psql, the PostgreSQL CLI client, needs to be available in your $PATH.

If you've installed PostgreSQL with Homebrew that should be the case.

If you used another method, make sure psql is available.`,
			},
		},
	},
	{
		name:               "Set up Redis database",
		autoFixing:         true,
		requiresRepository: true,
		dependencies: []*dependency{
			{
				name:  "Connection to Redis",
				check: getCheck("redis"),
				instructionsComment: `` +
					`Sourcegraph requires the Redis database to be running.
					We recommend installing it with Homebrew and starting it as a system service.`,
				instructionsCommands: "brew reinstall redis && brew services start redis",
			},
		},
	},
	{
		name:               "Set up proxy for local development",
		autoFixing:         true,
		requiresRepository: true,
		dependencies: []*dependency{
			{
				name:  "/etc/hosts contains sourcegraph.test",
				check: getCheck("sourcegraph-test-host"),
				instructionsComment: `` +
					`Sourcegraph should be reachable under https://sourcegraph.test:3443.
					To do that, we need to add sourcegraph.test to the /etc/hosts file.`,
				instructionsCommands: `./dev/add_https_domain_to_hosts.sh`,
			},
			{
				name:  "Caddy root certificate is trusted by system",
				check: getCheck("caddy-trusted"),
				instructionsComment: `` +
					`In order to use TLS to access your local Sourcegraph instance, you need to
trust the certificate created by Caddy, the proxy we use locally.

YOU NEED TO RESTART 'sg setup' AFTER RUNNING THIS COMMAND!`,
				instructionsCommands:   `./dev/caddy.sh trust`,
				requiresSgSetupRestart: true,
			},
		},
	},
	dependencyCategoryAdditionalSgConfiguration,
	{
		name:       "Set up cloud services",
		autoFixing: true,
		dependencies: []*dependency{
			dependencyGcloud,
			{
				name:          "1password",
				onlyTeammates: true,
				check:         check1password(),
				instructionsCommands: `
brew install --cask 1password/tap/1password-cli
eval $(op account add --address team-sourcegraph.1password.com --signin)
`,
			},
		},
	},
}
