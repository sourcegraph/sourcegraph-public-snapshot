package graphql

import (
	"context"
	"fmt"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"strings"
	"sync"

	orderedmap "github.com/wk8/go-ordered-map/v2"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	sharedresolvers "github.com/sourcegraph/sourcegraph/internal/codeintel/shared/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/resolvers/gitresolvers"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	uploadsgraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/transport/graphql"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	searchclient "github.com/sourcegraph/sourcegraph/internal/search/client"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type rootResolver struct {
	svc                            CodeNavService
	autoindexingSvc                AutoIndexingService
	gitserverClient                gitserver.Client
	siteAdminChecker               sharedresolvers.SiteAdminChecker
	repoStore                      database.RepoStore
	searcherClient                 searchclient.SearchClient
	uploadLoaderFactory            uploadsgraphql.UploadLoaderFactory
	indexLoaderFactory             uploadsgraphql.IndexLoaderFactory
	locationResolverFactory        *gitresolvers.CachedLocationResolverFactory
	hunkCache                      codenav.HunkCache
	indexResolverFactory           *uploadsgraphql.PreciseIndexResolverFactory
	maximumIndexesPerMonikerSearch int
	operations                     *operations
}

func NewRootResolver(
	observationCtx *observation.Context,
	svc CodeNavService,
	autoindexingSvc AutoIndexingService,
	gitserverClient gitserver.Client,
	siteAdminChecker sharedresolvers.SiteAdminChecker,
	repoStore database.RepoStore,
	searcherClient searchclient.SearchClient,
	uploadLoaderFactory uploadsgraphql.UploadLoaderFactory,
	indexLoaderFactory uploadsgraphql.IndexLoaderFactory,
	indexResolverFactory *uploadsgraphql.PreciseIndexResolverFactory,
	locationResolverFactory *gitresolvers.CachedLocationResolverFactory,
	maxIndexSearch int,
	hunkCacheSize int,
) (resolverstubs.CodeNavServiceResolver, error) {
	hunkCache, err := codenav.NewHunkCache(hunkCacheSize)
	if err != nil {
		return nil, err
	}

	return &rootResolver{
		svc:                            svc,
		autoindexingSvc:                autoindexingSvc,
		gitserverClient:                gitserverClient,
		siteAdminChecker:               siteAdminChecker,
		repoStore:                      repoStore,
		searcherClient:                 searcherClient,
		uploadLoaderFactory:            uploadLoaderFactory,
		indexLoaderFactory:             indexLoaderFactory,
		indexResolverFactory:           indexResolverFactory,
		locationResolverFactory:        locationResolverFactory,
		hunkCache:                      hunkCache,
		maximumIndexesPerMonikerSearch: maxIndexSearch,
		operations:                     newOperations(observationCtx),
	}, nil
}

// 🚨 SECURITY: dbstore layer handles authz for query resolution
func (r *rootResolver) GitBlobLSIFData(ctx context.Context, args *resolverstubs.GitBlobLSIFDataArgs) (_ resolverstubs.GitBlobLSIFDataResolver, err error) {
	opts := args.Options()
	ctx, _, endObservation := r.operations.gitBlobLsifData.WithErrors(ctx, &err, observation.Args{Attrs: opts.Attrs()})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	uploads, err := r.svc.GetClosestCompletedUploadsForBlob(ctx, opts)
	if err != nil || len(uploads) == 0 {
		return nil, err
	}

	if len(uploads) == 0 {
		// If we're on sourcegraph.com and it's a rust package repo, index it on-demand
		if dotcom.SourcegraphDotComMode() && strings.HasPrefix(string(args.Repo.Name), "crates/") {
			err = r.autoindexingSvc.QueueRepoRev(ctx, int(args.Repo.ID), string(args.Commit))
		}

		return nil, err
	}

	reqState := codenav.NewRequestState(
		uploads,
		r.repoStore,
		authz.DefaultSubRepoPermsChecker,
		r.gitserverClient,
		args.Repo,
		string(args.Commit),
		args.Path,
		r.maximumIndexesPerMonikerSearch,
		r.hunkCache,
	)

	return newGitBlobLSIFDataResolver(
		r.svc,
		r.indexResolverFactory,
		reqState,
		r.uploadLoaderFactory.Create(),
		r.indexLoaderFactory.Create(),
		r.locationResolverFactory.Create(),
		r.operations,
	), nil
}

