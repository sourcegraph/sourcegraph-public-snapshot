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

	if runtime.GOOS == "darwin" {
		instructions = macOSInstructions
	} else {
		instructions = linuxInstructions
	}

	conditions := map[string]bool{}

	for i, instruction := range instructions {
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
