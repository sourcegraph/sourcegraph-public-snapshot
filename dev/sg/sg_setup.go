package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/sourcegraph/lib/output"
)

var (
	setupFlagSet = flag.NewFlagSet("sg setup", flag.ExitOnError)
	setupCommand = &ffcli.Command{
		Name:       "setup",
		ShortUsage: "sg setup",
		ShortHelp:  "Reports which version of Sourcegraph is currently live in the given environment",
		LongHelp:   "Run 'sg setup' to setup the local dev environment",
		FlagSet:    setupFlagSet,
		Exec:       setupExec,
	}
)

func setupExec(ctx context.Context, args []string) error {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		out.WriteLine(output.Linef("", output.StyleWarning, "'sg setup' currently only supports macOS and Linux"))
		os.Exit(1)
	}

	out.WriteLine(output.Linef("", output.StyleLinesAdded, "Welcome to 'sg setup'!"))
	out.Write("")
	out.Write("")

	var instructions []instruction

	currentOS := runtime.GOOS
	if overridesOS, ok := os.LookupEnv("SG_FORCE_OS"); ok {
		currentOS = overridesOS
	}
	if currentOS == "darwin" {
		instructions = macOSInstructions
	} else {
		instructions = linuxInstructions
	}
	instructions = append(instructions, httpReverseProxyInstructions...)

	conditions := map[string]bool{}

	i := 0
	for _, instruction := range instructions {
		if instruction.ifBool != "" {
			val, ok := conditions[instruction.ifBool]
			if !ok {
				out.WriteLine(output.Line("", output.StyleWarning, "Something went wrong."))
				os.Exit(1)
			}
			if !val {
				continue
			}
		}
		if instruction.ifNotBool != "" {
			val, ok := conditions[instruction.ifNotBool]
			if !ok {
				out.WriteLine(output.Line("", output.StyleWarning, "Something went wrong."))
				os.Exit(1)
			}
			if val {
				continue
			}
		}

		i++
		out.WriteLine(output.Line("", output.StylePending, "------------------------------------------"))
		out.Writef("%sStep %d:%s%s %s%s", output.StylePending, i+1, output.StyleReset, output.StyleSuccess, instruction.prompt, output.StyleReset)
		out.Write("")

		if instruction.comment != "" {
			out.WriteLine(output.Line("", output.StyleSuggestion, instruction.comment))
			out.Write("")
		}

		if instruction.command != "" {
			out.WriteLine(output.Line("", output.StyleSuggestion, "Run the following command in another terminal:"))
			out.Write("")
			out.Write(strings.TrimSpace(instruction.command) + "\n")

			out.WriteLine(output.Linef("", output.StylePending, "Hit return to confirm that you ran the command..."))
			input := bufio.NewScanner(os.Stdin)
			input.Scan()
		}

		if instruction.readsBool != "" {
			// out.WriteLine(output.Linef("", output.StylePending, instruction.prompt))
			val := getBool()
			conditions[instruction.readsBool] = val
		}
	}
	return nil
}

type instruction struct {
	prompt, comment, command string

	readsBool string
	ifBool    string
	ifNotBool string
}

