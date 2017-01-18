package backend

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"path"
	"runtime"
	"sort"
	"time"

	"github.com/neelance/parallel"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/go-langserver/pkg/lspext"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/localstore"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang"
	"sourcegraph.com/sourcegraph/srclib/graph"
	srcstore "sourcegraph.com/sourcegraph/srclib/store"
)

// subSelectors is a map of language-specific data selectors. The
// input data is from the language server's workspace/xdefinition
// request, and the output data should be something that can be
// matched (using the jsonb containment operator) against the
// `attributes` field of `DependenceReference` (output of
// workspace/xdependencies).
var subSelectors = map[string]func(lspext.SymbolDescriptor) map[string]interface{}{
	"go": func(symbol lspext.SymbolDescriptor) map[string]interface{} {
		return map[string]interface{}{
			"package": symbol["package"],
		}
	},
	"php": func(symbol lspext.SymbolDescriptor) map[string]interface{} {
		if _, ok := symbol["package"]; !ok {
			// package can be missing if the symbol did not belong to a package, e.g. a project without
			// a composer.json file. In this case, there are no external references to this symbol.
			return nil
		}
		return map[string]interface{}{
			"name": symbol["package"].(map[string]interface{})["name"],
		}
	},
	"typescript": func(symbol lspext.SymbolDescriptor) map[string]interface{} {
		return map[string]interface{}{
			"name": symbol["package"].(map[string]interface{})["name"],
		}
	},
}

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

func (s *defs) TotalRefs(ctx context.Context, source string) (res int, err error) {
	ctx, done := trace(ctx, "Deps", "TotalRefs", source, &err)
	defer done()
	return localstore.GlobalDeps.TotalRefs(ctx, source)
}

func (s *defs) DeprecatedTotalRefs(ctx context.Context, repoURI string) (res int, err error) {
	ctx, done := trace(ctx, "Defs", "DeprecatedTotalRefs", repoURI, &err)
	defer done()
	return localstore.DeprecatedGlobalRefs.DeprecatedTotalRefs(ctx, repoURI)
}

func (s *defs) RefLocations(ctx context.Context, op sourcegraph.RefLocationsOptions) (res *sourcegraph.RefLocations, err error) {
	ctx, done := trace(ctx, "Defs", "RefLocations", op, &err)
	defer done()

	subSelector, ok := subSelectors[op.Language]
	if !ok {
		return nil, errors.New("language not supported")
	}

	span := opentracing.SpanFromContext(ctx)
	span.SetTag("language", op.Language)
	span.SetTag("repo_id", op.RepoID)
	span.SetTag("file", op.File)
	span.SetTag("line", op.Line)
	span.SetTag("character", op.Character)

	// Fetch repository information.
	repo, err := Repos.Get(ctx, &sourcegraph.RepoSpec{ID: op.RepoID})
	if err != nil {
		return nil, err
	}
	vcs := "git" // TODO: store VCS type in *sourcegraph.Repo object.
	span.SetTag("repo", repo.URI)

	// SECURITY: DO NOT REMOVE THIS CHECK! If a repository is private we must
	// ensure we do not leak its existence (or worse, LSP response info). We do
	// not support private repository global references yet.
	if repo.Private {
		return nil, accesscontrol.ErrRepoNotFound
	}

	// Determine the rootPath.
	rev, err := Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{Repo: repo.ID, Rev: repo.DefaultBranch})
	if err != nil {
		return nil, err
	}
	rootPath := vcs + "://" + repo.URI + "?" + rev.CommitID

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

	const (
		// Fetch up to 20 dependency references from the DB.
		depRefsLimit = 20

		// Up to 4 parallel workspace/xreferences requests at a time.
		//
		// It is important that multiple requests run in parallel here, because
		// some requests may hit large repositories (like kubernetes) which take
		// significantly longer to get results from.
		parallelWorkspaceRefs = 4

		// If we find minWorkspaceRefs and the xreferences requests aggregation
		// overall is observed to have taken longer than idealMaxAggregationTime,
		// we will return directly in order for the user to see (less) results
		// sooner.
		minWorkspaceRefs        = 2
		idealMaxAggregationTime = 2 * time.Second

		// Regardless of whether or not results are found, timeout after this
		// time period of waiting for results to aggregated.
		aggregationTimeout = 10 * time.Second
	)

	depRefs, err := localstore.GlobalDeps.Dependencies(ctx, localstore.DependenciesOptions{
		Language: op.Language,
		DepData:  subSelector(location.Symbol),
		Limit:    depRefsLimit,
	})
	if err != nil {
		return nil, err
	}

	// Now that we've gotten a list of potential repositories from the database
	// we must ask a language server for references via workspace/xreferences.
	//
	// TODO(slimsag): handle pagination here when it is important to us.
	var (
		xreferencesStart                  = time.Now()
		results                           = make(chan []lspext.ReferenceInformation, parallelWorkspaceRefs)
		run                               = parallel.NewRun(parallelWorkspaceRefs)
		xreferencesCtx, xreferencesCancel = context.WithCancel(ctx)
	)
	defer xreferencesCancel()
	for _, ref := range depRefs {
		ref := ref
		// If the dependency reference is the repository we're searching in,
		// then simply ignore it. This happens often because logically "the
		// best repo to find references to X" would include the one the user is
		// searching in.
		if ref.RepoID == op.RepoID {
			continue
		}

		run.Acquire()
		go func() (err error) {
			// Prevent any uncaught panics from taking the entire server down.
			defer func() {
				if r := recover(); r != nil {
					// Same as net/http
					const size = 64 << 10
					buf := make([]byte, size)
					buf = buf[:runtime.Stack(buf, false)]
					log.Printf("ignoring panic during Refs.RefLocations\n%s", buf)
					return
				}
			}()

			ctx := xreferencesCtx
			var refs []lspext.ReferenceInformation

			defer func() {
				run.Release()
				if err != nil {
					// Flag the span as an error for metrics purposes.
					ext.Error.Set(span, true)
					span.SetTag("err", err)
				}
				results <- refs
			}()

			repo, err := Repos.Get(ctx, &sourcegraph.RepoSpec{ID: ref.RepoID})
			if err != nil {
				return errors.Wrap(err, "Repos.Get")
			}
			vcs := "git" // TODO: store VCS type in *sourcegraph.Repo object.
			span.LogEventWithPayload("xdependency", repo.URI)

			// Determine the rootPath.
			rev, err := Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{Repo: repo.ID, Rev: repo.DefaultBranch})
			if err != nil {
				return errors.Wrap(err, "Repos.ResolveRev")
			}
			rootPath := vcs + "://" + repo.URI + "?" + rev.CommitID

			err = xlang.UnsafeOneShotClientRequest(ctx, op.Language, rootPath, "workspace/xreferences", lspext.WorkspaceReferencesParams{
				Query: location.Symbol,
				Hints: ref.Hints,
			}, &refs)
			if err != nil {
				return errors.Wrap(err, "LSP Call workspace/xreferences")
			}
			span.LogEventWithPayload("xreferences for "+repo.URI, len(refs))
			return nil
		}()
	}

	allRefs := &sourcegraph.RefLocations{}
	timeout := time.After(aggregationTimeout)
