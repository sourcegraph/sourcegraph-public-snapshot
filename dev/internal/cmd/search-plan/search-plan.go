// Command search-plan is a debug helper which outputs the plan for a query.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/client"
	"github.com/sourcegraph/sourcegraph/internal/search/job/jobutil"
	"github.com/sourcegraph/sourcegraph/internal/search/job/printer"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func run(w io.Writer, args []string) error {
	fs := flag.NewFlagSet(args[0], flag.ExitOnError)

	version := fs.String("version", "V3", "the version of the search API to use")
	patternType := fs.String("pattern_type", "", "optionally specify query.PatternType (regex, literal, ...)")
	smartSearch := fs.Bool("smart_search", false, "enable smart search mode instead of precise")
	dotCom := fs.Bool("dotcom", false, "enable sourcegraph.com parsing rules")

	fs.Parse(args[1:])
	if narg := fs.NArg(); narg != 1 {
		return errors.Errorf("expected 1 argument for the query got %d", narg)
	}

	// Further argument parsing
	query := fs.Arg(0)
	mode := search.Precise
	if *smartSearch {
		mode = search.SmartSearch
	}
	settings := &schema.Settings{}

	// Sourcegraph infra we need
	conf.Mock(&conf.Unified{})
	logger := log.Scoped("search-plan", "")

	cli := client.NewSearchClient(logger, nil, nil, nil, nil)

	inputs, err := cli.Plan(
		context.Background(),
		*version,
		strPtr(*patternType),
		query,
		mode,
		search.Streaming,
		settings,
		*dotCom,
	)
	if err != nil {
		return errors.Wrap(err, "failed to plan")
	}

	fmt.Fprintln(w, "plan", inputs.Plan)
	fmt.Fprintln(w, "query", inputs.Query)

	planJob, err := jobutil.NewPlanJob(inputs, inputs.Plan, nil)
	if err != nil {
		return errors.Wrap(err, "failed to create planJob")
	}
	fmt.Fprintln(w, printer.SexpPretty(planJob))

	return nil
}

func main() {
	liblog := log.Init(log.Resource{Name: "search-plan"})
	defer liblog.Sync()

	err := run(os.Stdout, os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
