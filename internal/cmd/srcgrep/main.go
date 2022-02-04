package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"golang.org/x/sync/errgroup"
)

func run(ctx context.Context, node query.Node, parameters []query.Parameter) error {
	pattern, ok := node.(query.Pattern)
	if !ok {
		return fmt.Errorf("only supports pattern queries, got %T: %s", node, node)
	}

	var args []string
	var repoParams []query.Parameter

	for _, p := range parameters {
		switch p.Field {
		case query.FieldCase:
			switch p.Value {
			case "yes":
				args = append(args, "--case-sensitive")
			case "no":
				args = append(args, "--ignore-case")
			default:
				return fmt.Errorf("unknown case value: %s", p)
			}

		case query.FieldRepo:
			repoParams = append(repoParams, p)

		// case query.FieldFile:
		// case	query.FieldFork:
		// case	query.FieldArchived:
		// case	query.FieldLang:
		// case	query.FieldType:
		// case	query.FieldRepoHasFile:
		// case	query.FieldRepoHasCommitAfter:
		// case	query.FieldPatternType:
		// case	query.FieldContent:
		// case	query.FieldVisibility:
		// case	query.FieldRev:
		// case	query.FieldContext:

		default:
			return fmt.Errorf("unsupported field: %s", p)
		}
	}

	args = append(args, pattern.Value)

	cmd := exec.CommandContext(ctx, "rg", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func do(q string) error {
	plan, err := query.Pipeline(
		query.Init(q, query.SearchTypeRegex),
	)
	if err != nil {
		return err
	}

	wg, ctx := errgroup.WithContext(context.Background())

	for _, p := range plan {
		p := p
		wg.Go(func() error {
			return run(ctx, p.Pattern, p.Parameters)
		})
	}

	return wg.Wait()
}

func main() {
	err := do(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
}
