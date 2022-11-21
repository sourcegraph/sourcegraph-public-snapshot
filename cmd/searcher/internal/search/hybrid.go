package search

import (
	"bytes"
	"context"
	"regexp/syntax"
	"sort"
	"time"

	"github.com/grafana/regexp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/zoekt"
	zoektquery "github.com/sourcegraph/zoekt/query"

	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
	"github.com/sourcegraph/sourcegraph/internal/api"
	zoektutil "github.com/sourcegraph/sourcegraph/internal/search/zoekt"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	metricHybridIndexChanged = promauto.NewCounter(prometheus.CounterOpts{
		Name: "searcher_hybrid_index_changed_total",
		Help: "Total number of times the zoekt index changed while doing a hybrid search.",
	})
	metricHybridFinalState = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "searcher_hybrid_final_state_total",
		Help: "Total number of times a hybrid search ended in a specific state.",
	}, []string{"state"})
)

// hybrid search is an experimental feature which will search zoekt only for
// the paths that are the same for p.Commit. unsearched is the paths that
// searcher needs to search on p.Commit. If ok is false, then the zoekt search
// failed in a way where we should fallback to a normal unindexed search on
// the whole commit.
//
// This only interacts with zoekt so that we can leverage the normal searcher
// code paths for the unindexed parts. IE unsearched is expected to be used to
// fetch a zip via the store and then do a normal unindexed search.
func (s *Service) hybrid(ctx context.Context, p *protocol.Request, sender matchSender) (unsearched []string, ok bool, err error) {
	rootLogger := logWithTrace(ctx, s.Log).Scoped("hybrid", "experimental hybrid search").With(
		log.String("repo", string(p.Repo)),
		log.String("commit", string(p.Commit)),
	)

	client := s.Indexed

	// There is a race condition between asking zoekt what is indexed vs
	// actually searching since the index may update. If the index changes,
	// which files we search need to change. As such we keep retrying until we
	// know we have had a consistent list and search on zoekt.
	for try := 0; try < 5; try++ {
		logger := rootLogger.With(log.Int("try", try))

		indexed, ok, err := zoektIndexedCommit(ctx, client, p.Repo)
		if err != nil {
			recordHybridFinalState("zoekt-list-error")
			return nil, false, err
		}
		if !ok {
			logger.Debug("failed to find indexed commit")
			recordHybridFinalState("zoekt-list-missing")
			return nil, false, nil
		}
		logger = logger.With(log.String("indexed", string(indexed)))

		// TODO if our store was more flexible we could cache just based on
		// indexed and p.Commit and avoid the need of running diff for each
		// search.
		out, err := s.GitDiffSymbols(ctx, p.Repo, indexed, p.Commit)
		if err != nil {
			recordHybridFinalState("git-diff-error")
			return nil, false, err
		}

		indexedIgnore, unindexedSearch, err := parseGitDiffNameStatus(out)
		if err != nil {
			logger.Debug("parseGitDiffNameStatus failed",
				log.Binary("out", out),
				log.Error(err))
			recordHybridFinalState("git-diff-parse-error")
			return nil, false, err
		}

		totalLenIndexedIgnore := totalStringsLen(indexedIgnore)
		totalLenUnindexedSearch := totalStringsLen(unindexedSearch)

		logger = logger.With(
			log.Int("indexedIgnorePaths", len(indexedIgnore)),
			log.Int("totalLenIndexedIgnorePaths", totalLenIndexedIgnore),
			log.Int("unindexedSearchPaths", len(unindexedSearch)),
			log.Int("totalLenUnindexedSearchPaths", totalLenUnindexedSearch))

		if totalLenIndexedIgnore > s.MaxTotalPathsLength || totalLenUnindexedSearch > s.MaxTotalPathsLength {
			logger.Debug("not doing hybrid search due to changed file list exceeding MAX_TOTAL_PATHS_LENGTH",
				log.Int("MAX_TOTAL_PATHS_LENGTH", s.MaxTotalPathsLength))
			recordHybridFinalState("diff-too-large")
			return nil, false, nil
		}

		logger.Debug("starting zoekt search")

		ok, err = zoektSearchIgnorePaths(ctx, client, p, sender, indexed, indexedIgnore)
		if err != nil {
			// Check for error conditions related to the request rather than
			// zoekt misbehaving.
			switch ctx.Err() {
			case context.Canceled:
				// We swallow the error since we only cancel requests once we
				// have hit limits or the RPC request has gone away.
				recordHybridFinalState("search-canceled")
				return nil, true, nil
			case context.DeadlineExceeded:
				// We return the error because hitting a deadline should be
				// unexpected. We also don't need to run the normal searcher
				// path in this case.
				recordHybridFinalState("search-timeout")
				return nil, true, err
			default:
				recordHybridFinalState("zoekt-search-error")
				return nil, false, err
			}
		} else if !ok {
			metricHybridIndexChanged.Inc()
			logger.Debug("retrying search since index changed while searching")
			continue
		}

		recordHybridFinalState("success")
		return unindexedSearch, true, nil
	}

	rootLogger.Warn("reached maximum try count, falling back to default unindexed search")
	recordHybridFinalState("max-retrys")
	return nil, false, nil
}

