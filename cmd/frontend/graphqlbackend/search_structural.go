package graphqlbackend

import (
	"context"
	"regexp/syntax"
	"time"

	"github.com/google/zoekt"
	zoektquery "github.com/google/zoekt/query"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/comby"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func HandleFilePathPatterns(query *search.TextPatternInfo) (zoektquery.Q, error) {
	var and []zoektquery.Q

	// Zoekt uses regular expressions for file paths.
	// Unhandled cases: PathPatternsAreCaseSensitive and whitespace in file path patterns.
	for _, p := range query.IncludePatterns {
		q, err := fileRe(p, query.IsCaseSensitive)
		if err != nil {
			return nil, err
		}
		and = append(and, q)
	}
	if query.ExcludePattern != "" {
		q, err := fileRe(query.ExcludePattern, query.IsCaseSensitive)
		if err != nil {
			return nil, err
		}
		and = append(and, &zoektquery.Not{Child: q})
	}

	// For conditionals that happen on a repo we can use type:repo queries. eg
	// (type:repo file:foo) (type:repo file:bar) will match all repos which
	// contain a filename matching "foo" and a filename matchinb "bar".
	//
	// Note: (type:repo file:foo file:bar) will only find repos with a
	// filename containing both "foo" and "bar".
	for _, p := range query.FilePatternsReposMustInclude {
		q, err := fileRe(p, query.IsCaseSensitive)
		if err != nil {
			return nil, err
		}
		and = append(and, &zoektquery.Type{Type: zoektquery.TypeRepo, Child: q})
	}
	for _, p := range query.FilePatternsReposMustExclude {
		q, err := fileRe(p, query.IsCaseSensitive)
		if err != nil {
			return nil, err
		}
		and = append(and, &zoektquery.Not{Child: &zoektquery.Type{Type: zoektquery.TypeRepo, Child: q}})
	}

	return zoektquery.NewAnd(and...), nil
}

func buildQuery(args *search.TextParameters, repos *indexedRepoRevs, filePathPatterns zoektquery.Q, shortcircuit bool) (zoektquery.Q, error) {
	regexString := comby.StructuralPatToRegexpQuery(args.PatternInfo.Pattern, shortcircuit)
	if len(regexString) == 0 {
		return &zoektquery.Const{Value: true}, nil
	}
	re, err := syntax.Parse(regexString, syntax.ClassNL|syntax.PerlX|syntax.UnicodeGroups)
	if err != nil {
		return nil, err
	}
	return zoektquery.NewAnd(
		&zoektquery.RepoBranches{Set: repos.repoBranches},
		filePathPatterns,
		&zoektquery.Regexp{
			Regexp:        re,
			CaseSensitive: true,
			Content:       true,
		},
	), nil
}

// zoektSearchHEADOnlyFiles searches repositories using zoekt, returning only the file paths containing
// content matching the given pattern.
//
// Timeouts are reported through the context, and as a special case errNoResultsInTimeout
// is returned if no results are found in the given timeout (instead of the more common
// case of finding partial or full results in the given timeout).
func zoektSearchHEADOnlyFiles(ctx context.Context, args *search.TextParameters, repos *indexedRepoRevs, since func(t time.Time) time.Duration) (fm []*FileMatchResolver, limitHit bool, partial map[api.RepoID]struct{}, err error) {
	if len(repos.repoRevs) == 0 {
		return nil, false, nil, nil
	}

	k := zoektResultCountFactor(len(repos.repoBranches), args.PatternInfo.FileMatchLimit, args.Mode == search.ZoektGlobalSearch)
	searchOpts := zoektSearchOpts(ctx, k, args.PatternInfo)

	if args.UseFullDeadline {
		// If the user manually specified a timeout, allow zoekt to use all of the remaining timeout.
		deadline, _ := ctx.Deadline()
		searchOpts.MaxWallTime = time.Until(deadline)

		// We don't want our context's deadline to cut off zoekt so that we can get the results
		// found before the deadline.
		//
		// We'll create a new context that gets cancelled if the other context is cancelled for any
		// reason other than the deadline being exceeded. This essentially means the deadline for the new context
		// will be `deadline + time for zoekt to cancel + network latency`.
		var cancel context.CancelFunc
		ctx, cancel = contextWithoutDeadline(ctx)
		defer cancel()
	}

	filePathPatterns, err := HandleFilePathPatterns(args.PatternInfo)
	if err != nil {
		return nil, false, nil, err
	}

	t0 := time.Now()
	q, err := buildQuery(args, repos, filePathPatterns, true)
	if err != nil {
		return nil, false, nil, err
	}
	resp, err := args.Zoekt.Client.Search(ctx, q, &searchOpts)
	if err != nil {
		return nil, false, nil, err
	}
	if since(t0) >= searchOpts.MaxWallTime {
		return nil, false, nil, errNoResultsInTimeout
	}

	// We always return approximate results (limitHit true) unless we run the branch to perform a more complete search.
	limitHit = true
	// If the previous indexed search did not return a substantial number of matching file candidates or count was
	// manually specified, run a more complete and expensive search.
	if resp.FileCount < 10 || args.PatternInfo.FileMatchLimit != defaultMaxSearchResults {
		q, err = buildQuery(args, repos, filePathPatterns, false)
		if err != nil {
			return nil, false, nil, err
		}
		resp, err = args.Zoekt.Client.Search(ctx, q, &searchOpts)
		if err != nil {
			return nil, false, nil, err
		}
		if since(t0) >= searchOpts.MaxWallTime {
			return nil, false, nil, errNoResultsInTimeout
		}
		// This is the only place limitHit can be set false, meaning we covered everything.
		limitHit = resp.FilesSkipped+resp.ShardsSkipped > 0
	}

	if len(resp.Files) == 0 {
		return nil, false, nil, nil
	}

	limitHit, files, partial := zoektLimitMatches(limitHit, int(args.PatternInfo.FileMatchLimit), resp.Files, func(file *zoekt.FileMatch) (repo *types.RepoName, revs []string, ok bool) {
		repo, inputRevs := repos.GetRepoInputRev(file)
		return repo, inputRevs, true
	})
	resp.Files = files

	maxLineMatches := 25 + k
	matches := make([]*FileMatchResolver, len(resp.Files))
	repoResolvers := make(RepositoryResolverCache)
	for i, file := range resp.Files {
		fileLimitHit := false
		if len(file.LineMatches) > maxLineMatches {
			file.LineMatches = file.LineMatches[:maxLineMatches]
			fileLimitHit = true
			limitHit = true
		}
		repoRev := repos.repoRevs[file.Repository]
		if repoResolvers[repoRev.Repo.Name] == nil {
			repoResolvers[repoRev.Repo.Name] = &RepositoryResolver{innerRepo: repoRev.Repo.ToRepo()}
		}
		matches[i] = &FileMatchResolver{
			JPath:     file.FileName,
			JLimitHit: fileLimitHit,
			uri:       fileMatchURI(repoRev.Repo.Name, "", file.FileName),
			Repo:      repoResolvers[repoRev.Repo.Name],
			CommitID:  api.CommitID(file.Version),
		}
	}

	return matches, limitHit, partial, nil
}
