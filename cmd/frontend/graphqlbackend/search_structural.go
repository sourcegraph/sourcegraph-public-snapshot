package graphqlbackend

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"regexp/syntax"
	"strings"
	"time"
	"unicode/utf8"

	zoektquery "github.com/google/zoekt/query"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gituri"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/symbols/protocol"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

func splitOnHolesPattern() string {
	word := `\w+`
	whitespaceAndOptionalWord := `[ ]+(` + word + `)?`
	holeAnything := `:\[` + word + `\]`
	holeAlphanum := `:\[\[` + word + `\]\]`
	holeWithPunctuation := `:\[` + word + `\.\]`
	holeWithNewline := `:\[` + word + `\\n\]`
	holeWhitespace := `:\[` + whitespaceAndOptionalWord + `\]`
	return strings.Join([]string{
		holeAnything,
		holeAlphanum,
		holeWithPunctuation,
		holeWithNewline,
		holeWhitespace,
	}, "|")
}

var matchHoleRegexp = lazyregexp.New(splitOnHolesPattern())

// Converts comby a structural pattern to a Zoekt regular expression query. It
// converts whitespace in the pattern so that content across newlines can be
// matched in the index. As an incomplete approximation, we use the regex
// pattern .*? to scan ahead.
// Example:
// "ParseInt(:[args]) if err != nil" -> "ParseInt(.*)\s+if\s+err!=\s+nil"
func StructuralPatToRegexpQuery(pattern string) (zoektquery.Q, error) {
	substrings := matchHoleRegexp.Split(pattern, -1)
	var children []zoektquery.Q
	var pieces []string
	for _, s := range substrings {
		piece := regexp.QuoteMeta(s)
		onMatchWhitespace := lazyregexp.New(`[\s]+`)
		piece = onMatchWhitespace.ReplaceAllLiteralString(piece, `[\s]+`)
		pieces = append(pieces, piece)
	}

	if len(pieces) == 0 {
		return &zoektquery.Const{Value: true}, nil
	}
	rs := "(" + strings.Join(pieces, ")(.|\\s)*?(") + ")"
	re, _ := syntax.Parse(rs, syntax.ClassNL|syntax.PerlX|syntax.UnicodeGroups)
	children = append(children, &zoektquery.Regexp{
		Regexp:        re,
		CaseSensitive: true,
		Content:       true,
	})
	return &zoektquery.And{Children: children}, nil
}

func StructuralPatToQuery(pattern string) (zoektquery.Q, error) {
	regexpQuery, err := StructuralPatToRegexpQuery(pattern)
	if err != nil {
		return nil, err
	}
	return &zoektquery.Or{Children: []zoektquery.Q{regexpQuery}}, nil
}

func structuralQueryToZoektQuery(query *search.TextPatternInfo, isSymbol bool) (zoektquery.Q, error) {
	var and []zoektquery.Q

	var q zoektquery.Q
	var err error
	if query.IsRegExp {
		fileNameOnly := query.PatternMatchesPath && !query.PatternMatchesContent
		q, err = parseRe(query.Pattern, fileNameOnly, query.IsCaseSensitive)
		if err != nil {
			return nil, err
		}
	} else if query.IsStructuralPat {
		q, err = StructuralPatToQuery(query.Pattern)
		if err != nil {
			return nil, err
		}
	} else {
		q = &zoektquery.Substring{
			Pattern:       query.Pattern,
			CaseSensitive: query.IsCaseSensitive,

			FileName: true,
			Content:  true,
		}
	}

	if isSymbol {
		q = &zoektquery.Symbol{
			Expr: q,
		}
	}

	and = append(and, q)

	// zoekt also uses regular expressions for file paths
	// TODO PathPatternsAreCaseSensitive
	// TODO whitespace in file path patterns?
	if !query.PathPatternsAreRegExps {
		return nil, errors.New("zoekt only supports regex path patterns")
	}
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

	return zoektquery.Simplify(zoektquery.NewAnd(and...)), nil
}