// zoektSearchIgnorePaths will execute the search for p on zoekt and stream
// out results via sender. It will not search paths listed under ignoredPaths.
//
// If we did not search the correct commit or we don't know if we did, ok is
// false.
func zoektSearchIgnorePaths(ctx context.Context, client zoekt.Streamer, p *protocol.Request, sender matchSender, indexed api.CommitID, ignoredPaths []string) (ok bool, err error) {
	qText, err := zoektCompile(&p.PatternInfo)
	if err != nil {
		return false, errors.Wrap(err, "failed to compile query for zoekt")
	}
	q := zoektquery.Simplify(zoektquery.NewAnd(
		zoektquery.NewSingleBranchesRepos("HEAD", uint32(p.RepoID)),
		qText,
		zoektIgnorePaths(ignoredPaths),
	))

	opts := (&zoektutil.Options{
		NumRepos:       1,
		FileMatchLimit: int32(p.Limit),
	}).ToSearch(ctx)
	if deadline, ok := ctx.Deadline(); ok {
		opts.MaxWallTime = time.Until(deadline) - 100*time.Millisecond
	}

	// We only support chunk matches below.
	opts.ChunkMatches = true

	res, err := client.Search(ctx, q, opts)
	if err != nil {
		return false, err
	}

	for _, fm := range res.Files {
		// Unexpected commit searched, signal to retry.
		if fm.Version != string(indexed) {
			return false, nil
		}

		cms := make([]protocol.ChunkMatch, 0, len(fm.ChunkMatches))
		for _, cm := range fm.ChunkMatches {
			if cm.FileName {
				continue
			}

			ranges := make([]protocol.Range, 0, len(cm.Ranges))
			for _, r := range cm.Ranges {
				ranges = append(ranges, protocol.Range{
					Start: protocol.Location{
						Offset: int32(r.Start.ByteOffset),
						Line:   int32(r.Start.LineNumber - 1),
						Column: int32(r.Start.Column - 1),
					},
					End: protocol.Location{
						Offset: int32(r.End.ByteOffset),
						Line:   int32(r.End.LineNumber - 1),
						Column: int32(r.End.Column - 1),
					},
				})
			}

			cms = append(cms, protocol.ChunkMatch{
				Content: string(cm.Content),
				ContentStart: protocol.Location{
					Offset: int32(cm.ContentStart.ByteOffset),
					Line:   int32(cm.ContentStart.LineNumber) - 1,
					Column: int32(cm.ContentStart.Column) - 1,
				},
				Ranges: ranges,
			})
		}

		sender.Send(protocol.FileMatch{
			Path:         fm.FileName,
			ChunkMatches: cms,
		})
	}

	// we have no matches, so we don't know if we searched the correct commit.
	if len(res.Files) == 0 {
		newIndexed, ok, err := zoektIndexedCommit(ctx, client, p.Repo)
		if err != nil {
			return false, errors.Wrap(err, "failed to double check indexed commit")
		}
		if !ok {
			// let the retry logic handle the call to zoektIndexedCommit again
			return false, nil
		}
		retry := newIndexed != indexed
		return !retry, nil
	}

	return true, nil
}

