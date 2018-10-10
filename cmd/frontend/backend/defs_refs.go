package backend

import (
	"context"
	"encoding/json"
	"fmt"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/go-lsp/lspext"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/rcache"
	"github.com/sourcegraph/sourcegraph/xlang"
)

var Defs = &defs{}

type defs struct{}

// totalRefsCache is a redis cache to avoid some queries for popular
// repositories (which can take ~1s) from causing any serious performance
// issues when the request rate is high.
var (
	totalRefsCache        = rcache.NewWithTTL("totalrefs", 4*60*60) // 4h
	totalRefsCacheCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "defs",
		Name:      "totalrefs_cache_hit",
		Help:      "Counts cache hits and misses for Defs.TotalRefs repo ref counts.",
	}, []string{"type"})

	listTotalRefsCache        = rcache.NewWithTTL("listtotalrefs", 4*60*60) // 4h
	listTotalRefsCacheCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "defs",
		Name:      "listtotalrefs_cache_hit",
		Help:      "Counts cache hits and misses for Defs.ListTotalRefs repo ref counts.",
	}, []string{"type"})
)

func init() {
	prometheus.MustRegister(totalRefsCacheCounter)
	prometheus.MustRegister(listTotalRefsCacheCounter)
}

func (s *defs) TotalRefs(ctx context.Context, source api.RepoURI) (res int, err error) {
	if Mocks.Defs.TotalRefs != nil {
		return Mocks.Defs.TotalRefs(ctx, source)
	}

	ctx, done := trace(ctx, "Deps", "TotalRefs", source, &err)
	defer done()

	// Check if value is in the cache.
	jsonRes, ok := totalRefsCache.Get(string(source))
	if ok {
		totalRefsCacheCounter.WithLabelValues("hit").Inc()
		if err := json.Unmarshal(jsonRes, &res); err != nil {
			return 0, err
		}
		return res, nil
	}

	// Query value from the database.
	rp, err := Repos.GetByURI(ctx, source)
	if err != nil {
		return 0, err
	}
	commitID, err := Repos.ResolveRev(ctx, rp, "")
	if err != nil {
		return 0, err
	}
	inv, err := Repos.GetInventory(ctx, rp, commitID)
	if err != nil {
		return 0, err
	}
	totalRefsCacheCounter.WithLabelValues("miss").Inc()
	res, err = db.GlobalDeps.TotalRefs(ctx, rp, inv.Languages)
	if err != nil {
		return 0, err
	}

	// Store value in the cache.
	jsonRes, err = json.Marshal(res)
	if err != nil {
		return 0, err
	}
	totalRefsCache.Set(string(source), jsonRes)
	return res, nil
}

func (s *defs) ListTotalRefs(ctx context.Context, source api.RepoURI) (repos []api.RepoID, err error) {
	if Mocks.Defs.ListTotalRefs != nil {
		return Mocks.Defs.ListTotalRefs(ctx, source)
	}

	ctx, done := trace(ctx, "Deps", "ListTotalRefs", source, &err)
	defer done()

	// Check if value is in the cache.
	jsonRes, ok := listTotalRefsCache.Get(string(source))
	if ok {
		listTotalRefsCacheCounter.WithLabelValues("hit").Inc()
		if err := json.Unmarshal(jsonRes, &repos); err != nil {
			return nil, err
		}
		return repos, nil
	}

	// Query value from the database.
	rp, err := Repos.GetByURI(ctx, source)
	if err != nil {
		return nil, err
	}
	commitID, err := Repos.ResolveRev(ctx, rp, "")
	if err != nil {
		return nil, err
	}
	inv, err := Repos.GetInventory(ctx, rp, commitID)
	if err != nil {
		return nil, err
	}
	listTotalRefsCacheCounter.WithLabelValues("miss").Inc()
	repos, err = db.GlobalDeps.ListTotalRefs(ctx, rp, inv.Languages)
	if err != nil {
		return nil, err
	}

	// Store value in the cache.
	_ = []api.RepoID(repos) // important so that we don't accidentally change encoding type
	jsonRes, err = json.Marshal(repos)
	if err != nil {
		return nil, err
	}
	listTotalRefsCache.Set(string(source), jsonRes)
	return repos, nil
}

func (s *defs) DependencyReferences(ctx context.Context, op types.DependencyReferencesOptions) (res *api.DependencyReferences, err error) {
	if Mocks.Defs.DependencyReferences != nil {
		return Mocks.Defs.DependencyReferences(ctx, op)
	}

	ctx, done := trace(ctx, "Defs", "RefLocations", op, &err)
	defer done()

	span := opentracing.SpanFromContext(ctx)
	span.SetTag("language", op.Language)
	span.SetTag("repo_id", op.RepoID)
	span.SetTag("commit_id", op.CommitID)
	span.SetTag("file", op.File)
	span.SetTag("line", op.Line)
	span.SetTag("character", op.Character)

	repo, err := Repos.Get(ctx, op.RepoID)
	if err != nil {
		return nil, err
	}
	vcs := "git" // TODO: store VCS type in *types.Repo object.
	span.SetTag("repo", repo.URI)

	// Determine the rootURI.
	rootURI := lsp.DocumentURI(vcs + "://" + string(repo.URI) + "?" + string(op.CommitID))

	// Find the metadata for the definition specified by op, such that we can
	// perform the DB query using that metadata.
	var locations []lspext.SymbolLocationInformation
	err = xlang.UnsafeOneShotClientRequest(ctx, op.Language, rootURI, "textDocument/xdefinition", lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: lsp.DocumentURI(string(rootURI) + "#" + op.File)},
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
	var depRefs []*api.DependencyReference
	if !xlang.IsSymbolReferenceable(op.Language, location.Symbol) {
		span.SetTag("nonreferencable", true)
	} else {
		pkgDescriptor, ok := xlang.SymbolPackageDescriptor(location.Symbol, op.Language)
		if !ok {
			return nil, err
		}

		depRefs, err = db.GlobalDeps.Dependencies(ctx, db.DependenciesOptions{
			Language: op.Language,
			DepData:  pkgDescriptor,
			Limit:    op.Limit,
		})
		if err != nil {
			return nil, err
		}
	}

	span.SetTag("# depRefs", len(depRefs))
	return &api.DependencyReferences{
		References: depRefs,
		Location:   location,
	}, nil
}

type MockDefs struct {
	TotalRefs            func(ctx context.Context, source api.RepoURI) (res int, err error)
	ListTotalRefs        func(ctx context.Context, source api.RepoURI) (repos []api.RepoID, err error)
	DependencyReferences func(ctx context.Context, op types.DependencyReferencesOptions) (res *api.DependencyReferences, err error)
}
