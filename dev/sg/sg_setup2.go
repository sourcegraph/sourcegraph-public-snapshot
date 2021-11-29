package main

import (
	"bufio"
	"context"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/garyburd/redigo/redis"
	"github.com/jackc/pgx/v4"
	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/output"
	"github.com/sourcegraph/sourcegraph/lib/postgresdsn"
)

var (
	setup2FlagSet = flag.NewFlagSet("sg setup", flag.ExitOnError)
	setup2Command = &ffcli.Command{
		Name:       "setup2",
		ShortUsage: "sg setup2",
		LongHelp:   "Run 'sg setup2' to setup the local dev environment",
		FlagSet:    setup2FlagSet,
		Exec:       setup2Exec,
	}
)

func setup2Exec(ctx context.Context, args []string) error {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		out.WriteLine(output.Linef("", output.StyleWarning, "'sg setup' currently only supports macOS and Linux"))
		os.Exit(1)
	}

	currentOS := runtime.GOOS
	if overridesOS, ok := os.LookupEnv("SG_FORCE_OS"); ok {
		currentOS = overridesOS
	}

	var categories []dependencyCategory
	if currentOS == "darwin" {
		categories = macOSDependencies
	} else {
		// TODO: Support Linux
		return errors.Newf("sg setup currently does not support %s", currentOS)
	}

	// Check whether we're in the sourcegraph/sourcegraph repository so we can
	// skip categories/dependencies that depend on the repository.
	_, err := root.RepositoryRoot()
	inRepo := err == nil

	failed := []int{}
	all := []int{}
	skipped := []int{}
	for i := range categories {
		failed = append(failed, i)
		all = append(all, i)
	}

	for len(failed) != 0 {
		out.ClearScreen()

		writeOrangeLine("-------------------------------------")
		writeOrangeLine("|        Welcome to sg setup!       |")
		writeOrangeLine("-------------------------------------")

		for i, category := range categories {
			idx := i + 1

			if category.requiresRepository && !inRepo {
				writeSkippedLine("%d. %s %s[SKIPPED. Requires 'sg setup' to be run in 'sourcegraph' repository]%s", idx, category.name, output.StyleBold, output.StyleReset)
				skipped = append(skipped, idx)
				failed = removeEntry(failed, i)
				continue
			}

			pending := out.Pending(output.Linef("", output.StylePending, "%d. %s - Determining status...", idx, category.name))
			for _, dep := range category.dependencies {
				dep.Update(ctx)
			}
			pending.Destroy()

			if combined := category.CombinedState(); combined {
				writeSuccessLine("%d. %s", idx, category.name)
				failed = removeEntry(failed, i)
			} else {
				writeFailureLine("%d. %s", idx, category.name)
			}
		}

		if len(failed) == 0 {
			if len(skipped) == 0 {
				out.Write("")
				out.WriteLine(output.Linef(output.EmojiOk, output.StyleBold, "Everything looks good! Happy hacking!"))
				return nil
			} else {
				out.Write("")
				out.WriteLine(output.Linef(output.EmojiWarningSign, output.StyleYellow, "Some checks were skipped because 'sg setup' is not run in the 'sourcegraph' repository."))
				writeFingerPointingLine("Restart 'sg setup' in the 'sourcegraph' repository to continue.")
				return nil
			}
		}

		out.Write("")
		out.WriteLine(output.Linef(output.EmojiWarningSign, output.StyleYellow, "Some checks failed. Which one do you want to fix?"))

		idx, err := getNumberOutOf(all)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		selectedCategory := categories[idx]

		out.ClearScreen()

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
				// instructionsCommands: `/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`,
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

See here on how to set that up: https://docs.github.com/en/authentication/connecting-to-github-with-ssh`,
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
					`In order to run the local development environment as a Sourcegraph employee, you'll need to clone another repository: github.com/sourcegraph/dev-private. It contains convenient preconfigured settings and code host connections.

It needs to be cloned into the same folder as sourcegraph/sourcegraph, so they sit alongside each other.
To illustrate:
   /dir
   |-- dev-private
   +-- sourcegraph
NOTE: Ensure that you periodically pull the latest changes from sourcegraph/dev-private as the secrets are updated from time to time.`,
				onlyEmployees: true,
			},
		},
		autoFixing: true,
	},
	{
		name:               "Programming languages & tooling",
		requiresRepository: true,
		dependencies: []*dependency{
			// TODO: install asdf
			{
				name: "go", check: checkInPath("go"),
				instructionsComment: `` +
					`Souregraph requires Go to be installed.

Check the .tool-versions file for which version.

We *highly recommend* using the asdf version manager to install and manage
programming languages and tools.

Find out how to install asdf here: https://asdf-vm.com/guide/getting-started.html

Once you have asdf, execute the commands below **in the sourcegraph repository**.`,
				instructionsCommands: `
asdf plugin-add golang https://github.com/kennyp/asdf-golang.git
asdf install golang
`,
			},
			{
				name: "yarn", check: checkInPath("yarn"),
				instructionsComment: `` +
					`Souregraph requires Yarn to be installed.

Check the .tool-versions file for which version.

We *highly recommend* using the asdf version manager to install and manage
programming languages and tools.

Find out how to install asdf here: https://asdf-vm.com/guide/getting-started.html

Once you have asdf, execute the commands below **in the sourcegraph repository**.`,
				instructionsCommands: `
asdf plugin-add yarn
asdf install yarn 
`,
			},
			{
				name:  "node",
				check: checkInPath("node"),
				instructionsComment: `` +
					`Souregraph requires Node.JS to be installed.

Check the .tool-versions file for which version.

We *highly recommend* using the asdf version manager to install and manage
programming languages and tools.

Find out how to install asdf here: https://asdf-vm.com/guide/getting-started.html

Once you have asdf, execute the commands below **in the sourcegraph repository**.`,
				instructionsCommands: `
asdf plugin add nodejs https://github.com/asdf-vm/asdf-nodejs.git 
asdf install node 
`,
			},
		},
		// TODO: customAllInOnecommand
		// - install asdf
		// - reload asdf
		// - check for sourcegraph repository
		// - go into sourcegraph repository
		// - run the other commands

	},
	{
		name:               "Setup PostgreSQL database",
		requiresRepository: true,
		dependencies: []*dependency{
			{
				name:  "Connection to 'sourcegraph' database",
				check: checkPostgresConnection,
				instructionsComment: `` +
					`Sourcegraph requires the PostgreSQL database to be running. We recommend installing it with Homebrew and starting it as a system service.

If you know what you're doing, you can also install PostgreSQL another way.

Alternative 1: Installing Postgres.app and following instructions at https://postgresapp.com/

If you're not sure: use the recommended commands to install PostgreSQL and start it`,
				instructionsCommands: "brew reinstall postgresql && brew services start postgresql",
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
		requiresRepository: true,
		dependencies: []*dependency{
			{
				name:  "Connection to Redis",
				check: retryCheck(checkRedisConnection, 5, 500*time.Millisecond),
				instructionsComment: `` +
					`Sourcegraph requires the Redis database to be running. We recommend installing it with Homebrew and starting it as a system service.`,
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
					`Sourcegraph should be reachable under https://sourcegraph.test:3443. To do that, we need to add sourcegraph.test to the /etc/hosts file.

The command needs to be run inside the 'sourcegraph' repository.`,
				instructionsCommands: `./dev/add_https_domain_to_hosts.sh`,
			},
			{
				name:  "Caddy root certificate is trusted by system",
				check: checkCaddyTrusted,
				instructionsComment: `` +
					`In order to use TLS to access your local Sourcegraph instance, you need to trust the certificate created by Caddy, the proxy we use locally.

The command needs to be run inside the 'sourcegraph' repository.

YOU NEED TO RESTART 'sg setup' AFTER RUNNING THIS COMMAND!`,
				instructionsCommands:   `./dev/caddy.sh trust`,
				requiresSgSetupRestart: true,
			},
		},
	},
}

func presentFailedCategoryWithOptions(ctx context.Context, categoryIdx int, category *dependencyCategory) error {
	printCategoryHeaderAndDependencies(categoryIdx, category)

	// TODO: It doesn't make a lot of sense to give a choice here if
	// there's only one dependency

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
		err = fixCategoryManually(ctx, category)
	case 2:
		out.ClearScreen()
		err = fixCategoryAutomatically(ctx, category)
	case 3:
		return nil
	}
	return err
}

