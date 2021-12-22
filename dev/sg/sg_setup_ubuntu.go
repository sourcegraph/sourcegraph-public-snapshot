package main

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

// This ties the check to having the library installed with apt-get on Ubuntu,
// which against the principle of checking dependencies independently of their
// installation method. Given they're just there for comby and sqlite, the chances
// that someone needs to install them in a different way is fairly low, making this
// check acceptable for the time being.
func checkUbuntuLib(name string) func(context.Context) error {
	return func(ctx context.Context) error {
		_, err := combinedFreshExec(ctx, fmt.Sprintf("dpkg -s %s", name))
		return err
	}
}

var ubuntuOSDependencies = []dependencyCategory{
	{
		name: "Install base utilities (git, docker, ...)",
		dependencies: []*dependency{
			{name: "git", check: checkInPath("git"), instructionsCommands: `sudo apt-get install -y git-all`},
			{name: "pcre", check: checkUbuntuLib("libpcre3-dev"), instructionsCommands: "sudo apt-get -y install libpcre3-dev"},
			{name: "sqlite", check: checkUbuntuLib("libsqlite3-dev"), instructionsCommands: "sudo apt-get -y install libsqlite3-dev"},
			{name: "libev", check: checkUbuntuLib("libev-dev"), instructionsCommands: "sudo apt-get -y install libev-dev"},
			{name: "pkg-config", check: checkInPath("pkg-config"), instructionsCommands: `sudo apt-get -y install pkg-config`},
			{name: "jq", check: checkInPath("jq"), instructionsCommands: `sudo apt-get -y install jq`},
			{name: "curl", check: checkInPath("curl"), instructionsCommands: `sudo apt-get -y install curl`},
			// Comby will fail systematically on linux/arm64 as there aren't binaries available for that platform.
			{name: "comby", check: checkInPath("comby"), instructionsCommands: `bash <(curl -sL get.comby.dev)`},
			{name: "bash", check: checkCommandOutputContains("bash --version", "version 5"), instructionsCommands: `sudo apt-get -y install bash`},
			{
				name: "docker",
				check: combineChecks(
					checkInPath("docker"),
					// It's possible that the user that installed Docker this way needs sudo to run it, which is not
					// convenient. The following check diagnose that case.
					checkCommandOutputContains("docker ps", "CONTAINER")),
				instructionsComment: `You may need to restart your terminal for the permissions needed for Docker to take effect
or you can run "newgrp docker" and restart the processe in this terminal.`,
				instructionsCommandsBuilder: stringCommandBuilder(func(ctx context.Context) string {
					return fmt.Sprintf(`curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
sudo add-apt-repository "deb [arch=%s] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
sudo apt-get install -y docker-ce docker-ce-cli
sudo groupadd docker || true
sudo usermod -aG docker $USER
`, runtime.GOARCH)
				}),
			},
		},
		autoFixing: true,
	},
	{
		name: "Clone repositories",
		dependencies: []*dependency{
			{
				name:  "SSH authentication with GitHub.com",
				check: checkCommandOutputContains("ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -T git@github.com", "successfully authenticated"),
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
					`In order to run the local development environment as a Sourcegraph employee,
you'll need to clone another repository: github.com/sourcegraph/dev-private.

It contains convenient preconfigured settings and code host connections.

It needs to be cloned into the same folder as sourcegraph/sourcegraph,
so they sit alongside each other, like this:

   /dir
   |-- dev-private
   +-- sourcegraph

NOTE: You can ignore this if you're not a Sourcegraph employee.
`,
				onlyEmployees: true,
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
				check: checkCommandOutputContains("asdf", "version"),
				instructionsCommandsBuilder: stringCommandBuilder(func(ctx context.Context) string {
					// Uses `&&` to avoid appending the shell config on failed installations attempts.
					return `git clone https://github.com/asdf-vm/asdf.git ~/.asdf --branch v0.8.1 && echo ". $HOME/.asdf/asdf.sh" >> ` + getUserShellConfigPath(ctx)
				}),
			},
		},
		dependencies: []*dependency{
			{
				name:  "go",
				check: combineChecks(checkInPath("go"), checkCommandOutputContains("go version", "go version")),
				instructionsComment: `` +
					`Souregraph requires Go to be installed.

Check the .tool-versions file for which version.

We *highly recommend* using the asdf version manager to install and manage
programming languages and tools. Find out how to install asdf here:

	https://asdf-vm.com/guide/getting-started.html

Once you have asdf, execute the commands below.`,
				instructionsCommands: `
asdf plugin-add golang https://github.com/kennyp/asdf-golang.git
asdf install golang`,
			},
			{
				name:  "yarn",
				check: combineChecks(checkInPath("yarn"), checkCommandOutputContains("yarn version", "yarn version")),
				instructionsComment: `` +
					`Souregraph requires Yarn to be installed.

			Check the .tool-versions file for which version.

			We *highly recommend* using the asdf version manager to install and manage
			programming languages and tools. Find out how to install asdf here:

				https://asdf-vm.com/guide/getting-started.html

			Once you have asdf, execute the commands below.`,
				// There's a bug on ubuntu-arm64: curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg => Segmentation fault (core dumped)
				// https://bugs.launchpad.net/ubuntu/+source/openssl/+bug/1951279
				instructionsCommands: `
			asdf plugin-add yarn
			asdf install yarn
			`,
			},
			{
				name:  "node",
				check: combineChecks(checkInPath("node"), checkCommandOutputContains(`node -e "console.log(\"foobar\")"`, "foobar")),
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
		},
	},
	{
		name:               "Setup PostgreSQL database",
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
				check: anyChecks(checkSourcegraphDatabase, checkPostgresConnection),
				instructionsComment: `` +
					`Sourcegraph requires the PostgreSQL database to be running.

We recommend installing it with your package manager and starting it as a system service.
If you know what you're doing, you can also install PostgreSQL another way.

If you're not sure: use the recommended commands to install PostgreSQL.`,
				instructionsCommands: `sudo apt-get install -y postgresql postgresql-contrib
sudo systemctl enable --now postgresql
sleep 5
sudo -u postgres createuser --superuser $USER
createdb
`,
			},
			{
				name:  "psql",
				check: checkInPath("psql"),
				instructionsComment: `` +
					`psql, the PostgreSQL CLI client, needs to be available in your $PATH.

If you've installed PostgreSQL with Homebrew that should be the case.

If you used another method, make sure psql is available.`,
			},
			{
				name:  "Connection to 'sourcegraph' database",
				check: checkSourcegraphDatabase,
				instructionsComment: `` +
					`Once PostgreSQL is installed and running, we need to setup Sourcegraph database itself and a
specific user.`,
				instructionsCommands: `createuser --superuser sourcegraph || true
psql -c "ALTER USER sourcegraph WITH PASSWORD 'sourcegraph';" 
createdb --owner=sourcegraph --encoding=UTF8 --template=template0 sourcegraph
`,
			},
		},
	},
	{
		name:               "Setup Redis database",
		autoFixing:         true,
		requiresRepository: true,
		dependencies: []*dependency{
			{
				name:  "Connection to Redis",
				check: retryCheck(checkRedisConnection, 5, 500*time.Millisecond),
				instructionsComment: `` +
					`Sourcegraph requires the Redis database to be running.
					We recommend installing it with your package manager and starting it as a system service.`,
				instructionsCommands: `sudo apt install -y redis-server
sudo systemctl enable --now redis-server.service`,
			},
		},
	},
	{
		name:               "Setup proxy for local development",
		requiresRepository: true,
		dependencies: []*dependency{
			{
				name:  "/etc/hosts contains sourcegraph.test",
				check: checkFileContains("/etc/hosts", "sourcegraph.test"),
				instructionsComment: `` +
					`Sourcegraph should be reachable under https://sourcegraph.test:3443.
					To do that, we need to add sourcegraph.test to the /etc/hosts file.`,
				instructionsCommands: `./dev/add_https_domain_to_hosts.sh`,
			},
			{
				name:  "Caddy root certificate is trusted by system",
				check: checkCaddyTrusted,
				instructionsComment: `` +
					`In order to use TLS to access your local Sourcegraph instance, you need to
trust the certificate created by Caddy, the proxy we use locally.

YOU NEED TO RESTART 'sg setup' AFTER RUNNING THIS COMMAND!`,
				instructionsCommands:   `./dev/caddy.sh trust`,
				requiresSgSetupRestart: true,
			},
		},
	},
}
