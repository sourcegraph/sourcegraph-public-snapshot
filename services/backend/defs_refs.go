package backend

import (
	"context"
	"net/url"
	"path"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	log15 "gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/traceutil"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/localstore"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang"
	"sourcegraph.com/sourcegraph/srclib/graph"
	srcstore "sourcegraph.com/sourcegraph/srclib/store"
)

func (s *defs) DeprecatedListRefs(ctx context.Context, op *sourcegraph.DeprecatedDefsListRefsOp) (res *sourcegraph.RefList, err error) {
	if Mocks.Defs.ListRefs != nil {
		return Mocks.Defs.ListRefs(ctx, op)
	}

	ctx, done := trace(ctx, "Defs", "ListRefs", op, &err)
	defer done()

	defSpec := op.Def
	opt := op.Opt
	if opt == nil {
		opt = &sourcegraph.DeprecatedDefListRefsOptions{}
	}

	// Restrict the ref search to a single repo and commit for performance.
	if opt.Repo == 0 && defSpec.Repo != 0 {
		opt.Repo = defSpec.Repo
	}
	if opt.CommitID == "" {
		opt.CommitID = defSpec.CommitID
	}
	if opt.Repo == 0 {
		return nil, legacyerr.Errorf(legacyerr.InvalidArgument, "ListRefs: Repo must be specified")
	}
	if opt.CommitID == "" {
		return nil, legacyerr.Errorf(legacyerr.InvalidArgument, "ListRefs: CommitID must be specified")
	}

	defRepoObj, err := Repos.Get(ctx, &sourcegraph.RepoSpec{ID: defSpec.Repo})
	if err != nil {
		return nil, err
	}
	if err := accesscontrol.VerifyUserHasReadAccess(ctx, "Defs.ListRefs", defRepoObj.ID); err != nil {
		return nil, err
	}

	refRepoObj, err := Repos.Get(ctx, &sourcegraph.RepoSpec{ID: opt.Repo})
	if err != nil {
		return nil, err
	}
	if err := accesscontrol.VerifyUserHasReadAccess(ctx, "Defs.ListRefs", refRepoObj.ID); err != nil {
		return nil, err
	}

	repoFilters := []srcstore.RefFilter{
		srcstore.ByRepos(refRepoObj.URI),
		srcstore.ByCommitIDs(opt.CommitID),
	}

	refFilters := []srcstore.RefFilter{
		srcstore.ByRefDef(graph.RefDefKey{
			DefRepo:     defRepoObj.URI,
			DefUnitType: defSpec.UnitType,
			DefUnit:     defSpec.Unit,
			DefPath:     defSpec.Path,
		}),
		srcstore.ByCommitIDs(opt.CommitID),
		srcstore.RefFilterFunc(func(ref *graph.Ref) bool { return !ref.Def }),
		srcstore.Limit(opt.Offset()+opt.Limit()+1, 0),
	}

	if len(opt.Files) > 0 {
		for i, f := range opt.Files {
			// Files need to be clean or else graphstore will panic.
			opt.Files[i] = path.Clean(f)
		}
		refFilters = append(refFilters, srcstore.ByFiles(false, opt.Files...))
	}

	filters := append(repoFilters, refFilters...)
	bareRefs, err := localstore.Graph.Refs(filters...)
	if err != nil {
		return nil, err
	}

	// Convert to sourcegraph.Ref and file bareRefs.
	refs := make([]*graph.Ref, 0, opt.Limit())
	for i, bareRef := range bareRefs {
		if i >= opt.Offset() && i < (opt.Offset()+opt.Limit()) {
			refs = append(refs, bareRef)
		}
	}
	hasMore := len(bareRefs) > opt.Offset()+opt.Limit()

	return &sourcegraph.RefList{
		Refs:           refs,
		StreamResponse: sourcegraph.StreamResponse{HasMore: hasMore},
	}, nil
}

func (s *defs) DeprecatedListRefLocations(ctx context.Context, op *sourcegraph.DeprecatedDefsListRefLocationsOp) (res *sourcegraph.DeprecatedRefLocationsList, err error) {
	if Mocks.Defs.ListRefLocations != nil {
		return Mocks.Defs.ListRefLocations(ctx, op)
	}

	ctx, done := trace(ctx, "Defs", "ListRefLocations", op, &err)
	defer done()

	return localstore.DeprecatedGlobalRefs.DeprecatedGet(ctx, op)
}

func (s *defs) TopDefs(ctx context.Context, op sourcegraph.TopDefsOptions) (res *sourcegraph.TopDefs, err error) {
	ctx, done := trace(ctx, "Defs", "TopDefs", op, &err)
	defer done()
	return localstore.GlobalRefs.TopDefs(ctx, op)
}

func (s *defs) RefLocations(ctx context.Context, op sourcegraph.RefLocationsOptions) (res *sourcegraph.RefLocations, err error) {
	ctx, done := trace(ctx, "Defs", "RefLocations", op, &err)
	defer done()

	// Query repo-local references.
	localRefs, err := s.localRefLocations(ctx, op)
	if err != nil {
		return nil, errors.Wrap(err, "localRefLocations")
	}

	// Query global references.
	op.Sources -= 1 // localRefs provides us one.
	refLocs, err := localstore.GlobalRefs.RefLocations(ctx, op)
	if err != nil {
		return nil, errors.Wrap(err, "localstore.GlobalRefs.RefLocations")
	}

	// Combine the results.
	refLocs.TotalSources++
	refLocs.SourceRefs = append([]*sourcegraph.SourceRef{localRefs}, refLocs.SourceRefs...)
	return refLocs, nil
}