func printCategoryHeaderAndDependencies(categoryIdx int, category *dependencyCategory) {
	out.WriteLine(output.Linef(output.EmojiLightbulb, output.CombineStyles(output.StyleSearchQuery, output.StyleBold), "%d. %s", categoryIdx, category.name))
	out.Write("")
	out.Write("Checks:")

	for i, dep := range category.dependencies {
		idx := i + 1
		if dep.IsMet() {
			writeSuccessLine("%d. %s", idx, dep.name)
		} else {
			if dep.err != nil {
				writeFailureLine("%d. %s: %s", idx, dep.name, dep.err)
			} else {
				writeFailureLine("%d. %s: %s", idx, dep.name, "check failed")
			}
		}
	}
}

func fixCategoryAutomatically(ctx context.Context, category *dependencyCategory) error {
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
	// TODO: Would it be better if we show the output of the commands here?
	pending := out.Pending(output.Linef("", output.StylePending, "Trying my hardest to fix %q automatically...", dep.name))

	// TODO: Instead of bash we should probably use the users shell?
	cmdOut, err := exec.CommandContext(ctx, "bash", "-c", dep.instructionsCommands).CombinedOutput()
	if err != nil {
		pending.WriteLine(output.Linef(output.EmojiFailure, output.StyleWarning, "failed to run command: %s\n\noutput: %s", err, cmdOut))
		return err
	}

	pending.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Done! %q should be fixed now!", dep.name))

	if dep.requiresSgSetupRestart {
		writeFingerPointingLine("This command requires restarting of 'sg setup' to pick up the changes.")
		os.Exit(0)
	}

	return nil
}

