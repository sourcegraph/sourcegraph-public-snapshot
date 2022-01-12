package main

import (
	"bufio"
	"context"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/gomodule/redigo/redis"
	"github.com/jackc/pgx/v4"
	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/internal/database/postgresdsn"
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

type userContextKey struct{}
type userContext struct {
	shellPath       string
	shellConfigPath string
}

func buildUserContext(userContext userContext, ctx context.Context) context.Context {
	return context.WithValue(ctx, userContextKey{}, userContext)
}

func getUserShellPath(ctx context.Context) string {
	v := ctx.Value(userContextKey{}).(userContext)
	return v.shellPath
}

func getUserShellConfigPath(ctx context.Context) string {
	v := ctx.Value(userContextKey{}).(userContext)
	return v.shellConfigPath
}

func setupExec(ctx context.Context, args []string) error {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "'sg setup' currently only supports macOS and Linux"))
		os.Exit(1)
	}

	currentOS := runtime.GOOS
	if overridesOS, ok := os.LookupEnv("SG_FORCE_OS"); ok {
		currentOS = overridesOS
	}

	// extract user environment general informations
	shellPath, shellConfigPath, err := guessUserShell()
	if err != nil {
		return err
	}
	userContext := userContext{shellPath: shellPath, shellConfigPath: shellConfigPath}
	ctx = buildUserContext(userContext, ctx)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			writeOrangeLinef("\n💡 You may need to restart your shell for the changes to work in this terminal.")
			writeOrangeLinef("   Close this terminal and open a new one or type the following command and press ENTER: " + filepath.Base(getUserShellPath(ctx)))
			os.Exit(0)
		}
	}()

	var categories []dependencyCategory
	if currentOS == "darwin" {
		categories = macOSDependencies
	} else {
		// DEPRECATED: The new 'sg setup' doesn't work on Linux yet, so we fall back to the old one.
		writeWarningLinef("'sg setup' on Linux provides instructions for Ubuntu Linux. If you're using another distribution, instructions might need to be adjusted.")
		return deprecatedSetupForLinux(ctx)
	}

	// Check whether we're in the sourcegraph/sourcegraph repository so we can
	// skip categories/dependencies that depend on the repository.
	_, err = root.RepositoryRoot()
	inRepo := err == nil

	failed := []int{}
	all := []int{}
	skipped := []int{}
	employeeFailed := []int{}
	for i := range categories {
		failed = append(failed, i)
		all = append(all, i)
	}

	for len(failed) != 0 {
		stdout.Out.ClearScreen()

		writeOrangeLinef("-------------------------------------")
		writeOrangeLinef("|        Welcome to sg setup!       |")
		writeOrangeLinef("-------------------------------------")
		writeOrangeLinef("Quit any time by typing ctrl-c\n")

		for i, category := range categories {
			idx := i + 1

			if category.requiresRepository && !inRepo {
				writeSkippedLinef("%d. %s %s[SKIPPED. Requires 'sg setup' to be run in 'sourcegraph' repository]%s", idx, category.name, output.StyleBold, output.StyleReset)
				skipped = append(skipped, idx)
				failed = removeEntry(failed, i)
				continue
			}

			pending := stdout.Out.Pending(output.Linef("", output.StylePending, "%d. %s - Determining status...", idx, category.name))
			category.Update(ctx)
			pending.Destroy()

			if combined := category.CombinedState(); combined {
				writeSuccessLinef("%d. %s", idx, category.name)
				failed = removeEntry(failed, i)
			} else {
				nonEmployeeState := category.CombinedStateNonEmployees()
				if nonEmployeeState {
					writeWarningLinef("%d. %s", idx, category.name)
					employeeFailed = append(skipped, idx)
				} else {
					writeFailureLinef("%d. %s", idx, category.name)
				}
			}
		}

		if len(failed) == 0 && len(employeeFailed) == 0 {
			if len(skipped) == 0 && len(employeeFailed) == 0 {
				stdout.Out.Write("")
				stdout.Out.WriteLine(output.Linef(output.EmojiOk, output.StyleBold, "Everything looks good! Happy hacking!"))
			}

			if len(skipped) != 0 {
				stdout.Out.Write("")
				writeWarningLinef("Some checks were skipped because 'sg setup' is not run in the 'sourcegraph' repository.")
				writeFingerPointingLinef("Restart 'sg setup' in the 'sourcegraph' repository to continue.")
			}

			return nil
		}

		stdout.Out.Write("")

		if len(employeeFailed) != 0 && len(failed) == len(employeeFailed) {
			writeWarningLinef("Some checks that are only relevant for Sourcegraph employees failed.\nIf you're not a Sourcegraph employee you're good to go. Hit Ctrl-C.\n\nIf you're a Sourcegraph employee: which one do you want to fix?")
		} else {
			writeWarningLinef("Some checks failed. Which one do you want to fix?")
		}

		idx, err := getNumberOutOf(all)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		selectedCategory := categories[idx]

		stdout.Out.ClearScreen()

		err = presentFailedCategoryWithOptions(ctx, idx, &selectedCategory)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}

	return nil
}

