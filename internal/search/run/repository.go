package run

import (
	"context"
	"math"
	"strings"

	"github.com/cockroachdb/errors"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/sourcegraph/internal/api"
	searchrepos "github.com/sourcegraph/sourcegraph/internal/search/repos"
	"github.com/sourcegraph/sourcegraph/internal/search/unindexed"
	zoektutil "github.com/sourcegraph/sourcegraph/internal/search/zoekt"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

type RepoSearch struct {
	Args  *search.TextParameters
	Limit int

	IsRequired bool
}

func (s *RepoSearch) Run(ctx context.Context, stream streaming.Sender, repos searchrepos.Pager) (err error) {
	tr, ctx := trace.New(ctx, "RepoSearch", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	fieldAllowlist := map[string]struct{}{
		query.FieldRepo:               {},
		query.FieldContext:            {},
		query.FieldType:               {},
		query.FieldDefault:            {},
		query.FieldIndex:              {},
		query.FieldCount:              {},
		query.FieldTimeout:            {},
		query.FieldFork:               {},
		query.FieldArchived:           {},
		query.FieldVisibility:         {},
		query.FieldCase:               {},
		query.FieldRepoHasFile:        {},
		query.FieldRepoHasCommitAfter: {},
		query.FieldPatternType:        {},
		query.FieldSelect:             {},
	}
	// Don't return repo results if the search contains fields that aren't on the allowlist.
	// Matching repositories based whether they contain files at a certain path (etc.) is not yet implemented.
	for field := range s.Args.Query.Fields() {
		if _, ok := fieldAllowlist[field]; !ok {
			tr.LazyPrintf("contains dissallowed field: %s", field)
			return nil
		}
	}

	tr.LogFields(
		otlog.String("pattern", s.Args.PatternInfo.Pattern),
		otlog.Int("limit", s.Limit))

	opts := s.Args.RepoOptions // copy

	if s.Args.PatternInfo.Pattern != "" {
		opts.RepoFilters = append(make([]string, 0, len(opts.RepoFilters)), opts.RepoFilters...)
		opts.CaseSensitiveRepoFilters = s.Args.Query.IsCaseSensitive()

		patternPrefix := strings.SplitN(s.Args.PatternInfo.Pattern, "@", 2)
		if len(patternPrefix) == 0 {
			// No "@" in pattern? We're good.
			opts.RepoFilters = append(opts.RepoFilters, s.Args.PatternInfo.Pattern)
		} else if patternPrefix[0] != "" {
			// Extend the repo search using the pattern value, but
			// since the pattern contains @, only search the part
			// prefixed by the first @. This because downstream
			// logic will get confused by the presence of @ and try
			// to resolve repo revisions. See #27816.
			opts.RepoFilters = append(opts.RepoFilters, patternPrefix[0])
		} else {
			// This pattern starts with @, of the form "@thing". We can't
			// consistently handle search repos of this form, because
			// downstream logic will attempt to interpret "thing" as a repo
			// revision, may fail, and cause us to raise an alert for any
			// non `type:repo` search. Better to not attempt a repo search.
			return nil
		}
	}

	ctx, stream, cleanup := streaming.WithLimit(ctx, stream, s.Limit)
	defer cleanup()

	err = repos.Paginate(ctx, &opts, func(page *searchrepos.Resolved) error {
		tr.LogFields(otlog.Int("resolved.len", len(page.RepoRevs)))

		// Filter the repos if there is a repohasfile: or -repohasfile field.
		if len(s.Args.PatternInfo.FilePatternsReposMustExclude) > 0 || len(s.Args.PatternInfo.FilePatternsReposMustInclude) > 0 {
			// Fallback to batch for reposToAdd
			page.RepoRevs, err = reposToAdd(ctx, s.Args, page.RepoRevs)
			if err != nil {
				return err
			}
		}

		stream.Send(streaming.SearchEvent{
			Results: repoRevsToRepoMatches(ctx, page.RepoRevs),
		})

		return nil
	})

	if errors.Is(err, searchrepos.ErrNoResolvedRepos) {
		err = nil
	}

	return err
}

func (*RepoSearch) Name() string {
	return "Repo"
}

func (s *RepoSearch) Required() bool {
	return s.IsRequired
}

func repoRevsToRepoMatches(ctx context.Context, repos []*search.RepositoryRevisions) []result.Match {
	matches := make([]result.Match, 0, len(repos))
	for _, r := range repos {
		revs, err := r.ExpandedRevSpecs(ctx)
		if err != nil { // fallback to just return revspecs
			revs = r.RevSpecs()
		}
		for _, rev := range revs {
			matches = append(matches, &result.RepoMatch{
				Name: r.Repo.Name,
				ID:   r.Repo.ID,
				Rev:  rev,
			})
		}
	}
	return matches
}

func reposContainingPath(ctx context.Context, args *search.TextParameters, repos []*search.RepositoryRevisions, pattern string) ([]*result.FileMatch, error) {
	// Use a max FileMatchLimit to ensure we get all the repo matches we
	// can. Setting it to len(repos) could mean we miss some repos since
	// there could be for example len(repos) file matches in the first repo
	// and some more in other repos. deduplicate repo results
	p := search.TextPatternInfo{
		IsRegExp:                     true,
		FileMatchLimit:               math.MaxInt32,
		IncludePatterns:              []string{pattern},
		PathPatternsAreCaseSensitive: false,
		PatternMatchesContent:        true,
		PatternMatchesPath:           true,
	}
	q, err := query.ParseLiteral("file:" + pattern)
	if err != nil {
		return nil, err
	}
	newArgs := *args
	newArgs.PatternInfo = &p
	newArgs.Repos = repos
	newArgs.Query = q
	newArgs.UseFullDeadline = true

	globalSearch := newArgs.Mode == search.ZoektGlobalSearch
	zoektArgs, err := zoektutil.NewIndexedSearchRequest(ctx, &newArgs, globalSearch, search.TextRequest, func([]*search.RepositoryRevisions) {})
	if err != nil {
		return nil, err
	}
	searcherArgs := &search.SearcherParameters{
		SearcherURLs:    newArgs.SearcherURLs,
		PatternInfo:     newArgs.PatternInfo,
		UseFullDeadline: newArgs.UseFullDeadline,
	}
	matches, _, err := unindexed.SearchFilesInReposBatch(ctx, zoektArgs, searcherArgs, newArgs.Mode != search.SearcherOnly)
	if err != nil {
		return nil, err
	}
	return matches, err
}

// reposToAdd determines which repositories should be included in the result set based on whether they fit in the subset
// of repostiories specified in the query's `repohasfile` and `-repohasfile` fields if they exist.
func reposToAdd(ctx context.Context, args *search.TextParameters, repos []*search.RepositoryRevisions) ([]*search.RepositoryRevisions, error) {
	// matchCounts will contain the count of repohasfile patterns that matched.
	// For negations, we will explicitly set this to -1 if it matches.
	matchCounts := make(map[api.RepoID]int)
	if len(args.PatternInfo.FilePatternsReposMustInclude) > 0 {
		for _, pattern := range args.PatternInfo.FilePatternsReposMustInclude {
			matches, err := reposContainingPath(ctx, args, repos, pattern)
			if err != nil {
				return nil, err
			}

			matchedIDs := make(map[api.RepoID]struct{})
			for _, m := range matches {
				matchedIDs[m.Repo.ID] = struct{}{}
			}

			// increment the count for all seen repos
			for id := range matchedIDs {
				matchCounts[id] += 1
			}
		}
	} else {
		// Default to including all the repos, then excluding some of them below.
		for _, r := range repos {
			matchCounts[r.Repo.ID] = 0
		}
	}

	if len(args.PatternInfo.FilePatternsReposMustExclude) > 0 {
		for _, pattern := range args.PatternInfo.FilePatternsReposMustExclude {
			matches, err := reposContainingPath(ctx, args, repos, pattern)
			if err != nil {
				return nil, err
			}
			for _, m := range matches {
				matchCounts[m.Repo.ID] = -1
			}
		}
	}

	var rsta []*search.RepositoryRevisions
	for _, r := range repos {
		if count, ok := matchCounts[r.Repo.ID]; ok && count == len(args.PatternInfo.FilePatternsReposMustInclude) {
			rsta = append(rsta, r)
		}
	}

	return rsta, nil
}