func (r *rootResolver) CodeGraphData(ctx context.Context, opts *resolverstubs.CodeGraphDataOpts) (_ *[]resolverstubs.CodeGraphDataResolver, err error) {
	ctx, _, endObservation := r.operations.codeGraphData.WithErrors(ctx, &err, observation.Args{Attrs: opts.Attrs()})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	makeResolvers := func(prov resolverstubs.CodeGraphDataProvenance) ([]resolverstubs.CodeGraphDataResolver, error) {
		indexer := ""
		if prov == resolverstubs.ProvenanceSyntactic {
			indexer = shared.SyntacticIndexer
		}
		uploads, err := r.svc.GetClosestCompletedUploadsForBlob(ctx, shared.UploadMatchingOptions{
			RepositoryID:       int(opts.Repo.ID),
			Commit:             string(opts.Commit),
			Path:               opts.Path,
			RootToPathMatching: shared.RootMustEnclosePath,
			Indexer:            indexer,
		})
		if err != nil || len(uploads) == 0 {
			return nil, err
		}
		resolvers := []resolverstubs.CodeGraphDataResolver{}
		for _, upload := range preferUploadsWithLongestRoots(uploads) {
			resolvers = append(resolvers, newCodeGraphDataResolver(r.svc, upload, opts, prov, r.operations))
		}
		return resolvers, nil
	}

	provs := opts.Args.ProvenancesForSCIPData()
	if provs.Precise {
		preciseResolvers, err := makeResolvers(resolverstubs.ProvenancePrecise)
		if len(preciseResolvers) != 0 || err != nil {
			return &preciseResolvers, err
		}
	}

	if provs.Syntactic {
		syntacticResolvers, err := makeResolvers(resolverstubs.ProvenanceSyntactic)
		if len(syntacticResolvers) != 0 || err != nil {
			return &syntacticResolvers, err
		}

		// Enhancement idea: if a syntactic SCIP index is unavailable,
		// but the language is supported by scip-syntax, we could generate
		// a syntactic SCIP index on-the-fly by having the syntax-highlighter
		// analyze the file.
	}

	// We do not currently have any way of generating SCIP data
	// during purely textual means.

	return &[]resolverstubs.CodeGraphDataResolver{}, nil
}

func preferUploadsWithLongestRoots(uploads []shared.CompletedUpload) []shared.CompletedUpload {
	// Use orderedmap instead of a map to preserve the order of the uploads
	// and to avoid introducing non-determinism.
	sortedMap := orderedmap.New[string, shared.CompletedUpload]()
	for _, upload := range uploads {
		key := fmt.Sprintf("%s:%s", upload.Indexer, upload.Commit)
		if val, found := sortedMap.Get(key); found {
			if len(val.Root) < len(upload.Root) {
				sortedMap.Set(key, upload)
			}
		} else {
			sortedMap.Set(key, upload)
		}
	}
	out := make([]shared.CompletedUpload, 0, sortedMap.Len())
	for pair := sortedMap.Oldest(); pair != nil; pair = pair.Next() {
		out = append(out, pair.Value)
	}
	return out
}

type matchingOccurrence struct {
	occurrence  *scip.Occurrence
	displayName string
}

