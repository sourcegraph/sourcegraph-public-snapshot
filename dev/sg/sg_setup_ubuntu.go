package main

import (
	"context"
	"fmt"
	"runtime"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/usershell"
)

var ubuntuOSDependencies = []dependencyCategory{
	{
		name: "Install base utilities (git, docker, ...)",
		dependencies: []*dependency{
			{name: "gcc", check: check.InPath("gcc"), instructionsCommands: `sudo apt-get update && sudo apt-get install -y build-essential`},
			{name: "git", check: getCheck("git"), instructionsCommands: `sudo apt-get update && sudo add-apt-repository ppa:git-core/ppa; sudo apt-get install -y git`},
			{name: "pcre", check: check.HasUbuntuLibrary("libpcre3-dev"), instructionsCommands: "sudo apt-get update && sudo apt-get -y install libpcre3-dev"},
			{name: "sqlite", check: check.HasUbuntuLibrary("libsqlite3-dev"), instructionsCommands: "sudo apt-get update && sudo apt-get -y install libsqlite3-dev"},
			{name: "libev", check: check.HasUbuntuLibrary("libev-dev"), instructionsCommands: "sudo apt-get update && sudo apt-get -y install libev-dev"},
			{name: "pkg-config", check: check.InPath("pkg-config"), instructionsCommands: `sudo apt-get update && sudo apt-get -y install pkg-config`},
			{name: "jq", check: check.InPath("jq"), instructionsCommands: `sudo apt-get update && sudo apt-get -y install jq`},
			{name: "curl", check: check.InPath("curl"), instructionsCommands: `sudo apt-get update && sudo apt-get -y install curl`},
			// Comby will fail systematically on linux/arm64 as there aren't binaries available for that platform.
			{name: "comby", check: check.InPath("comby"), instructionsCommands: `bash <(curl -sL get-comby.netlify.app)`},
			{name: "bash", check: check.CommandOutputContains("bash --version", "version 5"), instructionsCommands: `sudo apt-get update && sudo apt-get -y install bash`},
			{
				name: "docker",
				check: check.Combine(
					check.InPath("docker"),
					// It's possible that the user that installed Docker this way needs sudo to run it, which is not
					// convenient. The following check diagnose that case.
					check.CommandOutputContains("sudo docker ps", "CONTAINER")),
				instructionsCommandsBuilder: stringCommandBuilder(func(ctx context.Context) string {
					return fmt.Sprintf(`curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
sudo add-apt-repository "deb [arch=%s] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
sudo apt-get install -y docker-ce docker-ce-cli
`, runtime.GOARCH)
				}),
			},
		},
		autoFixing: true,
	},
	{
		name: "Groups and permissions",
		dependencies: []*dependency{
			{
				name: "docker without sudo",
				check: check.Combine(
					check.InPath("docker"),
					// It's possible that the user that installed Docker this way needs sudo to run it, which is not
					// convenient. The following check diagnose that case.
					check.CommandOutputContains("docker ps", "CONTAINER")),
				requiresSgSetupRestart: true,
				instructionsComment: `You may need to restart your terminal for the permissions needed for Docker to take effect
or you can run "newgrp docker" and restart the processe in this terminal.`,
				instructionsCommands: `sudo groupadd docker || true
sudo usermod -aG docker $USER`,
			},
		},
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
		autoFixingDependencies: []*dependency{
			{
				name:  "asdf",
				check: getCheck("asdf"),
				instructionsCommandsBuilder: stringCommandBuilder(func(ctx context.Context) string {
					// Uses `&&` to avoid appending the shell config on failed installations attempts.
					return `git clone https://github.com/asdf-vm/asdf.git ~/.asdf --branch v0.9.0 && echo ". $HOME/.asdf/asdf.sh" >> ` + usershell.ShellConfigPath(ctx)
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
asdf install golang`,
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
				// There's a bug on ubuntu-arm64: curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg => Segmentation fault (core dumped)
				// https://bugs.launchpad.net/ubuntu/+source/openssl/+bug/1951279
				instructionsCommands: `
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
				check: getCheck("psql"),
				instructionsComment: `` +
					`psql, the PostgreSQL CLI client, needs to be available in your $PATH.

If you used another method, make sure psql is available.`,
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
					We recommend installing it with your package manager and starting it as a system service.`,
				instructionsCommands: `sudo apt install -y redis-server
sudo systemctl enable --now redis-server.service`,
			},
		},
	},
	{
		name:               "Set up proxy for local development",
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
				// Convoluted directions from https://developer.1password.com/docs/cli/get-started/#install
				instructionsCommands: `
curl -sS https://downloads.1password.com/linux/keys/1password.asc | sudo gpg --dearmor --output /usr/share/keyrings/1password-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/1password-archive-keyring.gpg] https://downloads.1password.com/linux/debian/$(dpkg --print-architecture) stable main" | sudo tee /etc/apt/sources.list.d/1password.list
sudo mkdir -p /etc/debsig/policies/AC2D62742012EA22/
curl -sS https://downloads.1password.com/linux/debian/debsig/1password.pol | sudo tee /etc/debsig/policies/AC2D62742012EA22/1password.pol
sudo mkdir -p /usr/share/debsig/keyrings/AC2D62742012EA22
curl -sS https://downloads.1password.com/linux/keys/1password.asc | sudo gpg --dearmor --output /usr/share/debsig/keyrings/AC2D62742012EA22/debsig.gpg
sudo apt update && sudo apt install 1password-cli
eval $(op account add --address team-sourcegraph.1password.com --signin)
`,
			},
		},
	},
}
