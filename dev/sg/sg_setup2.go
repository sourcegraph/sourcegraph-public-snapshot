package main

import (
	"context"
	"flag"
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

	out.WriteLine(output.Linef("", output.StyleLinesAdded, "Welcome to 'sg setup'!"))

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

	pending := out.Pending(output.Linef("", output.StylePending, "Checking system..."))
	for _, category := range deps {
		for _, dep := range category.dependencies {
			time.Sleep(100 * time.Millisecond)
			dep.Update(ctx)
		}

	}

	pending.Destroy()

	for i, category := range deps {
		idx := i + 1
		if combined := category.CombinedState(); combined {
			out.WriteLine(output.Linef(output.EmojiSuccess, output.StyleSuccess, "%d %s", idx, category.name))
		} else {
			out.WriteLine(output.Linef(output.EmojiFailure, output.StyleWarning, "%d %s", idx, category.name))
			for _, dep := range category.dependencies {
				if dep.err != nil {
					out.WriteLine(output.Linef("  "+output.EmojiFailure, output.StyleWarning, "%s: %s", dep.name, dep.err))
				} else if !dep.state {
					out.WriteLine(output.Linef("  "+output.EmojiFailure, output.StyleWarning, "%s: %s", dep.name, "check failed"))
				} else {
					out.WriteLine(output.Linef("  "+output.EmojiSuccess, output.StyleSuccess, "%s", dep.name))
				}
			}
		}
	}

	return nil
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

func checkFileContains(file, content string) func(context.Context) (bool, error) {
	return func(ctx context.Context) (bool, error) {
		return false, errors.New("todo: not implemented")
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
		conn, err := redis.Dial("tcp", "localhost:6379")
		if err != nil {
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

	onlyEmployees bool

	state bool
	err   error
}

func (d *dependency) Update(ctx context.Context) {
	ok, err := d.check(ctx)
	if err != nil {
		d.err = err
		return
	}

	d.state = ok
}

type dependencyCategory struct {
	name         string
	dependencies []*dependency
}

func (cat *dependencyCategory) CombinedState() (state bool) {
	for _, dep := range cat.dependencies {
		if dep.err != nil {
			state = false
		} else {
			state = dep.state
		}
	}
	return state
}

var macOSDependencies = []dependencyCategory{
	{
		name: "Install homebrew",
		dependencies: []*dependency{
			{name: "brew", check: checkInPath("brew")},
		},
	},
	{
		name: "Install base utilities (git, docker, ...)",
		dependencies: []*dependency{
			{name: "git", check: checkInPath("git")},
			{name: "docker", check: checkInPath("docker")},
			{name: "gnu-sed", check: checkInPath("gsed")},
			{name: "comby", check: checkInPath("comby")},
			{name: "pcre", check: checkInPath("pcregrep")},
			{name: "sqlite", check: checkInPath("sqlite3")},
			{name: "jq", check: checkInPath("jq")},
		},
	},
	{
		name: "Clone repositories",
		dependencies: []*dependency{
			{name: "github.com/sourcegraph/sourcegraph", check: checkInMainRepoOrRepoInDirectory()},
			{name: "github.com/sourcegraph/dev-private", check: checkDevPrivateInParentOrInCurrentDirectory(), onlyEmployees: true},
		},
	},
	{
		name: "Programming languages & tooling",
		dependencies: []*dependency{
			{name: "go", check: checkInPath("go")},
			{name: "yarn", check: checkInPath("yarn")},
			{name: "node", check: checkInPath("node")},
		},
	},
	{
		name: "Setup PostgreSQL database",
		dependencies: []*dependency{
			{name: "Connection to 'sourcegraph' database", check: checkPostgresConnection()},
			{name: "psql", check: checkInPath("psql")},
		},
	},
	{
		name: "Setup Redis database",
		dependencies: []*dependency{
			{name: "Connection to Redis", check: checkRedisConnection()},
		},
	},
	{
		name: "Setup proxy for local development",
		dependencies: []*dependency{
			{name: "/etc/hosts contains sourcegraph.test", check: checkFileContains("/etc/hosts", "sourcegraph.test")},
			{name: "is there a way to check whether root certificate is trusted?", check: checkFileContains("/etc/hosts", "sourcegraph.test")},
		},
	},
}
