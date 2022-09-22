package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/client"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/job/jobutil"
	"github.com/sourcegraph/sourcegraph/internal/search/job/printer"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const (
	// cf. SearchQueryOutputPhase in GQL definitions.
	ParseTree = "PARSE_TREE"
	JobTree   = "JOB_TREE"

	// cf. SearchQueryOutputFormat in GQL definitions.
	Json    = "JSON"
	Sexp    = "SEXP"
	Mermaid = "MERMAID"

	// cf. SearchQueryOutputVerbosity in GQL definitions.
	Minimal = "MINIMAL"
	Basic   = "BASIC"
	Maximal = "MAXIMAL"
)

type args struct {
	Query           string
	PatternType     string
	OutputPhase     string
	OutputFormat    string
	OutputVerbosity string
}

func (r *schemaResolver) ParseSearchQuery(ctx context.Context, args *args) (*JSONValue, error) {
	var searchType query.SearchType
	switch args.PatternType {
	case "literal":
		searchType = query.SearchTypeLiteral
	case "structural":
		searchType = query.SearchTypeStructural
	case "regexp", "regex":
		searchType = query.SearchTypeRegex
	default:
		searchType = query.SearchTypeLiteral
	}

	switch args.OutputPhase {
	case ParseTree:
		return outputParseTree(searchType, args)
	case JobTree:
		return outputJobTree(ctx, searchType, args, r.db, r.logger)
	}
	return nil, nil
}

func outputParseTree(searchType query.SearchType, args *args) (*JSONValue, error) {
	plan, err := query.Pipeline(query.Init(args.Query, searchType))
	if err != nil {
		return nil, err
	}

	if args.OutputFormat != Json || args.OutputVerbosity != Minimal {
		return nil, errors.New("unsupported output options for PARSE_TREE, only JSON output with MINIMAL verbosity is supported")
	}
	jsonString, err := query.ToJSON(plan.ToQ())
	if err != nil {
		return nil, err
	}
	return &JSONValue{Value: jsonString}, nil
}

func outputJobTree(
	ctx context.Context,
	searchType query.SearchType,
	args *args,
	db database.DB,
	logger log.Logger,
) (*JSONValue, error) {
	plan, err := query.Pipeline(query.Init(args.Query, searchType))
	if err != nil {
		return nil, err
	}

	settings, err := DecodedViewerFinalSettings(ctx, db)
	if err != nil {
		return nil, err
	}

	inputs := &search.Inputs{
		UserSettings:        settings,
		PatternType:         searchType,
		Protocol:            search.Streaming,
		Features:            client.ToFeatures(featureflag.FromContext(ctx), logger),
		OnSourcegraphDotCom: envvar.SourcegraphDotComMode(),
	}
	j, err := jobutil.NewPlanJob(inputs, plan)
	if err != nil {
		return nil, err
	}

	var verbosity job.Verbosity
	switch args.OutputVerbosity {
	case Minimal:
		verbosity = job.VerbosityNone
	case Basic:
		verbosity = job.VerbosityBasic
	case Maximal:
		verbosity = job.VerbosityMax
	}

	switch args.OutputFormat {
	case Json:
		jsonString := printer.JSONVerbose(j, verbosity)
		return &JSONValue{Value: jsonString}, nil
	case Sexp:
		sexpString := printer.SexpVerbose(j, verbosity, true)
		return &JSONValue{Value: sexpString}, nil
	case Mermaid:
		mermaidString := printer.MermaidVerbose(j, verbosity)
		return &JSONValue{Value: mermaidString}, nil
	}
	return nil, nil
}