var macOSDependencies = []dependencyCategory{
	{
		name: "Install homebrew",
		dependencies: []*dependency{
			{
				name:  "brew",
				check: checkInPath("brew"),
				instructionsComment: `We depend on having the Homebrew package manager available on macOS.

Follow the instructions at https://brew.sh to install it, then rerun 'sg setup'.`,
			},
		},
		autoFixing: false,
	},
	{
		name: "Install base utilities (git, docker, ...)",
		dependencies: []*dependency{
			{name: "git", check: checkInPath("git"), instructionsCommands: `brew install git`},
			{name: "gnu-sed", check: checkInPath("gsed"), instructionsCommands: "brew install gnu-sed"},
			{name: "comby", check: checkInPath("comby"), instructionsCommands: "brew install comby"},
			{name: "pcre", check: checkInPath("pcregrep"), instructionsCommands: `brew install pcre`},
			{name: "sqlite", check: checkInPath("sqlite3"), instructionsCommands: `brew install sqlite`},
			{name: "jq", check: checkInPath("jq"), instructionsCommands: `brew install jq`},
			{name: "bash", check: checkCommandOutputContains("bash --version", "version 5"), instructionsCommands: `brew install bash`},
			{
				name:                 "docker",
				check:                wrapCheckErr(checkInPath("docker"), "if Docker is installed and the check fails, you might need to start Docker.app and restart terminal and 'sg setup'"),
				instructionsCommands: `brew install --cask docker`,
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
					return `brew install asdf && echo ". ${HOMEBREW_PREFIX:-/usr/local}/opt/asdf/libexec/asdf.sh" >> ` + getUserShellConfigPath(ctx)
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
asdf install golang
`,
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
				instructionsCommands: `
brew install gpg
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
				check: checkSourcegraphDatabase,
				instructionsComment: `` +
					`Once PostgreSQL is installed and running, we need to setup Sourcegraph database itself and a
specific user.`,
				instructionsCommands: `createuser --superuser sourcegraph || true
psql -c "ALTER USER sourcegraph WITH PASSWORD 'sourcegraph';" 
createdb --owner=sourcegraph --encoding=UTF8 --template=template0 sourcegraph
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
					We recommend installing it with Homebrew and starting it as a system service.`,
				instructionsCommands: "brew reinstall redis && brew services start redis",
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

func deprecatedSetupForLinux(ctx context.Context) error {
	var instructions []instruction
	instructions = append(instructions, linuxInstructionsBeforeClone...)
	// clone instructions come after dependency instructions because we need
	// `git` installed to `git` clone.
	instructions = append(instructions, cloneInstructions...)
	instructions = append(instructions, linuxInstructionsAfterClone...)
	instructions = append(instructions, httpReverseProxyInstructions...)

	conditions := map[string]bool{}

	i := 0
	for _, instruction := range instructions {
		if instruction.ifBool != "" {
			val, ok := conditions[instruction.ifBool]
			if !ok {
				stdout.Out.WriteLine(output.Line("", output.StyleWarning, "Something went wrong."))
				os.Exit(1)
			}
			if !val {
				continue
			}
		}
		if instruction.ifNotBool != "" {
			val, ok := conditions[instruction.ifNotBool]
			if !ok {
				stdout.Out.WriteLine(output.Line("", output.StyleWarning, "Something went wrong."))
				os.Exit(1)
			}
			if val {
				continue
			}
		}

		i++
		stdout.Out.WriteLine(output.Line("", output.StylePending, "------------------------------------------"))
		stdout.Out.Writef("%sStep %d:%s%s %s%s", output.StylePending, i, output.StyleReset, output.StyleSuccess, instruction.prompt, output.StyleReset)
		stdout.Out.Write("")

		if instruction.comment != "" {
			stdout.Out.Write(instruction.comment)
			stdout.Out.Write("")
		}

		if instruction.command != "" {
			stdout.Out.WriteLine(output.Line("", output.StyleSuggestion, "Run the following command(s) in another terminal:\n"))
			stdout.Out.WriteLine(output.Line("", output.CombineStyles(output.StyleBold, output.StyleYellow), strings.TrimSpace(instruction.command)))

			stdout.Out.WriteLine(output.Linef("", output.StyleSuggestion, "Hit return to confirm that you ran the command..."))
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
	{
		ifBool: "docker",
		prompt: `Even though you're going to run the database in docker you will probably want to install the CLI tooling for Redis and Postgres

redis-tools will provide redis-cli and postgresql will provide createdb and createuser`,
		command: `sudo apt install -y redis-tools postgresql postgresql-contrib`,
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
		command: `git clone git@github.com:sourcegraph/sourcegraph.git`,
	},
	{
		prompt:    "Are you a Sourcegraph employee?",
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
		command: `git clone git@github.com:sourcegraph/dev-private.git`,
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

func presentFailedCategoryWithOptions(ctx context.Context, categoryIdx int, category *dependencyCategory) error {
	printCategoryHeaderAndDependencies(categoryIdx+1, category)

	choices := map[int]string{1: "I want to fix these manually"}
	if category.autoFixing {
		choices[2] = "I'm feeling lucky. You try fixing all of it for me."
		choices[3] = "Go back"
	} else {
		choices[2] = "Go back"
	}

	choice, err := getChoice(choices)
	if err != nil {
		return err
	}

	switch choice {
	case 1:
		err = fixCategoryManually(ctx, categoryIdx, category)
	case 2:
		stdout.Out.ClearScreen()
		err = fixCategoryAutomatically(ctx, category)
	case 3:
		return nil
	}
	return err
}

func printCategoryHeaderAndDependencies(categoryIdx int, category *dependencyCategory) {
	stdout.Out.WriteLine(output.Linef(output.EmojiLightbulb, output.CombineStyles(output.StyleSearchQuery, output.StyleBold), "%d. %s", categoryIdx, category.name))
	stdout.Out.Write("")
	stdout.Out.Write("Checks:")

	for i, dep := range category.dependencies {
		idx := i + 1
		if dep.IsMet() {
			writeSuccessLinef("%d. %s", idx, dep.name)
		} else {
			var printer func(fmtStr string, args ...interface{})
			if dep.onlyEmployees {
				printer = writeWarningLinef
			} else {
				printer = writeFailureLinef
			}

			if dep.err != nil {
				printer("%d. %s: %s", idx, dep.name, dep.err)
			} else {
				printer("%d. %s: %s", idx, dep.name, "check failed")
			}
		}
	}
}

func fixCategoryAutomatically(ctx context.Context, category *dependencyCategory) error {
	// for go through sub dependencies that may be required to fix the dependencies themselves.
	for _, dep := range category.autoFixingDependencies {
		if dep.IsMet() {
			continue
		}
		if err := fixDependencyAutomatically(ctx, dep); err != nil {
			return err
		}
	}
	// now go through the real dependencies
	for _, dep := range category.dependencies {
		if dep.IsMet() {
			continue
		}

		if err := fixDependencyAutomatically(ctx, dep); err != nil {
			return err
		}
	}

	return nil
}

func fixDependencyAutomatically(ctx context.Context, dep *dependency) error {
	writeFingerPointingLinef("Trying my hardest to fix %q automatically...", dep.name)

	cmd := execFreshShell(ctx, dep.InstructionsCommands(ctx))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		writeFailureLinef("Failed to run command: %s", err)
		return err
	}

	writeSuccessLinef("Done! %q should be fixed now!", dep.name)

	if dep.requiresSgSetupRestart {
		writeFingerPointingLinef("This command requires restarting of 'sg setup' to pick up the changes.")
		os.Exit(0)
	}

	return nil
}

func fixCategoryManually(ctx context.Context, categoryIdx int, category *dependencyCategory) error {
	for {
		toFix := []int{}

		for i, dep := range category.dependencies {
			if dep.IsMet() {
				continue
			}

			toFix = append(toFix, i)
		}

		if len(toFix) == 0 {
			break
		}

		var idx int

		if len(toFix) == 1 {
			idx = toFix[0]
		} else {
			writeFingerPointingLinef("Which one do you want to fix?")
			var err error
			idx, err = getNumberOutOf(toFix)
			if err != nil {
				if err == io.EOF {
					return nil
				}
				return err
			}
		}

		dep := category.dependencies[idx]

		stdout.Out.WriteLine(output.Linef(output.EmojiFailure, output.CombineStyles(output.StyleWarning, output.StyleBold), "%s", dep.name))
		stdout.Out.Write("")

		if dep.err != nil {
			stdout.Out.WriteLine(output.Linef("", output.StyleBold, "Encountered the following error:\n\n%s%s\n", output.StyleReset, dep.err))
		}

		stdout.Out.WriteLine(output.Linef("", output.StyleBold, "How to fix:"))

		if dep.instructionsComment != "" {
			stdout.Out.Write("")
			stdout.Out.Write(dep.instructionsComment)
		}

		// If we don't have anything do run, we simply print instructions to
		// the user
		if dep.InstructionsCommands(ctx) == "" {
			writeFingerPointingLinef("Hit return once you're done")
			waitForReturn()
		} else {
			// Otherwise we print the command(s) and ask the user whether we should run it or not
			stdout.Out.Write("")
			if category.requiresRepository {
				stdout.Out.Writef("Run the following command(s) %sin the 'sourcegraph' repository%s:", output.StyleBold, output.StyleReset)
			} else {
				stdout.Out.Write("Run the following command(s):")
			}
			stdout.Out.Write("")

			stdout.Out.WriteLine(output.Line("", output.CombineStyles(output.StyleBold, output.StyleYellow), strings.TrimSpace(dep.InstructionsCommands(ctx))))

			choice, err := getChoice(map[int]string{
				1: "I'll fix this manually (either by running the command or doing something else)",
				2: "You can run the command for me",
				3: "Go back",
			})
			if err != nil {
				return err
			}

			switch choice {
			case 1:
				writeFingerPointingLinef("Hit return once you're done")
				waitForReturn()
			case 2:
				if err := fixDependencyAutomatically(ctx, dep); err != nil {
					return err
				}
			case 3:
				return nil
			}
		}

		pending := stdout.Out.Pending(output.Linef("", output.StylePending, "Determining status..."))
		for _, dep := range category.dependencies {
			dep.Update(ctx)
		}
		pending.Destroy()

		printCategoryHeaderAndDependencies(categoryIdx, category)
	}

	return nil
}

func removeEntry(s []int, val int) (result []int) {
	for _, e := range s {
		if e != val {
			result = append(result, e)
		}
	}
	return result
}

func checkCommandOutputContains(cmd, contains string) func(context.Context) error {
	return func(ctx context.Context) error {
		out, _ := combinedSourceExec(ctx, cmd)
		if !strings.Contains(string(out), contains) {
			return errors.Newf("command output of %q doesn't contain %q", cmd, contains)
		}
		return nil
	}
}

func checkFileContains(fileName, content string) func(context.Context) error {
	return func(ctx context.Context) error {
		file, err := os.Open(fileName)
		if err != nil {
			return errors.Wrapf(err, "failed to check that %q contains %q", fileName, content)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, content) {
				return nil
			}
		}

		if err := scanner.Err(); err != nil {
			return err
		}

		return errors.Newf("file %q did not contain %q", fileName, content)
	}
}

func guessUserShell() (string, string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", err
	}
	// Look up which shell the user is using, because that's most likely the
	// one that has all the environment correctly setup.
	shell, ok := os.LookupEnv("SHELL")
	var shellrc string
	if !ok {
		// If we can't find the shell in the environment, we fall back to `bash`
		shell = "bash"
	}
	switch {
	case strings.Contains(shell, "bash"):
		shellrc = ".bashrc"
	case strings.Contains(shell, "zsh"):
		shellrc = ".zshrc"
	}
	return shell, filepath.Join(home, shellrc), nil
}

// execFreshShell returns a command wrapped in a new shell process, enabling
// changes added by various checks to be run. This negates the new to ask the
// user to restart sg for many checks.
func execFreshShell(ctx context.Context, cmd string) *exec.Cmd {
	command := fmt.Sprintf("source %s || true; %s", getUserShellConfigPath(ctx), cmd)
	return exec.CommandContext(ctx, getUserShellPath(ctx), "-c", command)
}

// combinedSourceExec runs a command in a fresh shell environment,
// and returns stderr and stdout combined, along with an error.
func combinedSourceExec(ctx context.Context, cmd string) ([]byte, error) {
	return execFreshShell(ctx, cmd).CombinedOutput()
}

func checkInPath(cmd string) func(context.Context) error {
	return func(ctx context.Context) error {
		hashCmd := fmt.Sprintf("hash %s 2>/dev/null", cmd)
		_, err := combinedSourceExec(ctx, hashCmd)
		if err != nil {
			return errors.Newf("executable %q not found in $PATH", cmd)
		}
		return nil
	}
}

func checkInMainRepoOrRepoInDirectory(ctx context.Context) error {
	_, err := root.RepositoryRoot()
	if err != nil {
		ok, err := pathExists("sourcegraph")
		if !ok || err != nil {
			return errors.New("'sg setup' is not run in sourcegraph and repository is also not found in current directory")
		}
		return nil
	}
	return nil
}

func checkDevPrivateInParentOrInCurrentDirectory(context.Context) error {
	ok, err := pathExists("dev-private")
	if ok && err == nil {
		return nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "failed to check for dev-private repository")
	}

	p := filepath.Join(wd, "..", "dev-private")
	ok, err = pathExists(p)
	if ok && err == nil {
		return nil
	}
	return errors.New("could not find dev-private repository either in current directory or one above")
}

// checkPostgresConnection succeeds connecting to the default user database works, regardless
// of if it's running locally or with docker.
func checkPostgresConnection(ctx context.Context) error {
	dsns, err := dsnCandidates()
	if err != nil {
		return err
	}
	var errs []error
	for _, dsn := range dsns {
		conn, err := pgx.Connect(ctx, dsn)
		if err != nil {
			errs = append(errs, errors.Wrapf(err, "failed to connect to Postgresql Database at %s", dsn))
			continue
		}
		defer conn.Close(ctx)
		err = conn.Ping(ctx)
		if err == nil {
			// if ping passed
			return nil
		}
		errs = append(errs, errors.Wrapf(err, "failed to connect to Postgresql Database at %s", dsn))
	}

	messages := []string{"failed all attempts to connect to Postgresql database"}
	for _, e := range errs {
		messages = append(messages, "\t"+e.Error())
	}
	return errors.New(strings.Join(messages, "\n"))
}

func dsnCandidates() ([]string, error) {
	env := func(key string) string { val, _ := os.LookupEnv(key); return val }

	// best case scenario
	datasource := env("PGDATASOURCE")
	// most classic dsn
	baseURL := url.URL{Scheme: "postgres", Host: "127.0.0.1:5432"}
	// classic docker dsn
	dockerURL := baseURL
	dockerURL.User = url.UserPassword("postgres", "postgres")
	// other classic docker dsn
	dockerURL2 := baseURL
	dockerURL2.User = url.UserPassword("postgres", "password")
	// env based dsn
	envURL := baseURL
	username, ok := os.LookupEnv("PGUSER")
	if !ok {
		uinfo, err := user.Current()
		if err != nil {
			return nil, err
		}
		username = uinfo.Name
	}
	envURL.User = url.UserPassword(username, env("PGPASSWORD"))
	if host, ok := os.LookupEnv("PGHOST"); ok {
		if port, ok := os.LookupEnv("PGPORT"); ok {
			envURL.Host = fmt.Sprintf("%s:%s", host, port)
		}
		envURL.Host = fmt.Sprintf("%s:%s", host, "5432")
	}
	if sslmode := env("PGSSLMODE"); sslmode != "" {
		qry := envURL.Query()
		qry.Set("sslmode", sslmode)
		envURL.RawQuery = qry.Encode()
	}
	return []string{
		datasource,
		envURL.String(),
		baseURL.String(),
		dockerURL.String(),
		dockerURL2.String(),
	}, nil
}

func checkSourcegraphDatabase(ctx context.Context) error {
	// This check runs only in the `sourcegraph/sourcegraph` repository, so
	// we try to parse the globalConf and use its `Env` to configure the
	// Postgres connection.
	ok, _ := parseConf(*configFlag, *overwriteConfigFlag)
	if !ok {
		return errors.New("failed to read sg.config.yaml. This step of `sg setup` needs to be run in the `sourcegraph` repository")
	}

	getEnv := func(key string) string {
		// First look into process env, emulating the logic in makeEnv used
		// in internal/run/run.go
		val, ok := os.LookupEnv(key)
		if ok {
			return val
		}
		// Otherwise check in globalConf.Env
		return globalConf.Env[key]
	}

	dsn := postgresdsn.New("", "", getEnv)
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return errors.Wrapf(err, "failed to connect to Soucegraph Postgres database at %s. Please check the settings in sg.config.yml (see https://docs.sourcegraph.com/dev/background-information/sg#changing-database-configuration)", dsn)
	}
	defer conn.Close(ctx)
	return conn.Ping(ctx)
}

func checkRedisConnection(context.Context) error {
	conn, err := redis.Dial("tcp", ":6379", redis.DialConnectTimeout(5*time.Second))
	if err != nil {
		return errors.Wrap(err, "failed to connect to Redis at 127.0.0.1:6379")
	}

	if _, err := conn.Do("SET", "sg-setup", "was-here"); err != nil {
		return errors.Wrap(err, "failed to write to Redis at 127.0.0.1:6379")
	}

	retval, err := redis.String(conn.Do("GET", "sg-setup"))
	if err != nil {
		return errors.Wrap(err, "failed to read from Redis at 127.0.0.1:6379")
	}

	if retval != "was-here" {
		return errors.New("failed to test write in Redis")
	}
	return nil
}

type dependencyCheck func(context.Context) error

type dependency struct {
	name string

	check dependencyCheck

	onlyEmployees bool

	err error

	instructionsComment         string
	instructionsCommands        string
	instructionsCommandsBuilder commandBuilder
	requiresSgSetupRestart      bool
}

func (d *dependency) IsMet() bool { return d.err == nil }

func (d *dependency) InstructionsCommands(ctx context.Context) string {
	if d.instructionsCommandsBuilder != nil {
		return d.instructionsCommandsBuilder.Build(ctx)
	}
	return d.instructionsCommands
}

func (d *dependency) Update(ctx context.Context) {
	d.err = nil
	d.err = d.check(ctx)
}

type dependencyCategory struct {
	name         string
	dependencies []*dependency

	autoFixing             bool
	autoFixingDependencies []*dependency
	requiresRepository     bool
}

func (cat *dependencyCategory) CombinedState() bool {
	for _, dep := range cat.dependencies {
		if !dep.IsMet() {
			return false
		}
	}
	return true
}

func (cat *dependencyCategory) CombinedStateNonEmployees() bool {
	for _, dep := range cat.dependencies {
		if !dep.IsMet() && !dep.onlyEmployees {
			return false
		}
	}
	return true
}

func (cat *dependencyCategory) Update(ctx context.Context) {
	for _, dep := range cat.autoFixingDependencies {
		dep.Update(ctx)
	}
	for _, dep := range cat.dependencies {
		dep.Update(ctx)
	}
}

func getNumberOutOf(numbers []int) (int, error) {
	var strs []string
	var idx = make(map[int]struct{})
	for _, num := range numbers {
		strs = append(strs, fmt.Sprintf("%d", num+1))
		idx[num+1] = struct{}{}
	}

	for {
		fmt.Printf("[%s]: ", strings.Join(strs, ","))
		var num int
		_, err := fmt.Scan(&num)
		if err != nil {
			return 0, err
		}

		if _, ok := idx[num]; ok {
			return num - 1, nil
		}
		fmt.Printf("%d is an invalid choice :( Let's try again?\n", num)
	}
}

func waitForReturn() { fmt.Scanln() }

func getChoice(choices map[int]string) (int, error) {
	for {
		stdout.Out.Write("")
		writeFingerPointingLinef("What do you want to do?")

		for i := 0; i < len(choices); i++ {
			num := i + 1
			desc, ok := choices[num]
			if !ok {
				return 0, errors.Newf("internal error: %d not found in provided choices", i)
			}
			stdout.Out.Writef("%s[%d]%s: %s", output.StyleBold, num, output.StyleReset, desc)
		}

		fmt.Printf("Enter choice: ")

		var s int
		_, err := fmt.Scan(&s)
		if err != nil {
			return 0, err
		}

		if _, ok := choices[s]; ok {
			return s, nil
		}
		writeFailureLinef("Invalid choice")
	}
}

func anyChecks(checks ...dependencyCheck) dependencyCheck {
	return func(ctx context.Context) (err error) {
		for _, chk := range checks {
			err = chk(ctx)
			if err == nil {
				return nil
			}
		}
		return err
	}
}

func combineChecks(checks ...dependencyCheck) dependencyCheck {
	return func(ctx context.Context) (err error) {
		for _, chk := range checks {
			err = chk(ctx)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func retryCheck(check dependencyCheck, retries int, sleep time.Duration) dependencyCheck {
	return func(ctx context.Context) (err error) {
		for i := 0; i < retries; i++ {
			err = check(ctx)
			if err == nil {
				return nil
			}
			time.Sleep(sleep)
		}
		return err
	}
}

func wrapCheckErr(check dependencyCheck, message string) dependencyCheck {
	return func(ctx context.Context) error {
		err := check(ctx)
		if err != nil {
			return errors.Wrap(err, message)
		}
		return nil
	}
}

func checkCaddyTrusted(ctx context.Context) error {
	certPath, err := caddySourcegraphCertificatePath()
	if err != nil {
		return errors.Wrap(err, "failed to determine path where proxy stores certificates")
	}

	ok, err := pathExists(certPath)
	if !ok || err != nil {
		return errors.New("sourcegraph.test certificate not found. highly likely it's not trusted by system")
	}

	rawCert, err := os.ReadFile(certPath)
	if err != nil {
		return errors.Wrap(err, "could not read certificate")
	}

	cert, err := pemDecodeSingleCert(rawCert)
	if err != nil {
		return errors.Wrap(err, "decoding cert failed")
	}

	if trusted(cert) {
		return nil
	}
	return errors.New("doesn't look like certificate is trusted")
}

// caddyAppDataDir returns the location of the sourcegraph.test certificate
// that Caddy created or would create.
//
// It's copy&pasted&modified from here: https://sourcegraph.com/github.com/caddyserver/caddy@9ee68c1bd57d72e8a969f1da492bd51bfa5ed9a0/-/blob/storage.go?L114
func caddySourcegraphCertificatePath() (string, error) {
	if basedir := os.Getenv("XDG_DATA_HOME"); basedir != "" {
		return filepath.Join(basedir, "caddy"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	var appDataDir string
	switch runtime.GOOS {
	case "darwin":
		appDataDir = filepath.Join(home, "Library", "Application Support", "Caddy")
	case "linux":
		appDataDir = filepath.Join(home, ".local", "share", "caddy")
	default:
		return "", errors.Newf("unsupported OS: %s", runtime.GOOS)
	}

	return filepath.Join(appDataDir, "pki", "authorities", "local", "root.crt"), nil
}

func trusted(cert *x509.Certificate) bool {
	chains, err := cert.Verify(x509.VerifyOptions{})
	return len(chains) > 0 && err == nil
}

func pemDecodeSingleCert(pemDER []byte) (*x509.Certificate, error) {
	pemBlock, _ := pem.Decode(pemDER)
	if pemBlock == nil {
		return nil, fmt.Errorf("no PEM block found")
	}
	if pemBlock.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("expected PEM block type to be CERTIFICATE, but got '%s'", pemBlock.Type)
	}
	return x509.ParseCertificate(pemBlock.Bytes)
}

type commandBuilder interface {
	Build(context.Context) string
}

type stringCommandBuilder func(context.Context) string

func (l stringCommandBuilder) Build(ctx context.Context) string {
	return l(ctx)
}
