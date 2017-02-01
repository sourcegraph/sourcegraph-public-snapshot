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
		// Remove old bundle.
		rmCmd := exec.Command("git", "clean", "-fdx", "ui/assets")
		if err := execCmd(rmCmd); err != nil {
			return err
		}

		webpackCmd := exec.Command("npm", "run", "build")
		webpackCmd.Dir = "./ui"
		if err := execCmd(webpackCmd); err != nil {
			return err
		}
	}

	genCmd := exec.Command("go", "generate", "./app/assets", "./app/templates")
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
				"GOPATH=" + os.Getenv("GOPATH"),
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
	return []string{
		fmt.Sprintf("-X %q", fmt.Sprintf("sourcegraph.com/sourcegraph/sourcegraph/cli/buildvar.Version=%s", c.Args.Version)),
	}, nil
}