// localRefLocations returns reference locations for op.Source only, i.e.
// references located inside a repository (op.Source) to a definition (op.Name, op.ContainerName)
// that is also located inside the same repository (op.Source).
func (s *defs) localRefLocations(ctx context.Context, op sourcegraph.RefLocationsOptions) (res *sourcegraph.SourceRef, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "localRefLocations")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	// Resolve the commit ID of op.Source's default branch.
	repo, err := Repos.Resolve(ctx, &sourcegraph.RepoResolveOp{Path: op.Source, Remote: true})
	if err != nil {
		return nil, errors.Wrap(err, "Repos.Resolve")
	}
	rev, err := Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{Repo: repo.Repo})
	if err != nil {
		return nil, errors.Wrap(err, "Repos.ResolveRev")
	}

	// In order to invoke textDocument/references, we need to resolve (op.Name, op.ContainerName)
	// into a concrete location. Query the workspace's symbols to do this.
	language := "go" // TODO: Go specific
	rootPath := "git://" + op.Source + "?" + rev.CommitID
	var symbols []lsp.SymbolInformation
	err = xlang.OneShotClientRequest(ctx, language, rootPath, "workspace/symbol", lsp.WorkspaceSymbolParams{
		Query: op.Name,
		Limit: 100,
	}, &symbols)
	if err != nil {
		return nil, errors.Wrap(err, "LSP workspace/symbol")
	}

	// Find the matching symbol.
	var symbol *lsp.SymbolInformation
	for _, sym := range symbols {
		if sym.Name != op.Name || sym.ContainerName != op.ContainerName {
			continue
		}
		withoutFile, err := url.Parse(sym.Location.URI)
		if err != nil {
			return nil, errors.Wrap(err, "parsing symbol location URI")
		}
		withoutFile.Fragment = ""
		if withoutFile.String() != rootPath {
			continue
		}
		symbol = &sym
		break
	}
	if symbol == nil {
		log15.Warn("RefLocations: no symbol info matching top def from global references", "trace", traceutil.SpanURL(opentracing.SpanFromContext(ctx)))
		return nil, errors.New("RefLocations: no symbol info matching top def from global references")
	}

	// Invoke textDocument/references in order to find references to the symbol.
	var refs []lsp.Location
	err = xlang.OneShotClientRequest(ctx, language, rootPath, "textDocument/references", lsp.ReferenceParams{
		Context: lsp.ReferenceContext{
			IncludeDeclaration: false,
		},
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: symbol.Location.URI},
			Position:     symbol.Location.Range.Start,
		},
	}, &refs)
	if err != nil {
		return nil, errors.Wrap(err, "LSP textDocument/references")
	}

	// Formulate a proper *sourcegraph.SourceRef response by de-duplicating the
	// references by accumulating their positions via the fileRefsMap.
	var (
		fileRefsMap = make(map[string]*sourcegraph.FileRef)
		fileRefs    []*sourcegraph.FileRef
		refsCount   int
	)
	for i, ref := range refs {
		refsCount++
		if i > op.Files {
			break
		}
		if r, ok := fileRefsMap[ref.URI]; ok {
			r.Positions = append(r.Positions, ref.Range)
			continue
		}

		uri, err := url.Parse(ref.URI)
		if err != nil {
			return nil, errors.Wrap(err, "parsing reference URI")
		}
		r := &sourcegraph.FileRef{
			Scheme:    uri.Scheme,
			Source:    uri.Host + uri.Path,
			Version:   uri.RawQuery,
			File:      uri.Fragment,
			Positions: []lsp.Range{ref.Range},
			Score:     0,
		}
		fileRefsMap[ref.URI] = r
		fileRefs = append(fileRefs, r)
	}

	sourceURI, err := url.Parse(symbol.Location.URI)
	if err != nil {
		return nil, errors.Wrap(err, "parsing symbol location URI")
	}

	return &sourcegraph.SourceRef{
		Scheme:   sourceURI.Scheme,
		Source:   sourceURI.Host + sourceURI.Path,
		Version:  sourceURI.RawQuery,
		Files:    len(fileRefs),
		Refs:     refsCount,
		Score:    0,
		FileRefs: fileRefs,
	}, nil
}

var indexDuration = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: "src",
	Name:      "index_duration_seconds",
	Help:      "Duration of time that indexing a repository takes.",
})

func init() {
	prometheus.MustRegister(indexDuration)
}

func (s *defs) RefreshIndex(ctx context.Context, op *sourcegraph.DefsRefreshIndexOp) (err error) {
	start := time.Now()
	defer func() {
		indexDuration.Set(time.Since(start).Seconds())
	}()

	if Mocks.Defs.RefreshIndex != nil {
		return Mocks.Defs.RefreshIndex(ctx, op)
	}

	ctx, done := trace(ctx, "Defs", "RefreshIndex", op, &err)
	defer done()

	rev, err := Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{Repo: op.Repo})
	if err != nil {
		return err
	}

	// Refresh global references indexes.
	return localstore.GlobalRefs.RefreshIndex(ctx, op.Repo, rev.CommitID)
}
