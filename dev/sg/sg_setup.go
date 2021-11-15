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

	currentOS := runtime.GOOS
	if overridesOS, ok := os.LookupEnv("SG_FORCE_OS"); ok {
		currentOS = overridesOS
	}

	var instructions []instruction
	if currentOS == "darwin" {
		instructions = append(instructions, macOSInstructionsBeforeClone...)
	} else {
		instructions = append(instructions, linuxInstructionsBeforeClone...)
	}

	// clone instructions come after dependency instructions because we need
	// `git` installed to `git` clone.
	instructions = append(instructions, cloneInstructions...)

	if currentOS == "darwin" {
		instructions = append(instructions, macOSInstructionsAfterClone...)
	} else {
		instructions = append(instructions, linuxInstructionsAfterClone...)
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
		out.Writef("%sStep %d:%s%s %s%s", output.StylePending, i, output.StyleReset, output.StyleSuccess, instruction.prompt, output.StyleReset)
		out.Write("")

		if instruction.comment != "" {
			out.Write(instruction.comment)
			out.Write("")
		}

		if instruction.command != "" {
			out.WriteLine(output.Line("", output.StyleSuggestion, "Run the following command(s) in another terminal:\n"))
			out.WriteLine(output.Line("", output.CombineStyles(output.StyleBold, output.StyleYellow), strings.TrimSpace(instruction.command)))

			out.WriteLine(output.Linef("", output.StyleSuggestion, "Hit return to confirm that you ran the command..."))
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

var macOSInstructionsBeforeClone = []instruction{
	{
		comment: `Homewbrew is a tool to install programs on your machine that is very common on OSX.`,
		prompt:  "Install homebrew",
		command: `/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`,
	},
	{
		prompt: `Install Docker`,
		comment: `The following command installs Docker as a macOS application.

Make sure to start /Applications/Docker.app after the installation finished`,
		command: `brew install --cask docker`,
	},
	{
		prompt:  `Install Git, Comby, SQLite tools, and jq`,
		command: `brew install git gnu-sed comby pcre sqlite jq`,
	},
}

var macOSInstructionsAfterClone = []instruction{
	{
		prompt:  `Install Go, Yarn, Node`,
		comment: `Instead of homebrew you can also use asdf to install Go, Yarn, Node.js. See the install instructions for asdf here: https://asdf-vm.com/guide/getting-started.html See the .tool-versions file in the sourcegraph/sourcegraph repository later.`,
		command: `brew install go yarn node`,
	},
	{
		prompt:    `Do you want to use Docker to run PostgreSQL and Redis?`,
		comment:   `If you don't know, we recommend that you answer with 'No'`,
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
		command:   `brew install postgres redis`,
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
		prompt:    `Ensure psql, the PostgreSQL command line client, is available.`,
		comment:   `If this command prints "OK", you are free to move to the next step. Otherwise, you installed a Homebrew recipe that does not modify your $PATH by default. If not the next step will address that`,
		command:   `(hash psql && echo "OK") || echo "NOT OK"`,
	},
	{
		ifNotBool: "docker",
		prompt:    `Ensure psql, the PostgreSQL command line client, is available`,
		comment:   `If the previous command printed "NOT OK", you can run the command below to fix that`,
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
curl -L https://raw.githubusercontent.com/nvm-sh/nvm/"$NVM_VERSION"/install.sh -o /tmp/install-nvm.sh
sh /tmp/install-nvm.sh`,
	},
	{
		prompt: `After the NVM installation finished, run the following command to activate it in the current terminal without restarting:`,
		command: `export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"`,
	},
	{
		prompt: `Install the current recommended version of Node JS by running the following in the sourcegraph/sourcegraph repository clone`,
		comment: `After doing this, node -v should show the same version mentioned in .nvmrc at the root of the sourcegraph repository.
NOTE: Although there is a Homebrew package for Node, we advise using nvm instead, to ensure you get a Node version compatible with the current state of the sourcegraph repository.`,
		command: `# Run the following two commands in the 'sourcegraph' repository:
nvm install
nvm use --delete-prefix`,
	},
	// step 4
	{
		ifBool: "docker",
		prompt: `To initialize your database, you may have to set the appropriate environment variables before running the createdb command:`,
		comment: `The Sourcegraph server reads PostgreSQL connection configuration from the PG* environment variables.

The development server startup script as well as the docker compose file provide default settings, so it will work out of the box.`,
		command: `createdb --user=sourcegraph --owner=sourcegraph --host=localhost --encoding=UTF8 --template=template0 sourcegraph`,
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
}

var linuxInstructionsBeforeClone = []instruction{
	{
		prompt:  `Update repositories`,
		command: `sudo apt-get update`,
	},
	{
		prompt:  `Install dependencies`,
		command: `sudo apt install -y make git-all libpcre3-dev libsqlite3-dev pkg-config jq libnss3-tools`,
	},
}

var linuxInstructionsAfterClone = []instruction{
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
		prompt:    `Do you want to use Docker to run PostgreSQL and Redis?`,
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
		command: `createdb --user=sourcegraph --owner=sourcegraph --host=localhost --encoding=UTF8 --template=template0 sourcegraph`,
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
}

var cloneInstructions = []instruction{
	{
		prompt:  `Cloning the code`,
		comment: `We're now going to clone the Sourcegraph repository. Make sure you execute the following command in a folder where you want to keep the repository. Command will create a new sub-folder (sourcegraph) in this folder.`,
		command: `git clone https://github.com/sourcegraph/sourcegraph.git`,
	},
	{
		prompt:    "Are you a Sourcegraph employee",
		readsBool: "employee",
	},
	{
		prompt: `Getting access to private resources`,
		comment: `In order to run the local development environment as a Sourcegraph employee, you'll need to clone another repository: sourcegraph/dev-private. It contains convenient preconfigured settings and code host connections.
It needs to be cloned into the same folder as sourcegraph/sourcegraph, so they sit alongside each other.
To illustrate:
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
		prompt: `Making sourcegraph.test accessible`,
		comment: `In order to make Sourcegraph's development environment accessible under https://sourcegraph.test:3443 we need to add an entry to /etc/hosts.

The following command will add this entry. It may prompt you for your password.

Execute it in the 'sourcegraph' repository you cloned.`,
		command: `./dev/add_https_domain_to_hosts.sh`,
	},
	{
		prompt: `Initialize Caddy 2`,
		comment: `Caddy 2 automatically manages self-signed certificates and configures your system so that your web browser can properly recognize them.

The following command adds Caddy's keys to the system certificate store.

Execute it in the 'sourcegraph' repository you cloned.`,
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