func (r *rootResolver) UsagesForSymbol(ctx context.Context, unresolvedArgs *resolverstubs.UsagesForSymbolArgs) (_ resolverstubs.UsageConnectionResolver, err error) {
	ctx, _, endObservation := r.operations.usagesForSymbol.WithErrors(ctx, &err, observation.Args{Attrs: unresolvedArgs.Attrs()})
	numPreciseResults := 0
	numSyntacticResults := 0
	numSearchBasedResults := 0
	defer func() {
		endObservation.OnCancel(ctx, 1, observation.Args{Attrs: []attribute.KeyValue{
			attribute.Int("results.precise", numPreciseResults),
			attribute.Int("results.syntactic", numSyntacticResults),
			attribute.Int("results.searchBased", numSearchBasedResults),
		}})
	}()

	const maxUsagesCount = 100
	args, err := unresolvedArgs.Resolve(ctx, r.repoStore, r.gitserverClient, maxUsagesCount)
	if err != nil {
		return nil, err
	}
	remainingCount := int(*unresolvedArgs.First)
	provsForSCIPData := args.Symbol.ProvenancesForSCIPData()

	if provsForSCIPData.Precise {
		// Attempt to get up to remainingCount precise results.
		remainingCount = remainingCount - numPreciseResults
	}

	if remainingCount > 0 && provsForSCIPData.Syntactic {
		// Attempt to get up to remainingCount syntactic results.

		// Find matching uploads for the coordinates in the request
		repo, err := r.repoStore.GetByName(ctx, api.RepoName(args.Range.Repository))
		if err != nil {
			return nil, err
		}

		// TODO: Resolve revision?
		revision := "HEAD"
		if args.Range.Revision != nil {
			revision = *args.Range.Revision
		}

		uploads, err := r.svc.GetClosestCompletedUploadsForBlob(ctx, shared.UploadMatchingOptions{
			RepositoryID:       int(repo.ID),
			Commit:             revision,
			Path:               args.Range.Path,
			RootToPathMatching: shared.RootMustEnclosePath,
			Indexer:            shared.SyntacticIndexer,
		})

		if err != nil {
			return nil, err
		}

		if uploads == nil || len(uploads) == 0 {
			// TODO: probably just return 0 results instead?
			return nil, fmt.Errorf("no syntactic uploads found for repository %q", repo.Name)
		}

		if len(uploads) != 1 {
			// TODO: Is seeing multiple syntactic uploads an error?
		}

		syntacticUpload := uploads[0]

		doc, err := r.svc.SCIPDocument(ctx, syntacticUpload.ID, args.Range.Path)

		matchingOccurrences := make([]matchingOccurrence, 0)
		for _, occurrence := range doc.GetOccurrences() {
			occRange := occurrence.GetRange()
			var startLine, endLine, startCharacter, endCharacter int32
			if len(occRange) == 3 {
				startLine = occRange[0]
				startCharacter = occRange[1]
				endLine = occRange[0]
				endCharacter = occRange[2]
			} else {
				startLine = occRange[0]
				startCharacter = occRange[1]
				endLine = occRange[2]
				endCharacter = occRange[3]
			}

			// TODO: shouldn't need exact match, just overlap is enough
			if args.Range.Start.Line == startLine && args.Range.End.Line == endLine &&
				args.Range.Start.Character == startCharacter && args.Range.End.Character == endCharacter {

				for _, symbol := range doc.GetSymbols() {
					fmt.Printf("Symbol: %s: %s == %s\n", symbol.DisplayName, symbol.Symbol, occurrence.Symbol)
					if symbol.Symbol == occurrence.Symbol {
						matchingOccurrences = append(matchingOccurrences, matchingOccurrence{
							occurrence:  occurrence,
							displayName: symbol.DisplayName,
						})
					}
				}

			}
		}

		if len(matchingOccurrences) == 0 {
			return nil, fmt.Errorf("no matching occurrences found for range")
		}

		// TODO: Overlapping occurrences should lead to the same display name, but be scored separately.
		// (Meaning we just need a single Searcher/Zoekt search)
		matchingOccurrence := matchingOccurrences[0]

		fmt.Printf("Matching occurrence: %+v\n", matchingOccurrence)

		var contextLines int32 = 0
		patternType := "standard"
		repoName := fmt.Sprintf("^%s$", repo.Name)
		identifier := matchingOccurrence.displayName
		searchQuery := fmt.Sprintf("repo:%s rev:%s %s", repoName, revision, identifier)

		fmt.Printf("Sending: query=%s\n", searchQuery)

		plan, err := r.searcherClient.Plan(ctx, "V3", &patternType, searchQuery, search.Precise, 0, &contextLines)
		if err != nil {
			return nil, err
		}
		stream := streaming.NewAggregatingStream()
		_, err = r.searcherClient.Execute(ctx, stream, plan)
		if err != nil {
			return nil, err
		}
		for _, match := range stream.Results {
			t, ok := match.(*result.FileMatch)
			if !ok {
				continue
			}
			fmt.Printf("Matches in: %v/%s\n", match.RepoName().Name, match.Key().Path)
			for _, line := range t.ChunkMatches.AsLineMatches() {
				fmt.Printf("  %d:%d-%d:%d %s\n",
					line.LineNumber, line.OffsetAndLengths[0][0],
					line.LineNumber, line.OffsetAndLengths[0][0]+line.OffsetAndLengths[0][1],
					line.Preview,
				)
			}
		}
		remainingCount = remainingCount - numSyntacticResults
	}

	if remainingCount > 0 && provsForSCIPData.SearchBased {
		// Attempt to get up to remainingCount search-based results.
		_ = "shut up nogo linter complaining about empty branch"
	}

	return nil, errors.New("Not implemented yet")
}

// gitBlobLSIFDataResolver is the main interface to bundle-related operations exposed to the GraphQL API. This
// resolver concerns itself with GraphQL/API-specific behaviors (auth, validation, marshaling, etc.).
// All code intel-specific behavior is delegated to the underlying resolver instance, which is defined
// in the parent package.
type gitBlobLSIFDataResolver struct {
	codeNavSvc           CodeNavService
	indexResolverFactory *uploadsgraphql.PreciseIndexResolverFactory
	requestState         codenav.RequestState
	uploadLoader         uploadsgraphql.UploadLoader
	indexLoader          uploadsgraphql.IndexLoader
	locationResolver     *gitresolvers.CachedLocationResolver
	operations           *operations
}