var macOSInstructions = []instruction{
	{
		prompt:  "Install homebrew",
		command: "Open https://brew.sh and follow instructions there",
	},
	{
		prompt:  `Install Docker`,
		command: `brew install --cask docker`,
	},
	{
		prompt:  `Install Go, Yarn, Git, Comby, SQLite tools, and jq`,
		command: `brew install go yarn git gnu-sed comby pcre sqlite jq`,
	},
	{
		prompt:    `Do you want to use Docker or not?`,
		readsBool: `docker`,
	},
	{
		ifBool: "docker",
		prompt: "Nothing to do yet!",
		comment: `Nothing to do here, since you already installed Docker for Mac.
We provide a docker compose file at dev/redis-postgres.yml to make it easy to run Redis and PostgreSQL as Docker containers, with docker compose.`},
	{
		ifNotBool: "docker",
		prompt:    `Install PostgreSQL and Redis with the following commands`,
		command: `brew install postgres
brew install redis`,
	},
	{
		ifNotBool: "docker",
		prompt:    `(optional) Start the services (and configure them to start automatically)`,
		comment:   `You can stop them later by calling stop instead of start above.`,
		command: `brew services start postgresql
brew services start redis`,
	},
	{
		ifNotBool: "docker",
		prompt:    `Ensure psql, the PostgreSQL command line client, is on your $PATH.`,
		comment:   `Homebrew does not put it there by default. Homebrew gives you the command to run to insert psql in your path in the "Caveats" section of brew info postgresql. Alternatively, you can use the command below. It might need to be adjusted depending on your Homebrew prefix (/usr/local below) and shell (bash below).`,
		command: `hash psql || { echo 'export PATH="/usr/local/opt/postgresql/bin:$PATH"' >> ~/.bash_profile }
source ~/.bash_profile`,
	},
	{
		ifNotBool: "docker",
		prompt:    `Open a new Terminal window to ensure psql is now on your $PATH.`,
		command:   `which psql`,
	},
	{
		prompt: `Install the Node Version Manager (nvm) using:`,
		command: `NVM_VERSION="$(curl https://api.github.com/repos/nvm-sh/nvm/releases/latest | jq -r .name)"
curl -L https://raw.githubusercontent.com/nvm-sh/nvm/"$NVM_VERSION"/install.sh -o install-nvm.sh
sh install-nvm.sh`,
	},
	{
		prompt:  `After the install script is finished, re-source your shell profile (e.g., source ~/.zshrc) or restart your terminal session to pick up the nvm definitions. Re-running the install script will update the installation.`,
		command: `source ~/.zshrc`,
	},
	{
		prompt: `Install the current recommended version of Node JS by running the following in the sourcegraph/sourcegraph repository clone`,
		comment: `After doing this, node -v should show the same version mentioned in .nvmrc at the root of the sourcegraph repository.
		NOTE: Although there is a Homebrew package for Node, we advise using nvm instead, to ensure you get a Node version compatible with the current state of the sourcegraph repository.`,
		command: `nvm install
nvm use --delete-prefix`,
	},
	// step 4
	{
		ifBool: "docker",
		prompt: `To initialize your database, you may have to set the appropriate environment variables before running the createdb command:`,
		comment: `The Sourcegraph server reads PostgreSQL connection configuration from the PG* environment variables.

The development server startup script as well as the docker compose file provide default settings, so it will work out of the box.`,
		command: `export PGUSER=sourcegraph PGPASSWORD=sourcegraph PGDATABASE=sourcegraph
createdb --user=sourcegraph --owner=sourcegraph --encoding=UTF8 --template=template0 sourcegraph`,
	},
	{
		ifBool: "docker",
		prompt: `You can also use the PGDATA_DIR environment variable to specify a local folder (instead of a volume) to store the database files. See the dev/redis-postgres.yml file for more details.

This can also be spun up using sg run redis-postgres, with the following sg.config.override.yaml:`,
		command: `env:
    PGHOST: localhost
    PGPASSWORD: sourcegraph
    PGUSER: sourcegraph`,
	},
	{
		ifNotBool: "docker",
		prompt:    `Create a database for the current Unix user`,
		comment:   `You need a fresh Postgres database and a database user that has full ownership of that database.`,
		command: `sudo su - _postgres
createdb`,
	},
	{
		ifNotBool: "docker",
		prompt:    `Create the Sourcegraph user and password`,
		command: `createuser --superuser sourcegraph
psql -c "ALTER USER sourcegraph WITH PASSWORD 'sourcegraph';"`,
	},
	{
		ifNotBool: "docker",
		prompt:    `Create the Sourcegraph database`,
		command:   `createdb --owner=sourcegraph --encoding=UTF8 --template=template0 sourcegraph`,
	},
	{
		ifNotBool: "docker",
		prompt:    `Configure database settings in your environment`,
		comment: `The Sourcegraph server reads PostgreSQL connection configuration from the PG* environment variables.

Our configuration in sg.config.yaml (we'll see what sg is in the next step) sets values that work with the setup described here, but if you are using different values you can overwrite them, for example, in your ~/.bashrc:`,
		command: `export PGPORT=5432
export PGHOST=localhost
export PGUSER=sourcegraph
export PGPASSWORD=sourcegraph
export PGDATABASE=sourcegraph
export PGSSLMODE=disable`,
	},
}