// zoektSearchHEADOnlyFiles searches repositories using zoekt, returning only the file paths containing
// content matching the given pattern.
//
// Timeouts are reported through the context, and as a special case errNoResultsInTimeout
// is returned if no results are found in the given timeout (instead of the more common
// case of finding partial or full results in the given timeout).
func zoektSearchHEADOnlyFiles(ctx context.Context, args *search.TextParameters, repos []*search.RepositoryRevisions, isSymbol bool, since func(t time.Time) time.Duration) (fm []*FileMatchResolver, limitHit bool, reposLimitHit map[string]struct{}, err error) {
	if len(repos) == 0 {
		return nil, false, nil, nil
	}

	// Tell zoekt which repos to search
	repoSet := &zoektquery.RepoSet{Set: make(map[string]bool, len(repos))}
	repoMap := make(map[api.RepoName]*search.RepositoryRevisions, len(repos))
	for _, repoRev := range repos {
		repoSet.Set[string(repoRev.Repo.Name)] = true
		repoMap[api.RepoName(strings.ToLower(string(repoRev.Repo.Name)))] = repoRev
	}

	queryExceptRepos, err := queryToZoektQuery(args.PatternInfo, isSymbol)
	if err != nil {
		return nil, false, nil, err
	}
	finalQuery := zoektquery.NewAnd(repoSet, queryExceptRepos)

	tr, ctx := trace.New(ctx, "zoekt.Search", fmt.Sprintf("%d %+v", len(repoSet.Set), finalQuery.String()))
	defer func() {
		tr.SetError(err)
		if len(fm) > 0 {
			tr.LazyPrintf("%d file matches", len(fm))
		}
		tr.Finish()
	}()

	k := zoektResultCountFactor(len(repos), args.PatternInfo)
	searchOpts := zoektSearchOpts(k, args.PatternInfo)

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
		cNew, cancel := context.WithCancel(context.Background())
		go func(cOld context.Context) {
			<-cOld.Done()
			// cancel the new context if the old one is done for some reason other than the deadline passing.
			if cOld.Err() != context.DeadlineExceeded {
				cancel()
			}
		}(ctx)
		ctx = cNew
		defer cancel()
	}

	// If the query has a `repohasfile` or `-repohasfile` flag, we want to construct a new reposet based
	// on the values passed in to the flag.
	newRepoSet, err := createNewRepoSetWithRepoHasFileInputs(ctx, args.PatternInfo, args.Zoekt.Client, repoSet)
	if err != nil {
		return nil, false, nil, err
	}
	finalQuery = zoektquery.NewAnd(newRepoSet, queryExceptRepos)
	tr.LazyPrintf("after repohasfile filters: nRepos=%d query=%v", len(newRepoSet.Set), finalQuery)

	t0 := time.Now()
	resp, err := args.Zoekt.Client.Search(ctx, finalQuery, &searchOpts)
	if err != nil {
		return nil, false, nil, err
	}
	if resp.FileCount == 0 && resp.MatchCount == 0 && since(t0) >= searchOpts.MaxWallTime {
		return nil, false, nil, errNoResultsInTimeout
	}
	limitHit = resp.FilesSkipped+resp.ShardsSkipped > 0
	// Repositories that weren't fully evaluated because they hit the Zoekt or Sourcegraph file match limits.
	reposLimitHit = make(map[string]struct{})
	if limitHit {
		// Zoekt either did not evaluate some files in repositories, or ignored some repositories altogether.
		// In this case, we can't be sure that we have exhaustive results for _any_ repository. So, all file
		// matches are from repos with potentially skipped matches.
		for _, file := range resp.Files {
			if _, ok := reposLimitHit[file.Repository]; !ok {
				reposLimitHit[file.Repository] = struct{}{}
			}
		}
	}

	if len(resp.Files) == 0 {
		return nil, false, nil, nil
	}

	maxLineMatches := 25 + k
	maxLineFragmentMatches := 3 + k
	if limit := int(args.PatternInfo.FileMatchLimit); len(resp.Files) > limit {
		// List of files we cut out from the Zoekt response because they exceed the file match limit on the Sourcegraph end.
		// We use this to get a list of repositories that do not have complete results.
		fileMatchesInSkippedRepos := resp.Files[limit:]
		resp.Files = resp.Files[:limit]

		if !limitHit {
			// Zoekt evaluated all files and repositories, but Zoekt returned more file matches
			// than the limit we set on Sourcegraph, so we cut out more results.

			// Generate a list of repositories that had results cut because they exceeded the file match limit set on Sourcegraph.
			for _, file := range fileMatchesInSkippedRepos {
				if _, ok := reposLimitHit[file.Repository]; !ok {
					reposLimitHit[file.Repository] = struct{}{}
				}
			}
		}

		limitHit = true
	}

	matches := make([]*FileMatchResolver, len(resp.Files))
	for i, file := range resp.Files {
		fileLimitHit := false
		if len(file.LineMatches) > maxLineMatches {
			file.LineMatches = file.LineMatches[:maxLineMatches]
			fileLimitHit = true
			limitHit = true
		}
		repoRev := repoMap[api.RepoName(strings.ToLower(string(file.Repository)))]
		inputRev := repoRev.RevSpecs()[0]
		baseURI := &gituri.URI{URL: url.URL{Scheme: "git://", Host: string(repoRev.Repo.Name), RawQuery: "?" + url.QueryEscape(inputRev)}}
		lines := make([]*lineMatch, 0, len(file.LineMatches))
		symbols := []*searchSymbolResult{}
		for _, l := range file.LineMatches {
			if !l.FileName {
				if len(l.LineFragments) > maxLineFragmentMatches {
					l.LineFragments = l.LineFragments[:maxLineFragmentMatches]
				}
				offsets := make([][2]int32, len(l.LineFragments))
				for k, m := range l.LineFragments {
					offset := utf8.RuneCount(l.Line[:m.LineOffset])
					length := utf8.RuneCount(l.Line[m.LineOffset : m.LineOffset+m.MatchLength])
					offsets[k] = [2]int32{int32(offset), int32(length)}
					if isSymbol && m.SymbolInfo != nil {
						commit := &GitCommitResolver{
							repo:     &RepositoryResolver{repo: repoRev.Repo},
							oid:      GitObjectID(repoRev.IndexedHEADCommit()),
							inputRev: &inputRev,
						}

						symbols = append(symbols, &searchSymbolResult{
							symbol: protocol.Symbol{
								Name:       m.SymbolInfo.Sym,
								Kind:       m.SymbolInfo.Kind,
								Parent:     m.SymbolInfo.Parent,
								ParentKind: m.SymbolInfo.ParentKind,
								Path:       file.FileName,
								Line:       l.LineNumber,
							},
							lang:    strings.ToLower(file.Language),
							baseURI: baseURI,
							commit:  commit,
						})
					}
				}
				if !isSymbol {
					lines = append(lines, &lineMatch{
						JPreview:          string(l.Line),
						JLineNumber:       int32(l.LineNumber - 1),
						JOffsetAndLengths: offsets,
					})
				}
			}
		}
		matches[i] = &FileMatchResolver{
			JPath:        file.FileName,
			JLineMatches: lines,
			JLimitHit:    fileLimitHit,
			uri:          fileMatchURI(repoRev.Repo.Name, "", file.FileName),
			symbols:      symbols,
			Repo:         repoRev.Repo,
			CommitID:     repoRev.IndexedHEADCommit(),
		}
	}

	return matches, limitHit, reposLimitHit, nil
}
