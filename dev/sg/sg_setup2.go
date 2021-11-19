package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

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

	var deps []dependencyCategory
	if currentOS == "darwin" {
		deps = macOSDependencies
	} else {
		panic("unsupported os!")
	}

	failed := []int{}
	for i := range deps {
		failed = append(failed, i)
	}

	for len(failed) != 0 {

		out.WriteLine(output.Linef("", output.CombineStyles(output.StyleBold, output.StyleOrange), "-------------------------------------"))
		out.WriteLine(output.Linef("", output.CombineStyles(output.StyleBold, output.StyleOrange), "|        Welcome to sg setup!       |"))
		out.WriteLine(output.Linef("", output.CombineStyles(output.StyleBold, output.StyleOrange), "-------------------------------------"))

		for i, category := range deps {
			idx := i + 1

			for _, dep := range category.dependencies {
				dep.Update(ctx)
			}

			if combined := category.CombinedState(); combined {
				out.WriteLine(output.Linef(output.EmojiSuccess, output.StyleSuccess, "%d. %s", idx, category.name))
				failed = removeEntry(failed, i)
			} else {
				out.WriteLine(output.Linef(output.EmojiFailure, output.StyleWarning, "%d. %s", idx, category.name))
			}
		}

		if len(failed) == 0 {
			out.Write("")
			out.WriteLine(output.Linef(output.EmojiOk, output.StyleBold, "Everything looks good! Happy hacking!"))
			return nil
		}

		out.Write("")
		out.WriteLine(output.Linef(output.EmojiWarningSign, output.StyleYellow, "Some checks failed. Which one do you want to fix?"))

		toFix := getNumber(1, len(deps))

		// TODO: Check bounds
		category := deps[toFix-1]

		out.Write("")
		out.WriteLine(output.Linef(output.EmojiLightbulb, output.CombineStyles(output.StyleSearchQuery, output.StyleBold), "%d. %s", toFix, category.name))
		out.Write("")
		out.Write("Dependencies:")

		// TODO: This is duplicate from above
		for _, dep := range category.dependencies {
			if dep.err == nil || dep.state {
				out.WriteLine(output.Linef(output.EmojiSuccess, output.StyleSuccess, "%s", dep.name))
			} else if dep.err != nil {
				out.WriteLine(output.Linef(output.EmojiFailure, output.StyleWarning, "%s: %s", dep.name, dep.err))
			} else if !dep.state {
				out.WriteLine(output.Linef(output.EmojiFailure, output.StyleWarning, "%s: %s", dep.name, "check failed"))
			}
		}

		// TODO: We need to refactor `getChoice` to allow passing in "choices" with custom text
		// and here we have to ask the user:
		//
		// [m]: I want to fix the steps one-by-one
		// [l]: I'm feeling lucky. Try fixing all of it for me.
		//
		out.WriteLine(output.Linef(output.EmojiFingerPointRight, output.StyleBold, "What do you want to do?"))
		choice, err := getChoice()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		switch choice {
		case userChoiceManually:
			// TODO: ask for confirmation
			out.Write("")
			out.Write("Let's work through the failures...")
			out.Write("")

			for _, dep := range category.dependencies {
				if dep.err == nil || dep.state {
					continue
				}

				out.WriteLine(output.Linef(output.EmojiFailure, output.CombineStyles(output.StyleWarning, output.StyleBold), "---------------------------------------"))
				out.WriteLine(output.Linef(output.EmojiFailure, output.CombineStyles(output.StyleWarning, output.StyleBold), "|               %s", dep.name))
				out.WriteLine(output.Linef(output.EmojiFailure, output.CombineStyles(output.StyleWarning, output.StyleBold), "---------------------------------------"))
				if dep.err != nil {
					out.WriteLine(output.Linef("", output.StyleBold, "Error: %s%s", output.StyleReset, dep.err))
				}
				if dep.instructionsComment != "" {
					out.WriteLine(output.Linef("", output.StyleBold, "How to fix:"))
					out.Write("")
					out.Write(dep.instructionsComment)
				}

				if dep.instructionsCommands != "" {
					out.Write("")
					out.Write("Run the following command(s):")
					out.Write("")

					out.WriteLine(output.Line("", output.CombineStyles(output.StyleBold, output.StyleYellow), strings.TrimSpace(dep.instructionsCommands)))

					out.Write("")
				}

				out.WriteLine(output.Linef(output.EmojiFingerPointRight, output.StyleBold, "What do you want to do?"))
				choice, err := getChoice()
				if err != nil {
					if err == io.EOF {
						return nil
					}
					return err
				}

				switch choice {
				case userChoiceManually:
					out.WriteLine(output.Linef(output.EmojiFingerPointRight, output.StyleBold, "Hit return once you're done"))
					waitForReturn()
				case userChoiceAutomatic:
					if dep.instructionsCommands == "" {
						out.WriteLine(output.Linef(output.EmojiFailure, output.StyleWarning, "problem! not possible. exiting"))
						return nil
					}

					pending := out.Pending(output.Line("", output.StylePending, "Running command..."))
					c := exec.CommandContext(ctx, "bash", "-c", dep.instructionsCommands)
					cmdOut, err := c.CombinedOutput()
					if err != nil {

						pending.WriteLine(output.Linef(output.EmojiFailure, output.StyleWarning, "failed to run command: %s\n\noutput: %s", err, cmdOut))
						return err
					}
					pending.Complete(output.Line(output.EmojiSuccess, output.StyleSuccess, "Done!"))
				}
			}

		case userChoiceAutomatic:
			for _, dep := range category.dependencies {
				if dep.err == nil || dep.state {
					continue
				}

				pending := out.Pending(output.Line("", output.StylePending, "Running command..."))
				c := exec.CommandContext(ctx, "bash", "-c", dep.instructionsCommands)
				cmdOut, err := c.CombinedOutput()
				if err != nil {

					pending.WriteLine(output.Linef(output.EmojiFailure, output.StyleWarning, "failed to run command: %s\n\noutput: %s", err, cmdOut))
					return err
				}
				pending.Complete(output.Line(output.EmojiSuccess, output.StyleSuccess, "Done!"))
			}
		}
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
		out, err := exec.Command(elems[0], elems[1:]...).CombinedOutput()
		if err != nil {
			return false, err
		}
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

func checkInMainRepoOrRepoInDirectory() func(context.Context) (bool, error) {
	return func(ctx context.Context) (bool, error) {
		_, err := root.RepositoryRoot()
		if err != nil {
			return pathExists("sourcegraph")
		}
		return true, nil
	}
}

func checkCaddyTrusted() func(context.Context) (bool, error) {
	return func(ctx context.Context) (bool, error) {
		var path string
		repoRoot, err := root.RepositoryRoot()
		if err != nil {
			ok, err := pathExists("sourcegraph")
			if err != nil {
				return false, errors.Wrap(err, "failed to check whether we're in the sourcegraph repository or not")
			}
			if !ok {
				return false, errors.New("cannot find sourcegraph repository. rerun this command inside the repository")
			}

			wd, err := os.Getwd()
			if err != nil {
				return false, errors.Wrap(err, "failed to get current working directory")
			}
			path = filepath.Join(wd, "sourcegraph")
		} else {
			path = repoRoot
		}

		// TODO: This will download caddy the first time it's run
		cmd := exec.CommandContext(ctx, "dev/caddy.sh", "trust")
		cmd.Dir = path

		out, err := cmd.CombinedOutput()
		if err != nil {
			return false, errors.Wrap(err, "running 'dev/caddy.sh trust' failed")
		}

		return strings.Contains(string(out), "is already trusted by system"), nil
	}
}

func checkDevPrivateInParentOrInCurrentDirectory() func(context.Context) (bool, error) {
	return func(ctx context.Context) (bool, error) {
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
}

func checkPostgresConnection() func(context.Context) (bool, error) {
	return func(ctx context.Context) (bool, error) {
		// TODO: Do we need to use the globalconf from `sg` so
		// that we use the correct `PG*` env vars? But what if
		// the user doesn't have the repo cloned or is not in
		// the repo yet?
		ok, _ := parseConf(*configFlag, *overwriteConfigFlag)
		if !ok {
			return false, errors.New("failed to read sg.config.yaml. This step of `sg setup` needs to be run in the `sourcegraph` repository")
		}

		dns := postgresdsn.New("", "", os.Getenv)
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
}

func checkRedisConnection() func(context.Context) (bool, error) {
	return func(ctx context.Context) (bool, error) {
		out.Writef("redis dial")
		conn, err := redis.Dial("tcp", "localhost:6379")
		if err != nil {
			out.Writef("err=%s", err)
			return false, errors.Wrap(err, "failed to connect to Redis at localhost:6379")
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
}

type dependency struct {
	name  string
	check func(context.Context) (bool, error)

	// TODO: Still unused
	onlyEmployees bool

	state bool
	err   error

	instructionsComment  string
	instructionsCommands string
	automaticCommands    string
}

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

	// TODO: Rename this field
	enableAllInOneCommand bool
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

var macOSDependencies = []dependencyCategory{
	{
		name: "Install homebrew",
		dependencies: []*dependency{
			{name: "brew", check: checkInPath("brew")},
		},
		// TODO: Do not enable all in one?
	},
	{
		name: "Install base utilities (git, docker, ...)",
		dependencies: []*dependency{
			{name: "git", check: checkInPath("git"), instructionsCommands: `brew install git`},
			{name: "docker", check: checkInPath("docker"), instructionsCommands: `brew install --cask docker`},
			{name: "gnu-sed", check: checkInPath("gsed"), instructionsCommands: "brew install gnu-sed"},
			{name: "comby", check: checkInPath("comby"), instructionsCommands: "brew install comby"},
			{name: "pcre", check: checkInPath("pcregrep"), instructionsCommands: `brew install pcre`},
			{name: "sqlite", check: checkInPath("sqlite3"), instructionsCommands: `brew install sqlite`},
			{name: "jq", check: checkInPath("jq"), instructionsCommands: `brew install jq`},
		},
		enableAllInOneCommand: true,
		// TODO: Enable all in one?
	},
	{
		name: "Clone repositories",
		// TODO: enableAllInOneCommand??
		dependencies: []*dependency{
			{name: "github.com/sourcegraph/sourcegraph", check: checkInMainRepoOrRepoInDirectory()},
			{
				name:                 "github.com/sourcegraph/dev-private",
				check:                checkDevPrivateInParentOrInCurrentDirectory(),
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
	},
	{
		name: "Programming languages & tooling",
		// TODO: enableAllInOneCommand??
		dependencies: []*dependency{
			// TODO: install asdf
			{name: "go", check: checkInPath("go")},
			{name: "yarn", check: checkInPath("yarn")},
			{name: "node", check: checkInPath("node")},
		},
		// todo: customAllInOnecommand
		// - install asdf
		// - reload asdf
		// - check for sourcegraph repository
		// - go into sourcegraph repository
		// - run the other commands

	},
	{
		name: "Setup PostgreSQL database",
		dependencies: []*dependency{
			// TODO: No instructions
			{name: "Connection to 'sourcegraph' database", check: checkPostgresConnection()},
			{name: "psql", check: checkInPath("psql")},
		},
	},
	{
		name: "Setup Redis database",
		dependencies: []*dependency{
			{
				name:  "Connection to Redis",
				check: checkRedisConnection(),
				instructionsComment: `` +
					`Sourcegraph requires the Redis database to be running. We recommend installing it with Homebrew and starting it as a system service.`,
				instructionsCommands: "brew reinstall redis && brew services start redis",
			},
		},
	},
	{
		name: "Setup proxy for local development",
		dependencies: []*dependency{
			// TODO: No instructions
			{name: "/etc/hosts contains sourcegraph.test", check: checkFileContains("/etc/hosts", "sourcegraph.test")},
			// TODO: No instructions
			{name: "is there a way to check whether root certificate is trusted?", check: checkCaddyTrusted()},
		},
	},
}

func getNumber(min, max int) int {
	var s int

	fmt.Printf("[%d-%d]: ", min, max)
	_, err := fmt.Scan(&s)
	if err != nil {
		panic(err)
	}

	return s
}

func waitForReturn() { fmt.Scanln() }

type userChoice string

const (
	userChoiceManually  = "manually"
	userChoiceAutomatic = "automatic"
)

func getChoice() (userChoice, error) {
	var s string

	choices := map[string]string{
		"m": "I'll run the command manually",
		"a": "You can run the command for me",
	}

	for letter, desc := range choices {
		out.Writef("%s[%s]%s: %s", output.StyleBold, letter, output.StyleReset, desc)
	}

	fmt.Printf("Enter choice: ")

	_, err := fmt.Scan(&s)
	if err != nil {
		return "", err
	}

	s = strings.TrimSpace(s)
	s = strings.ToLower(s)

	if s == "m" {
		return userChoiceManually, nil
	}
	return userChoiceAutomatic, nil
}
