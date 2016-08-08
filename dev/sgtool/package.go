package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/neelance/parallel"
)

func init() {
	c, err := CLI.AddCommand("package",
		"build a binary package for release locally",
		"The package command builds and packages a new Sourcegraph binary version by (1) generating frontend assets (with webpack), (2) generating Go template and asset vfsdata, and (3) cross-compiling binaries for all target platforms.",
		&packageCmd,
	)
	if err != nil {
		log.Fatal(err)
	}
	c.Aliases = []string{"pkg"}
}

type PackageCmd struct {
	OS string `long:"os" description:"operating system targets to cross-compile for" default-mask:"$GOOS"`

	SkipWebpack bool `long:"skip-webpack" description:"skip webpack (JS/SCSS preparation, concatenation, minification, etc.)"`
	IgnoreDirty bool `long:"ignore-dirty" description:"ignore dirty working directory"`

	Args struct {
		Version string `name:"version" description:"version number ('1.2.3') or identifier ('snapshot' is default)"`
	} `positional-args:"yes"`
}

var packageCmd PackageCmd

func (c *PackageCmd) Execute(args []string) error {
	// Check for dependencies before starting.
	if err := requireCmds("npm", "go"); err != nil {
		return err
	}

	if c.OS == "" {
		c.OS = runtime.GOOS
	}

	if c.Args.Version == "" {
		c.Args.Version = "snapshot"
	}

	// Check for "rego" that rebuilds sourcegraph.
	if found, err := pgrep("rego"); found {
		return errors.New("Error: rego process is running, likely to rebuild Sourcegraph binary when source files change (e.g., via `make serve-dev`). This conflicts with the binary packaging process and can result in broken binaries. Kill the rego process and try again.")
	} else if err != nil {
		return err
	}

	if !c.SkipWebpack {
		webpackCmd := exec.Command("npm", "run", "build")
		webpackCmd.Dir = "./ui"
		if err := execCmd(webpackCmd); err != nil {
			return err
		}
	}

	gopath := strings.Join([]string{
		filepath.Join(os.Getenv("PWD"), "Godeps", "_workspace"),
		os.Getenv("GOPATH"),
	}, string(filepath.ListSeparator))

	genCmd := exec.Command("go", "generate", "./app/assets", "./app/templates")
	overrideEnv(genCmd, "GOPATH", gopath)
	if err := execCmd(genCmd); err != nil {
		return err
	}

	ldflagsInput, err := c.getLDFlags()
	if err != nil {
		return err
	}

	bins := []string{}
	par := parallel.NewRun(2)
	for _, osName := range strings.Split(c.OS, " ") {
		osName := osName
		dest, err := filepath.Abs(filepath.Join("release", c.Args.Version, osName+"-amd64"))
		if err != nil {
			return err
		}
		bins = append(bins, dest)
		par.Acquire()
		go func() {
			defer par.Release()
			ldflags := make([]string, len(ldflagsInput))
			copy(ldflags, ldflagsInput)

			env := []string{
				"GOOS=" + osName,
				"GOARCH=amd64",
				"GOPATH=" + gopath,
				"PATH=" + os.Getenv("PATH"),
				"CGO_ENABLED=0",
			}

			cmd := exec.Command("go", "build", "-a", "-installsuffix", "netgo", "-ldflags", strings.Join(ldflags, " "), "-tags", "dist netgo", "-o", dest, ".")
			cmd.Dir = filepath.Join("cmd", "src")
			cmd.Env = env
			if err := execCmd(cmd); err != nil {
				par.Error(err)
				return
			}
		}()
	}
	if err := par.Wait(); err != nil {
		return err
	}

	log.Printf("Binaries built at:")
	for _, bin := range bins {
		log.Printf(" - %s", bin)
	}

	return nil
}

func (c *PackageCmd) getLDFlags() ([]string, error) {
	buildvars := map[string]string{
		"Version": c.Args.Version,
		"dateStr": time.Now().Format(time.UnixDate),
	}

	commitID, err := cmdOutput("git", "rev-parse", "--verify", "HEAD")
	if err != nil {
		return nil, err
	}
	if commitID != "" {
		buildvars["commitID"] = strings.TrimSpace(commitID)
	}

	status, err := cmdOutput("git", "status", "--porcelain")
	if err != nil {
		return nil, err
	}
	if status != "" {
		if !c.IgnoreDirty {
			diff, err := cmdOutput("git", "diff")
			if err != nil {
				return nil, err
			}
			if mb := 1024 * 1024; len(diff) > mb {
				diff = diff[:mb] // Capped to 1MB in size at max
			}
			return nil, fmt.Errorf(`
Aborting! Working directory is dirty, binary would be compromised!

note: You can use --ignore-dirty to skip this check.
note: git status --porcelain reported:

%s

note: git diff reports:

%s
`, status, diff)
		}
		buildvars["dirtyStr"] = "true"
	}

	uname, err := cmdOutput("uname", "-a")
	if err != nil {
		return nil, err
	}
	if uname != "" {
		buildvars["host"] = strings.TrimSpace(uname)
	}

	if user := os.Getenv("USER"); user != "" {
		buildvars["user"] = user
	}

	var ldflags []string
	for name, val := range buildvars {
		ldflags = append(ldflags, fmt.Sprintf("-X %q", fmt.Sprintf("sourcegraph.com/sourcegraph/sourcegraph/cli/buildvar.%s=%s", name, val)))
	}

	// main ldflags
	mainLDFlags := map[string]string{
		"VERSION":    c.Args.Version,
		"BUILD_DATE": time.Now().Format(time.RFC3339),
	}
	for name, val := range mainLDFlags {
		ldflags = append(ldflags, fmt.Sprintf("-X %q", fmt.Sprintf("main.%s=%s", name, val)))
	}

	return ldflags, nil
}