var linuxInstructions = []instruction{
	{
		prompt:  `Add package repositories`,
		comment: "In order to install dependencies, we need to add some repositories to apt.",
		command: `
# Go
sudo add-apt-repository ppa:longsleep/golang-backports

# Docker
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"

# Yarn
curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | sudo apt-key add -
echo "deb https://dl.yarnpkg.com/debian/ stable main" | sudo tee /etc/apt/sources.list.d/yarn.list`,
	},
	{
		prompt:  `Update repositories`,
		command: `sudo apt-get update`,
	},
	{
		prompt: `Install dependencies`,
		command: `sudo apt install -y make git-all libpcre3-dev libsqlite3-dev pkg-config golang-go docker-ce docker-ce-cli containerd.io yarn jq libnss3-tools

# Install comby
curl -L https://github.com/comby-tools/comby/releases/download/0.11.3/comby-0.11.3-x86_64-linux.tar.gz | tar xvz

# The extracted binary must be in your $PATH available as 'comby'.
# Here's how you'd move it to /usr/local/bin (which is most likely in your $PATH):
chmod +x comby-*-linux
mv comby-*-linux /usr/local/bin/comby

# Install nvm (to manage Node.js)
NVM_VERSION="$(curl https://api.github.com/repos/nvm-sh/nvm/releases/latest | jq -r .name)"
curl -L https://raw.githubusercontent.com/nvm-sh/nvm/"$NVM_VERSION"/install.sh -o install-nvm.sh
sh install-nvm.sh

# In sourcegraph repository directory: install current recommendend version of Node JS
nvm install`,
	},
	{
		prompt:    `Do you want to use Docker or not?`,
		readsBool: `docker`,
	},
	{
		ifBool: "docker",
		prompt: "Nothing to do yet!",
		comment: `We provide a docker compose file at dev/redis-postgres.yml to make it easy to run Redis and PostgreSQL as docker containers.

NOTE: Although Ubuntu provides a docker-compose package, we recommend to install the latest version via pip so that it is compatible with our compose file.

See the official docker compose documentation at https://docs.docker.com/compose/install/ for more details on different installation options.
`,
	},
	// step 3, inserted here for convenience
	{
		ifBool:  "docker",
		prompt:  `The docker daemon might already be running, but if necessary you can use the following commands to start it:`,
		comment: `If you have issues running Docker, try adding your user to the docker group, and/or updating the socket file permissions, or try running these commands under sudo.`,
		command: `# as a system service
sudo systemctl enable --now docker

# manually
dockerd`,
	},
	{
		ifNotBool: "docker",
		prompt:    `Install PostgreSQL and Redis with the following commands`,
		command: `sudo apt install -y redis-server
sudo apt install -y postgresql postgresql-contrib`,
	},
	{
		ifNotBool: "docker",
		prompt:    `(optional) Start the services (and configure them to start automatically)`,
		command: `sudo systemctl enable --now postgresql
sudo systemctl enable --now redis-server.service`,
	},
	// step 4
	{
		ifBool: "docker",
		prompt: `To initialize your database, you may have to set the appropriate environment variables before running the createdb command:`,
		comment: `The Sourcegraph server reads PostgreSQL connection configuration from the PG* environment variables.

The development server startup script as well as the docker compose file provide default settings, so it will work out of the box.`,
		command: `export PGUSER=sourcegraph PGPASSWORD=sourcegraph PGDATABASE=sourcegraph
createdb --user=sourcegraph --owner=sourcegraph --encoding=UTF8 --template=template0 sourcegraph`,
	},
	{
		ifBool: "docker",
		prompt: `You can also use the PGDATA_DIR environment variable to specify a local folder (instead of a volume) to store the database files. See the dev/redis-postgres.yml file for more details.

This can also be spun up using sg run redis-postgres, with the following sg.config.override.yaml:`,
		command: `env:
    PGHOST: localhost
    PGPASSWORD: sourcegraph
    PGUSER: sourcegraph`,
	},
	{
		ifNotBool: "docker",
		prompt:    `Create a database for the current Unix user`,
		comment:   `You need a fresh Postgres database and a database user that has full ownership of that database.`,
		command: `sudo su - postgres
createdb`,
	},
	{
		ifNotBool: "docker",
		prompt:    `Create the Sourcegraph user and password`,
		command: `createuser --superuser sourcegraph
psql -c "ALTER USER sourcegraph WITH PASSWORD 'sourcegraph';"`,
	},
	{
		ifNotBool: "docker",
		prompt:    `Create the Sourcegraph database`,
		command:   `createdb --owner=sourcegraph --encoding=UTF8 --template=template0 sourcegraph`,
	},
	{
		ifNotBool: "docker",
		prompt:    `Configure database settings in your environment`,
		comment: `The Sourcegraph server reads PostgreSQL connection configuration from the PG* environment variables.

Our configuration in sg.config.yaml (we'll see what sg is in the next step) sets values that work with the setup described here, but if you are using different values you can overwrite them, for example, in your ~/.bashrc:`,
		command: `export PGPORT=5432
export PGHOST=localhost
export PGUSER=sourcegraph
export PGPASSWORD=sourcegraph
export PGDATABASE=sourcegraph
export PGSSLMODE=disable`,
	},
}

