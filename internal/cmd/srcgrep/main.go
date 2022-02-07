package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"golang.org/x/sync/errgroup"
)

func run(ctx context.Context, rgBaseArgs []string, node query.Node, parameters []query.Parameter) error {
	plan, err := plan(node, parameters)
	if err != nil {
		return err
	}

	var paths []string
	// look for git directories matching repoParams
	if len(plan.RepoParameters) > 0 {
		return fmt.Errorf("not implemented")
	}

	var args []string
	args = append(args, rgBaseArgs...)
	args = append(args, plan.RipGrepArgs...)
	args = append(args, paths...)

	cmd := exec.CommandContext(ctx, "rg", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type Plan struct {
	RepoParameters []query.Parameter

	RipGrepArgs []string
}

func plan(node query.Node, parameters []query.Parameter) (*Plan, error) {
	pattern, ok := node.(query.Pattern)
	if !ok {
		return nil, fmt.Errorf("only supports pattern queries, got %T: %s", node, node)
	}

	var args []string
	var repoParams []query.Parameter

	for _, p := range parameters {
		switch p.Field {
		case query.FieldRepo:
			repoParams = append(repoParams, p)

		case query.FieldFile:
			return nil, fmt.Errorf("need to implement regex to glob for file: patterns")

		case query.FieldCase:
			switch p.Value {
			case "yes":
				args = append(args, "--case-sensitive")
			case "no":
				args = append(args, "--ignore-case")
			default:
				return nil, fmt.Errorf("unknown case value: %s", p)
			}

		case query.FieldLang:
			if p.Negated {
				args = append(args, "--type-not", p.Value)
			} else {
				args = append(args, "--type", p.Value)
			}

		default:
			return nil, fmt.Errorf("unsupported field: %s", p)
		}
	}

	if pattern.Value != "" {
		args = append(args, "--", pattern.Value)
	} else {
		args = append(args, "--files")
	}

	return &Plan{
		RepoParameters: repoParams,
		RipGrepArgs:    args,
	}, nil
}

func parse(q string) (query.Plan, error) {
	return query.Pipeline(
		query.Init(q, query.SearchTypeRegex),
	)
}

func do(rgArgs []string, q string) error {
	plan, err := parse(q)
	if err != nil {
		return err
	}

	wg, ctx := errgroup.WithContext(context.Background())

	for _, p := range plan {
		p := p
		wg.Go(func() error {
			return run(ctx, rgArgs, p.Pattern, p.Parameters)
		})
	}

	return wg.Wait()
}

func main() {
	if len(os.Args) <= 1 {
		fmt.Fprintf(os.Stderr, "USAGE: %s [RIPGREP_OPTIONS] SOURCEGRAPH_QUERY\n", filepath.Base(os.Args[0]))
		os.Exit(2)
	}

	queryIdx := len(os.Args) - 1
	err := do(os.Args[1:queryIdx], os.Args[queryIdx])
	if err != nil {
		log.Fatal(err)
	}
}
