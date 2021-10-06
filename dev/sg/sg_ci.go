package main

import (
	"context"
	"flag"
	"os"
	"os/exec"

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
			"If the current branch were to be pushed, the CI would run as following:"))
		cmd := exec.Command("go", "run", "./enterprise/dev/ci/gen-pipeline.go", "-preview")
		cmd.Env = os.Environ()
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