aggregation:
	for range depRefs {
		select {
		case <-timeout:
			span.LogEvent("timed out waiting for results to aggregate")
			span.SetTag("timeout", true)
			return nil, errors.New("timed out waiting for workspace/xreferences")

		case refs := <-results:
			for _, ref := range refs {
				refURI, err := url.Parse(ref.Reference.URI)
				if err != nil {
					// Flag the span as an error for metrics purposes.
					ext.Error.Set(span, true)
					span.SetTag("err", errors.Wrap(err, "parsing workspace/xreferences Reference URI"))
					continue aggregation
				}

				allRefs.Locations = append(allRefs.Locations, &sourcegraph.RefLocation{
					Scheme:    refURI.Scheme,
					Host:      refURI.Host,
					Path:      refURI.Path,
					Version:   refURI.RawQuery,
					File:      refURI.Fragment,
					StartLine: ref.Reference.Range.Start.Line,
					StartChar: ref.Reference.Range.Start.Character,
					EndLine:   ref.Reference.Range.End.Line,
					EndChar:   ref.Reference.Range.End.Character,
				})
			}
			if len(allRefs.Locations) > minWorkspaceRefs && time.Since(xreferencesStart) > idealMaxAggregationTime {
				span.LogEvent(fmt.Sprintf("%v refs meets minWorkspaceRefs=%v AND idealMaxAggregationTime=%v met; returning results early", len(allRefs.Locations), minWorkspaceRefs, idealMaxAggregationTime))
				span.SetTag("earlyExit", true)
				xreferencesCancel()
				break aggregation
			}
			span.LogEvent(fmt.Sprintf("got %v refs (new total: %v)", len(refs), len(allRefs.Locations)))
		}
	}

	// Consistently sort the results, where possible.
	sort.Sort(allRefs)

	for _, l := range allRefs.Locations {
		span.LogEvent(fmt.Sprintf("result %s://%s%s?%s#%s:%d:%d-%d:%d\n", l.Scheme, l.Host, l.Path, l.Version, l.File, l.StartLine, l.StartChar, l.EndLine, l.EndChar))
	}
	span.SetTag("referencesFound", len(allRefs.Locations))
	return allRefs, nil
}

// UnsafeRefreshIndex refreshes the global deps index for the specified repo@commit.
//
// SECURITY: It is the caller's responsibility to ensure the repository is NOT
// a private one.
func (s *defs) UnsafeRefreshIndex(ctx context.Context, op *sourcegraph.DefsRefreshIndexOp) (err error) {
	if Mocks.Defs.RefreshIndex != nil {
		return Mocks.Defs.RefreshIndex(ctx, op)
	}

	ctx, done := trace(ctx, "Defs", "RefreshIndex", op, &err)
	defer done()

	inv, err := Repos.GetInventory(ctx, &sourcegraph.RepoRevSpec{op.RepoID, op.CommitID})
	if err != nil {
		return err
	}

	// Refresh global references indexes.
	return localstore.GlobalDeps.UnsafeRefreshIndex(ctx, op, inv.Languages)
}