func fixCategoryManually(ctx context.Context, category *dependencyCategory) error {
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

		writeFingerPointingLine("Which one do you want to fix?")
		idx, err := getNumberOutOf(toFix)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		dep := category.dependencies[idx]

		out.WriteLine(output.Linef(output.EmojiFailure, output.CombineStyles(output.StyleWarning, output.StyleBold), "%s", dep.name))
		out.Write("")

		if dep.err != nil {
			out.WriteLine(output.Linef("", output.StyleBold, "Encountered the following error:\n\n%s%s\n", output.StyleReset, dep.err))
		}

		out.WriteLine(output.Linef("", output.StyleBold, "How to fix:"))

		if dep.instructionsComment != "" {
			out.Write("")
			out.Write(dep.instructionsComment)
		}

		// If we don't have anything do run, we simply print instructions to
		// the user
		if dep.instructionsCommands == "" {
			writeFingerPointingLine("Hit return once you're done")
			waitForReturn()
			toFix = removeEntry(toFix, idx)
		} else {
			// Otherwise we print the command(s) and ask the user whether we should run it or not
			out.Write("")
			out.Write("Run the following command(s):")
			out.Write("")

			out.WriteLine(output.Line("", output.CombineStyles(output.StyleBold, output.StyleYellow), strings.TrimSpace(dep.instructionsCommands)))

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
				writeFingerPointingLine("Hit return once you're done")
				waitForReturn()
			case 2:
				if err := fixDependencyAutomatically(ctx, dep); err != nil {
					return err
				}
			case 3:
				return nil
			}
		}

		pending := out.Pending(output.Linef("", output.StylePending, "Determining status..."))
		for _, dep := range category.dependencies {
			dep.Update(ctx)
		}
		pending.Destroy()
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

func checkCommandOutputContains(cmd, contains string) func(context.Context) (bool, error) {
	return func(ctx context.Context) (bool, error) {
		elems := strings.Split(cmd, " ")
		out, _ := exec.Command(elems[0], elems[1:]...).CombinedOutput()
		return strings.Contains(string(out), contains), nil
	}
}

func checkFileContains(fileName, content string) func(context.Context) (bool, error) {
	return func(ctx context.Context) (bool, error) {
		file, err := os.Open(fileName)
		if err != nil {
			return false, errors.Wrapf(err, "failed to check that %q contains %q", fileName, content)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, content) {
				return true, nil
			}
		}

		if err := scanner.Err(); err != nil {
			return false, err
		}

		return false, errors.Newf("file %q did not contain %q", fileName, content)
	}
}

func checkInPath(cmd string) func(context.Context) (bool, error) {
	return func(ctx context.Context) (bool, error) {
		p, err := exec.LookPath(cmd)
		if err != nil {
			return false, errors.Newf("executable %q not found in $PATH", cmd)
		}
		return p != "", nil
	}
}

func checkInMainRepoOrRepoInDirectory(ctx context.Context) (bool, error) {
	_, err := root.RepositoryRoot()
	if err != nil {
		ok, err := pathExists("sourcegraph")
		if !ok || err != nil {
			return false, errors.New("'sg setup' is not run in sourcegraph and repository is also not found in current directory")
		}
		return true, nil
	}
	return true, nil
}

func checkDevPrivateInParentOrInCurrentDirectory(context.Context) (bool, error) {
	ok, err := pathExists("dev-private")
	if ok && err == nil {
		return true, nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return false, errors.Wrap(err, "failed to check for dev-private repository")
	}

	p := filepath.Join(wd, "..", "dev-private")
	ok, err = pathExists(p)
	if ok && err == nil {
		return true, nil
	}
	return false, errors.New("could not find dev-private repository either in current directory or one above")
}

