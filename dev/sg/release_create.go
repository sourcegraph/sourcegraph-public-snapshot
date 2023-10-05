package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/sourcegraph/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

type createReleaseManifest struct {
	ProductName  string   `yaml:"productName"`
	Owners       []string `yaml:"owners"`
	Repository   string   `yaml:"repository"`
	Requirements []struct {
		Name            string `yaml:"name"`
		Cmd             string `yaml:"cmd"`
		Env             string `yaml:"env"`
		FixInstructions string `yaml:"fixInstructions"`
	} `yaml:"requirements"`
	Steps struct {
		Patch []cmdManifest `yaml:"patch"`
		Minor []cmdManifest `yaml:"minor"`
		Major []cmdManifest `yaml:"major"`
	} `yaml:"steps"`
}

type cmdManifest struct {
	Name string `yaml:"name"`
	Cmd  string `yaml:"cmd"`
}

func createReleaseCommand(cctx *cli.Context) error {
	pretend := cctx.Bool("pretend")
	version := cctx.String("version")
	vars := map[string]string{
		"version": version,
		"tag":     strings.TrimPrefix(version, "v"),
	}

	workdir := cctx.String("workdir")
	announce2("setup", "Finding release manifest in %q", workdir)
	if err := os.Chdir(cctx.String("workdir")); err != nil {
		return err
	}

	f, err := os.Open("release.yaml")
	if err != nil {
		say("setup", "failed to find release manifest")
		return err
	}
	defer f.Close()

	var m createReleaseManifest
	dec := yaml.NewDecoder(f)
	if err := dec.Decode(&m); err != nil {
		say("setup", "failed to decode manifest")
	}

	saySuccess("setup", "Found manifest for %q (%s)", m.ProductName, m.Repository)
	say("meta", "Owners: %s", strings.Join(m.Owners, ", "))
	say("meta", "Will create a patch release %q", version)

	announce2("reqs", "Checking requirements...")
	var failed bool
	for _, req := range m.Requirements {
		if req.Env != "" && req.Cmd != "" {
			return errors.Newf("requirement %q can't have both env and cmd defined", req.Name)
		}
		if req.Env != "" {
			if _, ok := os.LookupEnv(req.Env); !ok {
				failed = true
				sayFail("reqs", "FAIL %s, $%s is not defined.", req.Name, req.Env)
				continue
			}
			saySuccess("reqs", "OK %s", req.Name)
			continue
		}

		lines, err := run.Cmd(cctx.Context, req.Cmd).Run().Lines()
		if err != nil {
			failed = true
			sayFail("reqs", "FAIL %s", req.Name)
			sayFail("reqs", "  Error: %s", err.Error())
			for _, line := range lines {
				sayFail("reqs", "  "+line)
			}
		} else {
			saySuccess("reqs", "OK %s", req.Name)
		}
	}
	if failed {
		announce2("reqs", "Requirement checks failed, aborting.")
		return errors.New("failed requirements")
	}

	var steps []cmdManifest
	switch cctx.String("type") {
	case "patch":
		steps = m.Steps.Patch
	case "minor":
		steps = m.Steps.Minor
	case "major":
		steps = m.Steps.Major
	}

	for _, step := range steps {
		cmd := interpolate(step.Cmd, vars)
		if pretend {
			announce2("step", "Pretending to run step %q", step.Name)
			for _, line := range strings.Split(cmd, "\n") {
				say(step.Name, line)
			}
			continue
		}
		announce2("step", "Running step %q", step.Name)
		err := run.Bash(cctx.Context, cmd).Run().StreamLines(func(line string) {
			say(step.Name, line)
		})
		if err != nil {
			sayFail(step.Name, "Step failed: %v", err)
			return err
		} else {
			saySuccess("step", "Step %q succeeded", step.Name)
		}
	}
	return nil
}

func interpolate(s string, m map[string]string) string {
	for k, v := range m {
		s = strings.ReplaceAll(s, fmt.Sprintf("{{%s}}", k), v)
	}
	return s
}

func announce2(section string, format string, a ...any) {
	std.Out.WriteLine(output.Linef("👉", output.StyleBold, fmt.Sprintf("[%10s] %s", section, format), a...))
}

func say(section string, format string, a ...any) {
	sayKind(output.StyleReset, section, format, a...)
}

func sayWarn(section string, format string, a ...any) {
	sayKind(output.StyleOrange, section, format, a...)
}

func sayFail(section string, format string, a ...any) {
	sayKind(output.StyleRed, section, format, a...)
}

func saySuccess(section string, format string, a ...any) {
	sayKind(output.StyleGreen, section, format, a...)
}

func sayKind(style output.Style, section string, format string, a ...any) {
	std.Out.WriteLine(output.Linef("  ", style, fmt.Sprintf("[%10s] %s", section, format), a...))
}
