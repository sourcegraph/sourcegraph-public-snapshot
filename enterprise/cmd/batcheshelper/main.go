package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/batcheshelper/log"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/batcheshelper/run"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/batcheshelper/util"
	"github.com/sourcegraph/sourcegraph/internal/sanitycheck"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/execution"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func main() {
	sanitycheck.Pass()
	if err := doMain(); err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

func doMain() error {
	inputPath := flag.String("input", "input.json", "The input JSON file for the workspace execution. Defaults to \"input.json\".")
	previousPath := flag.String("previousStepPath", "", "The path to the previous step's result file. Defaults to current working directory.")
	workspaceFilesPath := flag.String("workspaceFiles", "/job/workspace-files", "The path to the workspace files. Defaults to \"/job/workspace-files\".")
	flag.Usage = usage

	// So golang flags get confused when arguments are mixed in. We need to do a little work to support `args -flags`.
	var flags []string
	var programArgs []string

	argLen := len(os.Args[1:])
	for i := 0; i < argLen; i++ {
		token := os.Args[i+1]
		if token[0] == '-' {
			flags = append(flags, token, os.Args[i+2])
			i++
		} else {
			programArgs = append(programArgs, token)
		}
	}
	if err := flag.CommandLine.Parse(flags); err != nil {
		return err
	}

	arguments, err := parseArgs(programArgs)
	if err != nil {
		return err
	}

	executionInput, err := parseInput(*inputPath)
	if err != nil {
		return err
	}

	previousResult, err := parsePreviousStepResult(*previousPath, arguments.step)
	if err != nil {
		return err
	}

	logger := &log.Logger{Writer: os.Stdout}

	wd, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "getting working directory")
	}

	ctx := context.Background()
	switch arguments.mode {
	case "pre":
		return run.Pre(ctx, logger, arguments.step, executionInput, previousResult, wd, *workspaceFilesPath)
	case "post":
		return run.Post(ctx, logger, &util.RealCmdRunner{}, arguments.step, executionInput, previousResult, wd, *workspaceFilesPath)
	default:
		return errors.Newf("invalid mode %q", arguments.mode)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s <pre|post> <step index> [OPTIONS]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "OPTIONS:\n")
	flag.PrintDefaults()
}

func parseArgs(arguments []string) (args, error) {
	if len(arguments) < 2 {
		return args{}, errors.New("missing arguments")
	} else if len(arguments) > 2 {
		return args{}, errors.New("too many arguments")
	}

	mode := arguments[0]
	if mode != "pre" && mode != "post" {
		return args{}, errors.Newf("invalid mode %q", mode)
	}

	step, err := strconv.Atoi(arguments[1])
	if err != nil {
		return args{}, errors.Wrap(err, "failed to parse step")
	}

	return args{mode, step}, nil
}

type args struct {
	mode string
	step int
}

func parseInput(inputPath string) (batcheslib.WorkspacesExecutionInput, error) {
	var executionInput batcheslib.WorkspacesExecutionInput

	input, err := os.ReadFile(inputPath)
	if err != nil {
		return executionInput, errors.Wrapf(err, "failed to read execution input file %q", inputPath)
	}

	if err = json.Unmarshal(input, &executionInput); err != nil {
		return executionInput, errors.Wrap(err, "failed to unmarshal execution input")
	}
	return executionInput, nil
}

func parsePreviousStepResult(path string, step int) (execution.AfterStepResult, error) {
	if step > 0 {
		// Read the previous step's result file.
		return getPreviouslyExecutedStep(path, step-1)
	}
	return execution.AfterStepResult{}, nil
}

func getPreviouslyExecutedStep(path string, previousStep int) (execution.AfterStepResult, error) {
	for i := previousStep; i >= 0; i-- {
		var previousResult execution.AfterStepResult
		stepResultPath := filepath.Join(path, util.StepJSONFile(i))
		stepJSON, err := os.ReadFile(stepResultPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return previousResult, errors.Wrap(err, "failed to read step result file")
		}
		if err = json.Unmarshal(stepJSON, &previousResult); err != nil {
			return previousResult, errors.Wrap(err, "failed to unmarshal step result file")
		}
		if !previousResult.Skipped {
			return previousResult, nil
		}
	}
	return execution.AfterStepResult{}, nil
}
