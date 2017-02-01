package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"path"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/go-langserver/pkg/lspext"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/rcache"
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

// totalRefsCache is a redis cache to avoid some queries for popular
// repositories (which can take ~1s) from causing any serious performance
// issues when the request rate is high.
var (
	totalRefsCache        = rcache.NewWithTTL("totalrefs", 3600) // 1h
	totalRefsCacheCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "defs",
		Name:      "totalrefs_cache_hit",
		Help:      "Counts cache hits and misses for Defs.TotalRefs repo ref counts.",
	}, []string{"type"})
)

func init() {
	prometheus.MustRegister(totalRefsCacheCounter)
}

func (s *defs) TotalRefs(ctx context.Context, source string) (res int, err error) {
	ctx, done := trace(ctx, "Deps", "TotalRefs", source, &err)
	defer done()

	// Check if value is in the cache.
	jsonRes, ok := totalRefsCache.Get(source)
	if ok {
		totalRefsCacheCounter.WithLabelValues("hit").Inc()
		if err := json.Unmarshal(jsonRes, &res); err != nil {
			return 0, err
		}
		return res, nil
	}

	// Query value from the database.
	totalRefsCacheCounter.WithLabelValues("miss").Inc()
	res, err = localstore.GlobalDeps.TotalRefs(ctx, source)
	if err != nil {
		return 0, err
	}

	// Store value in the cache.
	jsonRes, err = json.Marshal(res)
	if err != nil {
		return 0, err
	}
	totalRefsCache.Set(source, jsonRes)
	return res, nil
}

func (s *defs) DeprecatedTotalRefs(ctx context.Context, repoURI string) (res int, err error) {
	ctx, done := trace(ctx, "Defs", "DeprecatedTotalRefs", repoURI, &err)
	defer done()
	return localstore.DeprecatedGlobalRefs.DeprecatedTotalRefs(ctx, repoURI)
}

func (s *defs) DependencyReferences(ctx context.Context, op sourcegraph.DependencyReferencesOptions) (res *sourcegraph.DependencyReferences, err error) {
	ctx, done := trace(ctx, "Defs", "RefLocations", op, &err)
	defer done()

	span := opentracing.SpanFromContext(ctx)
	span.SetTag("language", op.Language)
	span.SetTag("repo_id", op.RepoID)
	span.SetTag("commit_id", op.CommitID)
	span.SetTag("file", op.File)
	span.SetTag("line", op.Line)
	span.SetTag("character", op.Character)

	// SECURITY: We first must call textDocument/xdefinition on a ref
	// to figure out what to query the global deps database for. The
	// ref might exist in a private repo, so we MUST check that the
	// user has access to that private repo first prior to calling it
	// in xlang (xlang has unlimited, unchecked access to gitserver).
	//
	// For example, if a user is browsing a private repository but
	// looking for references to a public repository's symbol
	// (fmt.Println), we support that, but we DO NOT support looking
	// for references to a private repository's symbol ever (in fact,
	// they are not even indexed by the global deps database).
	if err := accesscontrol.VerifyUserHasReadAccess(ctx, "Defs.DependencyReferences", op.RepoID); err != nil {
		return nil, err
	}

	// Fetch repository information.
	repo, err := Repos.Get(ctx, &sourcegraph.RepoSpec{ID: op.RepoID})
	if err != nil {
		return nil, err
	}
	vcs := "git" // TODO: store VCS type in *sourcegraph.Repo object.
	span.SetTag("repo", repo.URI)

	// Determine the rootPath.
	rootPath := vcs + "://" + repo.URI + "?" + op.CommitID

	// Find the metadata for the definition specified by op, such that we can
	// perform the DB query using that metadata.
	var locations []lspext.SymbolLocationInformation
	err = xlang.UnsafeOneShotClientRequest(ctx, op.Language, rootPath, "textDocument/xdefinition", lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: rootPath + "#" + op.File},
		Position:     lsp.Position{Line: op.Line, Character: op.Character},
	}, &locations)
	if err != nil {
		return nil, errors.Wrap(err, "LSP textDocument/xdefinition")
	}
	if len(locations) == 0 {
		return nil, fmt.Errorf("textDocument/xdefinition returned zero locations")
	}

	// TODO(slimsag): figure out how to handle multiple location responses here
	// once we have a language server that uses it.
	location := locations[0]

	// If the symbol is not referenceable according to language semantics, then
	// there is no need to consult the database or perform roundtrips for
	// workspace/xreferences requests.
	var depRefs []*sourcegraph.DependencyReference
	if !xlang.IsSymbolReferenceable(op.Language, location.Symbol) {
		span.SetTag("nonreferencable", true)
	} else {
		pkgDescriptor, ok := xlang.SymbolPackageDescriptor(location.Symbol, op.Language)
		if !ok {
			return nil, err
		}

		depRefs, err = localstore.GlobalDeps.Dependencies(ctx, localstore.DependenciesOptions{
			Language: op.Language,
			DepData:  pkgDescriptor,
			Limit:    op.Limit,
		})
		if err != nil {
			return nil, err
		}
	}

	span.SetTag("# depRefs", len(depRefs))
	return &sourcegraph.DependencyReferences{
		References: depRefs,
		Location:   location,
	}, nil
}

// UnsafeRefreshIndex refreshes the global deps index for the specified repo@commit.
//
// SECURITY: It is the caller's responsibility to ensure the repository
// described by the op parameter is accurately specified as private or not.
func (s *defs) UnsafeRefreshIndex(ctx context.Context, op *sourcegraph.DefsRefreshIndexOp) (err error) {
	if Mocks.Defs.RefreshIndex != nil {
		return Mocks.Defs.RefreshIndex(ctx, op)
	}

	ctx, done := trace(ctx, "Defs", "RefreshIndex", op, &err)
	defer done()

	inv, err := Repos.GetInventory(ctx, &sourcegraph.RepoRevSpec{Repo: op.RepoID, CommitID: op.CommitID})
	if err != nil {
		return err
	}

	// Refresh global references indexes.
	return localstore.GlobalDeps.UnsafeRefreshIndex(ctx, op, inv.Languages)
}