// zoektCompile builds a text search zoekt query for p.
//
// This function should support the same features as the "compile" function,
// but return a zoektquery instead of a readerGrep.
//
// Note: This is used by hybrid search and not structural search.
func zoektCompile(p *protocol.PatternInfo) (zoektquery.Q, error) {
	var parts []zoektquery.Q
	// we are redoing work here, but ensures we generate the same regex and it
	// feels nicer than passing in a readerGrep since handle path directly.
	if rg, err := compile(p); err != nil {
		return nil, err
	} else if rg.re == nil { // we are just matching paths
		parts = append(parts, &zoektquery.Const{Value: true})
	} else {
		re, err := syntax.Parse(rg.re.String(), syntax.Perl)
		if err != nil {
			return nil, err
		}
		re = zoektquery.OptimizeRegexp(re, syntax.Perl)
		if p.PatternMatchesContent && p.PatternMatchesPath {
			parts = append(parts, zoektquery.NewOr(
				&zoektquery.Regexp{
					Regexp:        re,
					Content:       true,
					CaseSensitive: !rg.ignoreCase,
				},
				&zoektquery.Regexp{
					Regexp:        re,
					FileName:      true,
					CaseSensitive: !rg.ignoreCase,
				},
			))
		} else {
			parts = append(parts, &zoektquery.Regexp{
				Regexp:        re,
				Content:       p.PatternMatchesContent,
				FileName:      p.PatternMatchesPath,
				CaseSensitive: !rg.ignoreCase,
			})
		}
	}

	for _, pat := range p.IncludePatterns {
		re, err := syntax.Parse(pat, syntax.Perl)
		if err != nil {
			return nil, err
		}
		parts = append(parts, &zoektquery.Regexp{
			Regexp:        re,
			FileName:      true,
			CaseSensitive: p.PathPatternsAreCaseSensitive,
		})
	}

	if p.ExcludePattern != "" {
		re, err := syntax.Parse(p.ExcludePattern, syntax.Perl)
		if err != nil {
			return nil, err
		}
		parts = append(parts, &zoektquery.Not{Child: &zoektquery.Regexp{
			Regexp:        re,
			FileName:      true,
			CaseSensitive: p.PathPatternsAreCaseSensitive,
		}})
	}

	return zoektquery.Simplify(zoektquery.NewAnd(parts...)), nil
}

func zoektIgnorePaths(paths []string) zoektquery.Q {
	if len(paths) == 0 {
		return &zoektquery.Const{Value: true}
	}

	parts := make([]zoektquery.Q, 0, len(paths))
	for _, p := range paths {
		re, err := syntax.Parse("^"+regexp.QuoteMeta(p)+"$", syntax.Perl)
		if err != nil {
			panic("failed to regex compile escaped literal: " + err.Error())
		}
		parts = append(parts, &zoektquery.Regexp{
			Regexp:        re,
			FileName:      true,
			CaseSensitive: true,
		})
	}

	return &zoektquery.Not{Child: zoektquery.NewOr(parts...)}
}

// zoektIndexedCommit returns the default indexed commit for a repository.
func zoektIndexedCommit(ctx context.Context, client zoekt.Streamer, repo api.RepoName) (api.CommitID, bool, error) {
	// TODO check we are using the most efficient way to List. I tested with
	// NewSingleBranchesRepos and it went through a slow path.
	q := zoektquery.NewRepoSet(string(repo))

	resp, err := client.List(ctx, q, &zoekt.ListOptions{Minimal: true})
	if err != nil {
		return "", false, err
	}

	for _, v := range resp.Minimal {
		return api.CommitID(v.Branches[0].Version), true, nil
	}

	return "", false, nil
}

// parseGitDiffNameStatus returns the paths changedA and changedB for commits
// A and B respectively. It expects to be parsing the output of the command
// git diff -z --name-status --no-renames A B.
func parseGitDiffNameStatus(out []byte) (changedA, changedB []string, err error) {
	if len(out) == 0 {
		return nil, nil, nil
	}

	slices := bytes.Split(bytes.TrimRight(out, "\x00"), []byte{0})
	if len(slices)%2 != 0 {
		return nil, nil, errors.New("uneven pairs")
	}

	for i := 0; i < len(slices); i += 2 {
		path := string(slices[i+1])
		switch slices[i][0] {
		case 'D': // no longer appears in B
			changedA = append(changedA, path)
		case 'M':
			changedA = append(changedA, path)
			changedB = append(changedB, path)
		case 'A': // doesn't exist in A
			changedB = append(changedB, path)
		}
	}
	sort.Strings(changedA)
	sort.Strings(changedB)

	return changedA, changedB, nil
}

func totalStringsLen(ss []string) int {
	sum := 0
	for _, s := range ss {
		sum += len(s)
	}
	return sum
}

// logWithTrace is a helper which returns l.WithTrace if there is a
// TraceContext associated with ctx.
func logWithTrace(ctx context.Context, l log.Logger) log.Logger {
	return l.WithTrace(trace.Context(ctx))
}

// recordHybridFinalState is a wrapper around metricHybridState to make the
// callsites more succinct.
func recordHybridFinalState(state string) {
	metricHybridFinalState.WithLabelValues(state).Inc()
}
