package context

import (
	"context"
	"math"
	"regexp"
	"sort"
	"sync"

	"github.com/sourcegraph/conc/pool"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/embed"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/client"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/settings"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type FileChunkContext struct {
	RepoName  api.RepoName
	RepoID    api.RepoID
	CommitID  api.CommitID
	Path      string
	StartLine int
	EndLine   int
}

func NewContextClient(logger log.Logger, db edb.EnterpriseDB, embeddingsClient embeddings.Client, searchClient client.SearchClient) *ContextClient {
	return &ContextClient{
		logger:           logger,
		db:               db,
		embeddingsClient: embeddingsClient,
		searchClient:     searchClient,
	}
}

type ContextClient struct {
	logger           log.Logger
	db               edb.EnterpriseDB
	embeddingsClient embeddings.Client
	searchClient     client.SearchClient
}

type GetContextArgs struct {
	Repos            []types.RepoIDName
	Query            string
	CodeResultsCount int32
	TextResultsCount int32
}

func (c *ContextClient) GetContext(ctx context.Context, args GetContextArgs) ([]FileChunkContext, error) {
	embeddingRepos, keywordRepos, err := c.partitionRepos(ctx, args.Repos)
	if err != nil {
		return nil, err
	}

	// NOTE: We use a pretty simple heuristic for combining results from
	// embeddings and keyword search. We use the ratio of repos with embeddings
	// to decide how many results out of our limit should be reserved for
	// embeddings results. We can't easily compare the scores between embeddings
	// and keyword search.
	embeddingsResultRatio := float32(len(embeddingRepos)) / float32(len(args.Repos))

	embeddingsArgs := GetContextArgs{
		Repos:            embeddingRepos,
		Query:            args.Query,
		CodeResultsCount: int32(float32(args.CodeResultsCount) * embeddingsResultRatio),
		TextResultsCount: int32(float32(args.TextResultsCount) * embeddingsResultRatio),
	}
	keywordArgs := GetContextArgs{
		Repos: keywordRepos,
		Query: args.Query,
		// Assign the remaining result budget to keyword search
		CodeResultsCount: args.CodeResultsCount - embeddingsArgs.CodeResultsCount,
		TextResultsCount: args.TextResultsCount - embeddingsArgs.TextResultsCount,
	}

	// Fetch keyword results and embeddings results concurrently
	p := pool.NewWithResults[[]FileChunkContext]().WithErrors()
	p.Go(func() ([]FileChunkContext, error) {
		return c.getEmbeddingsContext(ctx, embeddingsArgs)
	})
	p.Go(func() ([]FileChunkContext, error) {
		return c.getKeywordContext(ctx, keywordArgs)
	})

	results, err := p.Wait()
	if err != nil {
		return nil, err
	}

	return append(results[0], results[1]...), nil
}

// partitionRepos splits a set of repos into repos with embeddings and repos without embeddings
func (c *ContextClient) partitionRepos(ctx context.Context, input []types.RepoIDName) (embedded, notEmbedded []types.RepoIDName, err error) {
	for _, repo := range input {
		exists, err := c.db.Repos().RepoEmbeddingExists(ctx, repo.ID)
		if err != nil {
			return nil, nil, err
		}

		if exists {
			embedded = append(embedded, repo)
		} else {
			notEmbedded = append(notEmbedded, repo)
		}
	}
	return embedded, notEmbedded, nil
}

func (c *ContextClient) getEmbeddingsContext(ctx context.Context, args GetContextArgs) (_ []FileChunkContext, err error) {
	repoNames := make([]api.RepoName, len(args.Repos))
	repoIDs := make([]api.RepoID, len(args.Repos))
	for i, repo := range args.Repos {
		repoNames[i] = repo.Name
		repoIDs[i] = repo.ID
	}

	results, err := c.embeddingsClient.Search(ctx, embeddings.EmbeddingsSearchParameters{
		RepoNames:        repoNames,
		RepoIDs:          repoIDs,
		Query:            args.Query,
		CodeResultsCount: int(args.CodeResultsCount),
		TextResultsCount: int(args.TextResultsCount),
	})
	if err != nil {
		return nil, err
	}

	idsByName := make(map[api.RepoName]api.RepoID)
	for i, repoName := range repoNames {
		idsByName[repoName] = repoIDs[i]
	}

	res := make([]FileChunkContext, 0, len(results.CodeResults)+len(results.TextResults))
	for _, result := range append(results.CodeResults, results.TextResults...) {
		res = append(res, FileChunkContext{
			RepoName:  result.RepoName,
			RepoID:    idsByName[result.RepoName],
			CommitID:  result.Revision,
			Path:      result.FileName,
			StartLine: result.StartLine,
			EndLine:   result.EndLine,
		})
	}
	return res, nil
}