var codeInstructions = []instruction{
	{
		prompt:  `Run the following command in a folder where you want to keep a copy of the code. Command will create a new sub-folder (sourcegraph) in this folder.`,
		command: `git clone https://github.com/sourcegraph/sourcegraph.git`,
	},
	{
		prompt:    "Are you a Sourcegraph employee",
		readsBool: "employee",
	},
	{
		prompt: `In order to run the local development environment as a Sourcegraph employee, you'll need to clone another repository: sourcegraph/dev-private. It contains convenient preconfigured settings and code host connections.`,
		comment: `It needs to be cloned into the same folder as sourcegraph/sourcegraph, so they sit alongside each other. To illustrate:
/dir
 |-- dev-private
 +-- sourcegraph
NOTE: Ensure that you periodically pull the latest changes from sourcegraph/dev-private as the secrets are updated from time to time.`,
		command: `git clone https://github.com/sourcegraph/dev-private.git`,
		ifBool:  "employee",
	},
}

var httpReverseProxyInstructions = []instruction{
	{
		prompt: `Sourcegraph's development environment ships with a Caddy 2 HTTPS reverse proxy that allows you to access your local sourcegraph instance via https://sourcegraph.test:3443 (a fake domain with a self-signed certificate that's added to /etc/hosts).

If you'd like Sourcegraph to be accessible under https://sourcegraph.test (port 443) instead, you can set up authbind and set the environment variable SOURCEGRAPH_HTTPS_PORT=443.

 Prerequisites
In order to configure the HTTPS reverse-proxy, you'll need to edit /etc/hosts and initialize Caddy 2.

 Add sourcegraph.test to /etc/hosts
sourcegraph.test needs to be added to /etc/hosts as an alias to 127.0.0.1. There are two main ways of accomplishing this:

Manually append 127.0.0.1 sourcegraph.test to /etc/hosts
Use the provided ./dev/add_https_domain_to_hosts.sh convenience script (sudo may be required).`,
		command: `./dev/add_https_domain_to_hosts.sh`,
	},
	{
		prompt: `Initialize Caddy 2
Caddy 2 automatically manages self-signed certificates and configures your system so that your web browser can properly recognize them. The first time that Caddy runs, it needs root/sudo permissions to add its keys to your system's certificate store. You can get this out the way after installing Caddy 2 by running the following command and entering your password if prompted:`,
		comment: `(firefox users) If you are using Firefox and have a master password set, the following prompt will come up first:
Enter Password or Pin for "NSS Certificate DB":
Enter your Firefox master password here and proceed.

See this Github issue for more informations: https://github.com/FiloSottile/mkcert/issues/50
`,
		command: `./dev/caddy.sh trust`,
	},
}

func getBool() bool {
	var s string

	fmt.Printf("(y/N): ")
	_, err := fmt.Scan(&s)
	if err != nil {
		panic(err)
	}

	s = strings.TrimSpace(s)
	s = strings.ToLower(s)

	if s == "y" || s == "yes" {
		return true
	}
	return false
}
