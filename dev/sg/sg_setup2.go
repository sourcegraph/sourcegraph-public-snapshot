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
	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/output"
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
					out.WriteLine(output.Linef("\t"+output.EmojiFailure, output.StyleWarning, "%s: %s", dep.name, dep.err))
				} else if !dep.state {
					out.WriteLine(output.Linef("\t"+output.EmojiFailure, output.StyleWarning, "%s: %s", dep.name, "check failed"))
				} else {
					out.WriteLine(output.Linef("\t"+output.EmojiSuccess, output.StyleWarning, "%s", dep.name))
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

func checkInPath(cmd string) func(context.Context) (bool, error) {
	return func(ctx context.Context) (bool, error) {
		p, err := os.Executable()
		if err != nil {
			return false, err
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
}
