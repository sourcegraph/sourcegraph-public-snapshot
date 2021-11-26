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
		panic("unsupported os!")
	}

	// TODO: Check whether we are in the repository or not

	failed := []int{}
	all := []int{}
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
			out.Write("")
			out.WriteLine(output.Linef(output.EmojiOk, output.StyleBold, "Everything looks good! Happy hacking!"))
			return nil
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

func presentFailedCategoryWithOptions(ctx context.Context, categoryIdx int, category *dependencyCategory) error {
	printCategoryHeaderAndDependencies(categoryIdx, category)

	// TODO: It doesn't make a lot of sense to give a choice here if
	// there's only one dependency

	choices := map[int]string{1: "I want to fix these one-by-one"}
	if category.enableAutoFixing {
		choices[2] = "I'm feeling lucky. You try fixing all of it for me."
		choices[3] = "Go back"
	} else {
		choices[2] = "Go back"
	}

	choice, err := getChoice(choices)
	if err != nil {
		return err
	}

	out.ClearScreen()

	switch choice {
	case 1:
		err = fixCategoryManually(ctx, category)
	case 2:
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

	for _, dep := range category.dependencies {
		if dep.err == nil || dep.state {
			writeSuccessLine("%s", dep.name)
		} else if dep.err != nil {
			writeFailureLine("%s: %s", dep.name, dep.err)
		} else if !dep.state {
			writeFailureLine("%s: %s", dep.name, "check failed")
		}
	}
}

func fixCategoryAutomatically(ctx context.Context, category *dependencyCategory) error {
	for _, dep := range category.dependencies {
		if dep.err == nil || dep.state {
			continue
		}

		pending := out.Pending(output.Linef("", output.StylePending, "Trying my hardest to fix %q automatically...", dep.name))
		c := exec.CommandContext(ctx, "bash", "-c", dep.instructionsCommands)
		cmdOut, err := c.CombinedOutput()
		if err != nil {
			pending.Complete(output.Linef(output.EmojiFailure, output.StyleWarning, "Error: %s\n\nOutput: %s", err, cmdOut))
			return err
		}
		pending.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Done! %q should work now", dep.name))
	}

	return nil
}

func fixCategoryManually(ctx context.Context, category *dependencyCategory) error {
	// TODO: ask for confirmation
	for _, dep := range category.dependencies {
		if dep.err == nil || dep.state {
			continue
		}

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
			out.WriteLine(output.Linef(output.EmojiFingerPointRight, output.StyleBold, "Hit return once you're done"))
			waitForReturn()
			return nil
		}

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

		out.ClearScreen()
		switch choice {
		case 1:
			out.WriteLine(output.Linef(output.EmojiFingerPointRight, output.StyleBold, "Hit return once you're done"))
			waitForReturn()
		case 2:
			pending := out.Pending(output.Linef("", output.StylePending, "Trying my hardest to fix %q...", dep.name))

			cmdOut, err := exec.CommandContext(ctx, "bash", "-c", dep.instructionsCommands).CombinedOutput()
			if err != nil {
				pending.WriteLine(output.Linef(output.EmojiFailure, output.StyleWarning, "failed to run command: %s\n\noutput: %s", err, cmdOut))
				return err
			}

			pending.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Done! %q should be fixed now!", dep.name))
		case 3:
			return nil
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
	connectToRedis := func(ctx context.Context) (bool, error) {
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

	return func(ctx context.Context) (ok bool, err error) {
		for i := 0; i < 5; i++ {
			ok, err = connectToRedis(ctx)
			if ok {
				return true, nil
			}
			time.Sleep(500 * time.Millisecond)
		}
		return ok, err
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
	enableAutoFixing bool
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
			{
				name:  "brew",
				check: checkInPath("brew"),
				// instructionsCommands: `/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`,
				instructionsComment: `We depend on having the Homebrew package manager available on macOS.

Follow the instructions at https://brew.sh to install it, then rerun 'sg setup'.`,
			},
		},
		enableAutoFixing: false,
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
			{name: "bash", check: checkCommandOutputContains("bash --version", "version 5"), instructionsCommands: `brew install bash`},
		},
		enableAutoFixing: true,
	},
	{
		name: "Clone repositories",
		dependencies: []*dependency{
			{
				name:                 "github.com/sourcegraph/sourcegraph",
				check:                checkInMainRepoOrRepoInDirectory(),
				instructionsCommands: `git clone git@github.com:sourcegraph/sourcegraph.git`,
				instructionsComment: `` +
					`The 'sourcegraph' repository contains the Sourcegraph codebase and everything to run Sourcegraph locally.`,
			},
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
		// TODO: customAllInOnecommand
		// - install asdf
		// - reload asdf
		// - check for sourcegraph repository
		// - go into sourcegraph repository
		// - run the other commands

	},
	{
		name: "Setup PostgreSQL database",
		dependencies: []*dependency{
			{
				name:  "Connection to 'sourcegraph' database",
				check: checkPostgresConnection(),
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
	out.WriteLine(output.Linef(output.EmojiFingerPointRight, output.StyleBold, "What do you want to do?"))

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
