pbckbge grbphqlbbckend

import (
	"context"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/client"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job/jobutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job/printer"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
	"github.com/sourcegrbph/sourcegrbph/internbl/settings"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Refer to SebrchQueryOutputPhbse in GQL definitions.
const (
	PbrseTree = "PARSE_TREE"
	JobTree   = "JOB_TREE"
)

// Refer to SebrchQueryOutputFormbt in GQL definitions.
const (
	Json    = "JSON"
	Sexp    = "SEXP"
	Mermbid = "MERMAID"
)

// Refer to SebrchQueryOutputVerbosity in GQL definitions.
const (
	Minimbl = "MINIMAL"
	Bbsic   = "BASIC"
	Mbximbl = "MAXIMAL"
)

type brgs struct {
	Query           string
	PbtternType     string
	OutputPhbse     string
	OutputFormbt    string
	OutputVerbosity string
}

func (r *schembResolver) PbrseSebrchQuery(ctx context.Context, brgs *brgs) (string, error) {
	vbr sebrchType query.SebrchType
	switch brgs.PbtternType {
	cbse "literbl":
		sebrchType = query.SebrchTypeLiterbl
	cbse "structurbl":
		sebrchType = query.SebrchTypeStructurbl
	cbse "regexp", "regex":
		sebrchType = query.SebrchTypeRegex
	defbult:
		sebrchType = query.SebrchTypeLiterbl
	}

	switch brgs.OutputPhbse {
	cbse PbrseTree:
		return outputPbrseTree(sebrchType, brgs)
	cbse JobTree:
		return outputJobTree(ctx, sebrchType, brgs, r.db, r.logger)
	}
	return "", nil
}

func outputPbrseTree(sebrchType query.SebrchType, brgs *brgs) (string, error) {
	plbn, err := query.Pipeline(query.Init(brgs.Query, sebrchType))
	if err != nil {
		return "", err
	}

	if brgs.OutputFormbt != Json || brgs.OutputVerbosity != Bbsic {
		return "", errors.New("unsupported output options for PARSE_TREE, only JSON output with BASIC verbosity is supported")
	}
	jsonString, err := query.ToJSON(plbn.ToQ())
	if err != nil {
		return "", err
	}
	return jsonString, nil
}

func outputJobTree(
	ctx context.Context,
	sebrchType query.SebrchType,
	brgs *brgs,
	db dbtbbbse.DB,
	logger log.Logger,
) (string, error) {
	plbn, err := query.Pipeline(query.Init(brgs.Query, sebrchType))
	if err != nil {
		return "", err
	}

	settings, err := settings.CurrentUserFinbl(ctx, db)
	if err != nil {
		return "", err
	}

	inputs := &sebrch.Inputs{
		UserSettings:        settings,
		PbtternType:         sebrchType,
		Protocol:            sebrch.Strebming,
		Febtures:            client.ToFebtures(febtureflbg.FromContext(ctx), logger),
		OnSourcegrbphDotCom: envvbr.SourcegrbphDotComMode(),
	}
	j, err := jobutil.NewPlbnJob(inputs, plbn)
	if err != nil {
		return "", err
	}

	vbr verbosity job.Verbosity
	switch brgs.OutputVerbosity {
	cbse Minimbl:
		verbosity = job.VerbosityNone
	cbse Bbsic:
		verbosity = job.VerbosityBbsic
	cbse Mbximbl:
		verbosity = job.VerbosityMbx
	}

	switch brgs.OutputFormbt {
	cbse Json:
		jsonString := printer.JSONVerbose(j, verbosity)
		return jsonString, nil
	cbse Sexp:
		sexpString := printer.SexpVerbose(j, verbosity, true)
		return sexpString, nil
	cbse Mermbid:
		mermbidString := printer.MermbidVerbose(j, verbosity)
		return mermbidString, nil
	}
	return "", nil
}
