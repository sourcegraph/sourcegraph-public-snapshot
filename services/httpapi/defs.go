package httpapi

import (
	"errors"
	"net/http"

	"context"

	"github.com/gorilla/mux"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/feature"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/universe"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/langp"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"
	"sourcegraph.com/sourcegraph/srclib/graph"
)

type serveDefOpt struct {
	sourcegraph.DefGetOptions
	File            string
	Line, Character int
}

func serveDef(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)

	var opt serveDefOpt
	err := schemaDecoder.Decode(&opt, r.URL.Query())
	if err != nil {
		return err
	}

	var def *sourcegraph.Def
	repo, err := handlerutil.GetRepo(r.Context(), vars)
	if err != nil {
		return err
	}
	if feature.Features.NoSrclib && universe.EnabledRepo(repo) {
		def, err = universeDef(opt, r, repo)
		if err != nil {
			return err
		}
	} else if universe.Shadow(repo.URI) {
		go universeDef(opt, r, repo)
	}

	if def == nil {
		def, _, err = handlerutil.GetDefCommon(r.Context(), vars, &opt.DefGetOptions)
		if err != nil {
			return err
		}
	}
	return writeJSON(w, def)
}

func universeDef(opt serveDefOpt, r *http.Request, repo *sourcegraph.Repo) (*sourcegraph.Def, error) {
	vars := mux.Vars(r)
	cl := handlerutil.Client(r)
	// TODO(slimsag): The URLs for this are quite ugly when omitting the
	// defkey:
	//
	//  /.api/repos/github.com/slimsag/mux/-/def/-/-/-/-?File=mux.go&Line=57&Character=17
	//
	// We should consider ways of making this cleaner.

	// TODO(slimsag): This code does not fill out a number of
	// sourcegraph.Def fields (in fact, it's easier to list the ones
	// that it does fill out). We should change this endpoint to return
	// only data that the frontend actually needs. We only return the
	// ones that the DefInfo page needs here.
	def := &sourcegraph.Def{}
	if opt.Doc {
		return nil, errors.New("httpapi: serveDef: DefGetOptions.Doc not implemented by universe")
	}
	if opt.ComputeLineRange == true {
		return nil, errors.New("httpapi: serveDef: DefGetOptions.ComputeLineRange not implemented by universe")
	}
	def.Def = graph.Def{}

	defSpec := routevar.ToDefAtRev(vars)

	// Determine commit ID based on the request.
	repoRev := routevar.ToRepoRev(vars)
	res, err := cl.Repos.ResolveRev(r.Context(), &sourcegraph.ReposResolveRevOp{
		Repo: repo.ID,
		Rev:  repoRev.Rev,
	})
	if err != nil {
		return nil, err
	}

	if opt.File == "" {
		lpDefSpec, err := langp.DefaultClient.DefSpecToPosition(r.Context(), &langp.DefSpec{
			Repo:     repo.URI,
			Commit:   res.CommitID,
			UnitType: defSpec.UnitType,
			Unit:     defSpec.Unit,
			Path:     defSpec.Path,
		})
		universeObserve("DefSpecToPosition", err)
		if err != nil {
			return nil, err
		}
		opt.File = lpDefSpec.File
		opt.Line = lpDefSpec.Line
		opt.Character = lpDefSpec.Character
	}

	hover, err := langp.DefaultClient.Hover(r.Context(), &langp.Position{
		Repo:      repo.URI,
		Commit:    res.CommitID,
		File:      opt.File,
		Line:      opt.Line,
		Character: opt.Character,
	})
	universeObserve("DefHover", err)
	if err != nil {
		return nil, err
	}

	return &sourcegraph.Def{
		Def: graph.Def{
			DefKey: graph.DefKey{
				Repo:     repo.URI,
				CommitID: res.CommitID,
				UnitType: defSpec.UnitType,
				Unit:     defSpec.Unit,
				Path:     defSpec.Path,
			},
			File: opt.File,
		},
		FmtStrings: &graph.DefFormatStrings{
			Name: graph.QualFormatStrings{
				ScopeQualified: hover.Title,
				DepQualified:   hover.Title,
			},
		},
	}, nil
}

func resolveDef(ctx context.Context, def routevar.DefAtRev) (*sourcegraph.DefSpec, error) {
	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return nil, err
	}
	res, err := cl.Repos.Resolve(ctx, &sourcegraph.RepoResolveOp{Path: def.Repo})
	if err != nil {
		return nil, err
	}
	rev, err := cl.Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{Repo: res.Repo, Rev: def.Rev})
	if err != nil {
		return nil, err
	}
	return &sourcegraph.DefSpec{
		Repo:     res.Repo,
		CommitID: rev.CommitID,
		UnitType: def.UnitType,
		Unit:     def.Unit,
		Path:     def.Path,
	}, nil
}
