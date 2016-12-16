package backend

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"sort"
	"sync"

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

// subSelectors is a map of language-specific data selectors. The input data is
// from the language's workspace/xdefinition request, and the output data
// should be something that can be matched (using jsonb containment operator)
// against the metadata output of the build server method workspace/xdependencies.
//
// TODO(slimsag): move to a plugin-based architecture. This will work for the
// first ten or so languages.
var subSelectors = map[string]func(lspext.SymbolDescriptor) map[string]interface{}{
	"go": func(symbol lspext.SymbolDescriptor) map[string]interface{} {
		return map[string]interface{}{
			"package": symbol["package"],
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

func (s *defs) DeprecatedTotalRefs(ctx context.Context, repoURI string) (res int, err error) {
	ctx, done := trace(ctx, "Defs", "DeprecatedTotalRefs", repoURI, &err)
	defer done()
	return localstore.DeprecatedGlobalRefs.DeprecatedTotalRefs(ctx, repoURI)
}

func (s *defs) TotalRefs(ctx context.Context, source string) (res int, err error) {
	ctx, done := trace(ctx, "Defs", "TotalRefs", source, &err)
	defer done()
	return localstore.GlobalRefs.TotalRefs(ctx, source)
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
	rootPath := vcs + "://" + repo.URI + "?" + repo.DefaultBranch

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
	// one we have a language server that uses it.
	location := locations[0]

	depRefs, err := localstore.GlobalRefs.RefLocations(ctx, localstore.RefLocationsOptions{
		Language: op.Language,
		DepData:  subSelector(location.Symbol),
	})
	if err != nil {
		return nil, err
	}

	// TODO(slimsag): handle pagination here when it is important to us
	var mu sync.RWMutex
	allRefs := &sourcegraph.RefLocations{}
	run := parallel.NewRun(8)
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
		go func() {
			defer run.Release()

			repo, err := Repos.Get(ctx, &sourcegraph.RepoSpec{ID: ref.RepoID})
			if err != nil {
				run.Error(errors.Wrap(err, "Repos.Get"))
				return
			}
			vcs := "git" // TODO: store VCS type in *sourcegraph.Repo object.
			span.LogEventWithPayload("xdependency", repo.URI)

			rootPath := vcs + "://" + repo.URI + "?" + repo.DefaultBranch
			var refs []lspext.ReferenceInformation
			err = xlang.UnsafeOneShotClientRequest(ctx, op.Language, rootPath, "workspace/xreferences", lspext.WorkspaceReferencesParams{Query: location.Symbol, Hints: ref.Hints}, &refs)
			if err != nil {
				run.Error(errors.Wrap(err, "LSP Call workspace/xreferences"))
				return
			}
			span.LogEventWithPayload("xreferences for "+repo.URI, len(refs))
			mu.Lock()
			defer mu.Unlock()
			for _, ref := range refs {
				refURI, err := url.Parse(ref.Reference.URI)
				if err != nil {
					run.Error(errors.Wrap(err, "parsing workspace/xreferences Reference URI"))
					return
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
		}()
	}
	if err := run.Wait(); err != nil {
		// Flag the span as an error for metrics purposes, but still return
		// results if we have any.
		ext.Error.Set(span, true)
		span.SetTag("err", err.Error())
		if len(allRefs.Locations) == 0 {
			return nil, err
		}
	}

	// Consistently sort the results, where possible.
	sort.Sort(allRefs)

	for _, l := range allRefs.Locations {
		span.LogEvent(fmt.Sprintf("result %s://%s%s?%s#%s:%d:%d-%d:%d\n", l.Scheme, l.Host, l.Path, l.Version, l.File, l.StartLine, l.StartChar, l.EndLine, l.EndChar))
	}
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

	// Refresh global references indexes.
	return localstore.GlobalRefs.UnsafeRefreshIndex(ctx, op)
}
