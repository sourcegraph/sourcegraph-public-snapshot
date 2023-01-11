package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

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
	ctx := context.Background()
	args := os.Args[1:]
	if len(args) != 2 {
		return errors.New("invalid argument count")
	}

	mode := args[0]

	stepIdx, err := strconv.Atoi(args[1])
	if err != nil {
		return errors.Wrap(err, "invalid step index")
	}

	var executionInput batcheslib.WorkspacesExecutionInput
	var previousResult execution.AfterStepResult

	c, err := os.ReadFile("input.json")
	if err != nil {
		return errors.Wrap(err, "failed to read execution input file")
	}

	if err := json.Unmarshal(c, &executionInput); err != nil {
		return errors.Wrap(err, "failed to unmarshal execution input")
	}

	if stepIdx > 0 {
		stepResultPath := fmt.Sprintf("step%d.json", stepIdx-1)
		c, err := os.ReadFile(stepResultPath)
		if err != nil {
			return errors.Wrap(err, "failed to read step result file")
		}
		if err := json.Unmarshal(c, &previousResult); err != nil {
			return errors.Wrap(err, "failed to unmarshal step result file")
		}
	}

	switch mode {
	case "pre":
		if err := execPre(ctx, stepIdx, executionInput, previousResult); err != nil {
			return err
		}

	case "post":
		if err := execPost(ctx, stepIdx, executionInput, previousResult); err != nil {
			return err
		}

	default:
		return errors.Newf("invalid mode %q", mode)
	}

	return nil
}
