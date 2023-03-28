package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/batcheshelper/run"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/execution"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func main() {
	if err := doMain(); err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

func doMain() error {
	inputPath := flag.String("input", "input.json", "The path to the input file. Defaults to \"input.json\".")
	flag.Usage = usage
	flag.Parse()

	arguments, err := parseArgs(os.Args[1:])
	if err != nil {
		return err
	}

	executionInput, err := parseInput(*inputPath)
	if err != nil {
		return err
	}

	previousResult, err := parsePreviousStepResult(arguments.step)
	if err != nil {
		return err
	}

	ctx := context.Background()
	switch arguments.mode {
	case "pre":
		return run.Pre(ctx, arguments.step, executionInput, previousResult)
	case "post":
		return run.Post(ctx, arguments.step, executionInput, previousResult)
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
	if len(arguments) != 2 {
		return args{}, errors.New("missing arguments")
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

func parsePreviousStepResult(step int) (execution.AfterStepResult, error) {
	var previousResult execution.AfterStepResult
	if step > 0 {
		stepResultPath := fmt.Sprintf("step%d.json", step-1)
		stepJSON, err := os.ReadFile(stepResultPath)
		if err != nil {
			return previousResult, errors.Wrap(err, "failed to read step result file")
		}
		if err = json.Unmarshal(stepJSON, &previousResult); err != nil {
			return previousResult, errors.Wrap(err, "failed to unmarshal step result file")
		}
	}
	return previousResult, nil
}
