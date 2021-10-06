package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var (
	ciFlagSet = flag.NewFlagSet("sg ci", flag.ExitOnError)
	ciCommand = &ffcli.Command{
		Name:       "ci",
		ShortUsage: "sg ci preview",
		ShortHelp:  "Preview CI steps for the current branch.",
		LongHelp:   "Preview CI steps",
		FlagSet:    ciFlagSet,
		Exec:       ciExec,
	}
)

func ciExec(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return flag.ErrHelp
	}
	switch args[0] {
	case "preview":
		stdout.Out.WriteLine(output.Linef("", output.StyleSuggestion,
			"If the current branch were to be pushed, the following pipeline would be run:"))

		branch, err := run.GitCmd("branch", "--show-current")
		if err != nil {
			return err
		}
		message, err := run.GitCmd("show", "--format=%s\\n%b")
		if err != nil {
			return err
		}
		cmd := exec.Command("go", "run", "./enterprise/dev/ci/gen-pipeline.go", "-preview")
		cmd.Env = append(os.Environ(),
			fmt.Sprintf("BUILDKITE_BRANCH=%s", strings.TrimSpace(branch)),
			fmt.Sprintf("BUILDKITE_MESSAGE=%s", strings.TrimSpace(message)))
		out, err := run.InRoot(cmd)
		if err != nil {
			return err
		}
		stdout.Out.Write(out)
		return nil
	default:
		return flag.ErrHelp
	}
}