// getKeywordContext uses keyword search to find relevant bits of context for Cody
func (c *ContextClient) getKeywordContext(ctx context.Context, args GetContextArgs) (_ []FileChunkContext, err error) {
	settings, err := settings.CurrentUserFinal(ctx, c.db)
	if err != nil {
		return nil, err
	}

	// mini-HACK: pass in the scope using repo: filters. In an ideal world, we
	// would not be using query text manipulation for this and would be using
	// the job structs directly.
	regexEscapedRepoNames := make([]string, len(args.Repos))
	for i, repo := range args.Repos {
		regexEscapedRepoNames[i] = regexp.QuoteMeta(string(repo.Name))
	}
	query := "repo:^" + query.UnionRegExps(regexEscapedRepoNames) + "$ " + args.Query

	patternTypeKeyword := "keyword"
	plan, err := c.searchClient.Plan(
		ctx,
		"V3",
		&patternTypeKeyword,
		query,
		search.Precise,
		search.Streaming,
		settings,
		envvar.SourcegraphDotComMode(),
	)
	if err != nil {
		return nil, err
	}

	// Create a cancellable context to exit early once we have enough results
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		mux         sync.Mutex
		textResults []FileChunkContext
		codeResults []FileChunkContext
	)
	collectFileMatch := func(fm *result.FileMatch) {
		mux.Lock()
		defer mux.Unlock()

		if embed.IsValidTextFile(fm.Path) {
			if len(textResults) < int(args.TextResultsCount) {
				textResults = append(textResults, fileMatchToContextMatches(fm)...)
				textResults = truncate(textResults, int(args.TextResultsCount))
			}
		} else {
			if len(codeResults) < int(args.CodeResultsCount) {
				codeResults = append(codeResults, fileMatchToContextMatches(fm)...)
				codeResults = truncate(codeResults, int(args.CodeResultsCount))
			}
		}

		if len(textResults) >= int(args.TextResultsCount) && len(codeResults) >= int(args.CodeResultsCount) {
			cancel()
		}
	}

	s := streaming.StreamFunc(func(e streaming.SearchEvent) {
		for _, res := range e.Results {
			fm, ok := res.(*result.FileMatch)
			if !ok {
				continue
			}

			collectFileMatch(fm)
		}
	})

	alert, err := c.searchClient.Execute(ctx, s, plan)
	if err != nil {
		return nil, err
	}
	if alert != nil {
		c.logger.Warn("received alert from keyword search execution",
			log.String("title", alert.Title),
			log.String("description", alert.Description),
		)
	}

	return append(codeResults, textResults...), nil
}

func fileMatchToContextMatches(fm *result.FileMatch) []FileChunkContext {
	var relevantLines []int
	for _, chunk := range fm.ChunkMatches {
		for _, chunkRange := range chunk.Ranges {
			// Assume that all matches are just a single line.
			// We add context in both directions, so this is
			// probably a pretty decent assumption.
			relevantLines = append(relevantLines, chunkRange.Start.Line)
		}
	}

	sort.Ints(relevantLines)

	// Some quick numbers suggested that embeddings chunks are roughly 5-10
	// lines per chunk. 4 lines in either direction gives us an 8-line chunk.
	const expandLines = 4
	var res []FileChunkContext
	lastLine := -1
	for _, line := range relevantLines {
		// Skip any lines that have already been included
		if line <= lastLine {
			continue
		}

		res = append(res, FileChunkContext{
			RepoName: fm.Repo.Name,
			RepoID:   fm.Repo.ID,
			CommitID: fm.CommitID,
			Path:     fm.Path,
			// Add three lines of context before and after
			StartLine: max(0, lastLine, line-expandLines),
			EndLine:   line + expandLines,
		})

		lastLine = line + expandLines
	}

	return res
}

func max(vals ...int) int {
	res := math.MinInt32
	for _, val := range vals {
		if val > res {
			res = val
		}
	}
	return res
}

func min(vals ...int) int {
	res := math.MaxInt32
	for _, val := range vals {
		if val < res {
			res = val
		}
	}
	return res
}

func truncate[T any](input []T, size int) []T {
	return input[:min(len(input), size)]
}