// NewQueryResolver creates a new QueryResolver with the given resolver that defines all code intel-specific
// behavior. A cached location resolver instance is also given to the query resolver, which should be used
// to resolve all location-related values.
func newGitBlobLSIFDataResolver(
	codeNavSvc CodeNavService,
	indexResolverFactory *uploadsgraphql.PreciseIndexResolverFactory,
	requestState codenav.RequestState,
	uploadLoader uploadsgraphql.UploadLoader,
	indexLoader uploadsgraphql.IndexLoader,
	locationResolver *gitresolvers.CachedLocationResolver,
	operations *operations,
) resolverstubs.GitBlobLSIFDataResolver {
	return &gitBlobLSIFDataResolver{
		codeNavSvc:           codeNavSvc,
		uploadLoader:         uploadLoader,
		indexLoader:          indexLoader,
		indexResolverFactory: indexResolverFactory,
		requestState:         requestState,
		locationResolver:     locationResolver,
		operations:           operations,
	}
}

func (r *gitBlobLSIFDataResolver) ToGitTreeLSIFData() (resolverstubs.GitTreeLSIFDataResolver, bool) {
	return r, true
}

func (r *gitBlobLSIFDataResolver) ToGitBlobLSIFData() (resolverstubs.GitBlobLSIFDataResolver, bool) {
	return r, true
}

func (r *gitBlobLSIFDataResolver) VisibleIndexes(ctx context.Context) (_ *[]resolverstubs.PreciseIndexResolver, err error) {
	ctx, traceErrs, endObservation := r.operations.visibleIndexes.WithErrors(ctx, &err, observation.Args{Attrs: r.requestState.Attrs()})
	defer endObservation(1, observation.Args{})

	visibleUploads, err := r.codeNavSvc.VisibleUploadsForPath(ctx, r.requestState)
	if err != nil {
		return nil, err
	}

	resolvers := make([]resolverstubs.PreciseIndexResolver, 0, len(visibleUploads))
	for _, u := range visibleUploads {
		upload := u.ConvertToUpload()
		resolver, err := r.indexResolverFactory.Create(
			ctx,
			r.uploadLoader,
			r.indexLoader,
			r.locationResolver,
			traceErrs,
			&upload,
			nil,
		)
		if err != nil {
			return nil, err
		}
		resolvers = append(resolvers, resolver)
	}

	return &resolvers, nil
}

type codeGraphDataResolver struct {
	// Retrieved data/state
	retrievedDocument      sync.Once
	document               *scip.Document
	documentRetrievalError error

	// Arguments
	svc        CodeNavService
	upload     shared.CompletedUpload
	opts       *resolverstubs.CodeGraphDataOpts
	provenance resolverstubs.CodeGraphDataProvenance

	// O11y
	operations *operations
}

func newCodeGraphDataResolver(
	svc CodeNavService,
	upload shared.CompletedUpload,
	opts *resolverstubs.CodeGraphDataOpts,
	provenance resolverstubs.CodeGraphDataProvenance,
	operations *operations,
) resolverstubs.CodeGraphDataResolver {
	return &codeGraphDataResolver{
		sync.Once{},
		/*document*/ nil,
		/*documentRetrievalError*/ nil,
		svc,
		upload,
		opts,
		provenance,
		operations,
	}
}

func (c *codeGraphDataResolver) tryRetrieveDocument(ctx context.Context) (*scip.Document, error) {
	// NOTE(id: scip-doc-optimization): In the case of pagination, if we retrieve the document ID
	// from the database, we can avoid performing a JOIN between codeintel_scip_document_lookup
	// and codeintel_scip_documents
	c.retrievedDocument.Do(func() {
		c.document, c.documentRetrievalError = c.svc.SCIPDocument(ctx, c.upload.ID, c.opts.Path)
	})
	return c.document, c.documentRetrievalError
}

func (c *codeGraphDataResolver) Provenance(_ context.Context) (resolverstubs.CodeGraphDataProvenance, error) {
	return c.provenance, nil
}

func (c *codeGraphDataResolver) Commit(_ context.Context) (string, error) {
	return c.upload.Commit, nil
}

func (c *codeGraphDataResolver) ToolInfo(_ context.Context) (*resolverstubs.CodeGraphToolInfo, error) {
	return &resolverstubs.CodeGraphToolInfo{Name_: &c.upload.Indexer, Version_: &c.upload.IndexerVersion}, nil
}