func checkPostgresConnection(ctx context.Context) (bool, error) {
	// This check runs only in the `sourcegraph/sourcegraph` repository, so
	// we try to parse the globalConf and use its `Env` to configure the
	// Postgres connection.
	ok, _ := parseConf(*configFlag, *overwriteConfigFlag)
	if !ok {
		return false, errors.New("failed to read sg.config.yaml. This step of `sg setup` needs to be run in the `sourcegraph` repository")
	}

	getEnv := func(key string) string {
		// First look into process env, emulating the logic in makeEnv used
		// in internal/run/run.go
		val, ok := os.LookupEnv(key)
		if ok {
			return val
		}
		// Otherwise check in globalConf.Env
		val, _ = globalConf.Env[key]
		return val
	}

	dns := postgresdsn.New("", "", getEnv)
	conn, err := pgx.Connect(ctx, dns)
	if err != nil {
		return false, err
	}
	defer conn.Close(ctx)

	var result int
	row := conn.QueryRow(ctx, "SELECT 1;")
	if err := row.Scan(&result); err != nil {
		return false, err
	}
	if result != 1 {
		return false, errors.New("failed to read a test value from database")
	}
	return true, nil
}

func checkRedisConnection(context.Context) (bool, error) {
	conn, err := redis.Dial("tcp", ":6379", redis.DialConnectTimeout(5*time.Second))
	if err != nil {
		return false, errors.Wrap(err, "failed to connect to Redis at 127.0.0.1:6379")
	}

	if _, err := conn.Do("SET", "sg-setup", "was-here"); err != nil {
		return false, err
	}

	retval, err := redis.String(conn.Do("GET", "sg-setup"))
	if err != nil {
		return false, err
	}

	return retval == "was-here", nil
}

// TODO: We should change the signature to be `func(context.Context) error`
// and use the convention "success = err == nil" and use the errors to
// provide helpful messages as to why a check didn't pass.
type dependencyCheck func(context.Context) (bool, error)

type dependency struct {
	name string

	check dependencyCheck

	// TODO: Still unused
	onlyEmployees bool

	state bool
	err   error

	instructionsComment    string
	instructionsCommands   string
	requiresSgSetupRestart bool
}

func (d *dependency) IsMet() bool { return d.err == nil && d.state }

func (d *dependency) Update(ctx context.Context) {
	d.err = nil
	d.state = false

	ok, err := d.check(ctx)
	if err != nil {
		d.err = err
		d.state = false
		return
	}

	d.state = ok
}

type dependencyCategory struct {
	name         string
	dependencies []*dependency

	autoFixing         bool
	requiresRepository bool
}

func (cat *dependencyCategory) CombinedState() bool {
	for _, dep := range cat.dependencies {
		if dep.err != nil {
			return false
		} else if !dep.state {
			return false
		}
	}
	return true
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
	out.Write("")
	writeFingerPointingLine("What do you want to do?")

	for i := 0; i < len(choices); i++ {
		num := i + 1
		desc, ok := choices[num]
		if !ok {
			return 0, errors.Newf("internal error: %d not found in provided choices", i)
		}
		out.Writef("%s[%d]%s: %s", output.StyleBold, num, output.StyleReset, desc)
	}

	fmt.Printf("Enter choice: ")

	var s int
	_, err := fmt.Scan(&s)
	if err != nil {
		return 0, err
	}

	return s, nil
}

func retryCheck(check dependencyCheck, retries int, sleep time.Duration) dependencyCheck {
	return func(ctx context.Context) (ok bool, err error) {
		for i := 0; i < retries; i++ {
			ok, err = check(ctx)
			if ok {
				return true, nil
			}
			time.Sleep(sleep)
		}
		return ok, err
	}
}

func wrapCheckErr(check dependencyCheck, message string) dependencyCheck {
	return func(ctx context.Context) (bool, error) {
		ok, err := check(ctx)
		if err != nil {
			return ok, errors.Wrap(err, message)
		}
		return ok, err
	}
}

func checkCaddyTrusted(ctx context.Context) (bool, error) {
	certPath, err := caddySourcegraphCertificatePath()
	if err != nil {
		return false, errors.Wrap(err, "failed to determine path where proxy stores certificates")
	}

	ok, err := pathExists(certPath)
	if !ok || err != nil {
		return false, errors.New("sourcegraph.test certificate not found. highly likely it's not trusted by system")
	}

	rawCert, err := os.ReadFile(certPath)
	if err != nil {
		return false, errors.Wrap(err, "could not read certificate")
	}

	cert, err := pemDecodeSingleCert(rawCert)
	if err != nil {
		return false, errors.Wrap(err, "decoding cert failed")
	}

	if trusted(cert) {
		return true, nil
	}
	return false, errors.New("doesn't look like certificate is trusted")
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
