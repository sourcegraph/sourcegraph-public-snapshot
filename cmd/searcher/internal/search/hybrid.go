package search

import (
	"bytes"
	"context"
	"regexp/syntax"
	"sort"
	"time"
	"unicode/utf8"

	"github.com/grafana/regexp"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/zoekt"
	zoektquery "github.com/sourcegraph/zoekt/query"

	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
	"github.com/sourcegraph/sourcegraph/internal/api"
	zoektutil "github.com/sourcegraph/sourcegraph/internal/search/zoekt"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// TODO(keegancsmith) prometheus metrics

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
	logger := logWithTrace(ctx, s.Log).Scoped("hybrid", "experimental hybrid search").With(
		log.String("repo", string(p.Repo)),
		log.String("commit", string(p.Commit)),
		log.Int("endpoints", len(p.IndexerEndpoints)))

	client := getZoektClient(p.IndexerEndpoints)

	// There is a race condition between asking zoekt what is indexed vs
	// actually searching since the index may update. If the index changes,
	// which files we search need to change. As such we keep retrying until we
	// know we have had a consistent list and search on zoekt.
	for try := 0; try < 5; try++ {
		indexed, ok, err := zoektIndexedCommit(ctx, client, p.Repo)
		if err != nil {
			return nil, false, err
		}
		if !ok {
			logger.Warn("failed to find indexed commit")
			return nil, false, nil
		}

		// TODO if our store was more flexible we could cache just based on
		// indexed and p.Commit and avoid the need of running diff for each
		// search.
		out, err := s.GitDiffSymbols(ctx, p.Repo, indexed, p.Commit)
		if err != nil {
			return nil, false, err
		}

		indexedIgnore, unindexedSearch, err := parseGitDiffNameStatus(out)
		if err != nil {
			logger.Debug("parseGitDiffNameStatus failed",
				log.String("indexed", string(indexed)),
				log.Binary("out", out),
				log.Error(err))
			return nil, false, err
		}

		logger.Info("starting zoekt search",
			log.Int("try", try),
			log.String("indexed", string(indexed)),
			log.Int("indexedIgnorePaths", len(indexedIgnore)),
			log.Int("unindexedSearchPaths", len(unindexedSearch)))

		ok, err = zoektSearchIgnorePaths(ctx, client, p, sender, indexed, indexedIgnore)
		if err != nil {
			return nil, false, err
		} else if !ok {
			logger.Debug("retrying search since index changed while searching", log.String("indexed", string(indexed)))
			continue
		}

		return unindexedSearch, true, nil
	}

	logger.Warn("reached maximum try count, falling back to default unindexed search")
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

	k := zoektutil.ResultCountFactor(1, int32(p.Limit), false)
	opts := zoektutil.SearchOpts(ctx, k, int32(p.Limit), nil)
	if deadline, ok := ctx.Deadline(); ok {
		opts.MaxWallTime = time.Until(deadline) - 100*time.Millisecond
	}

	res, err := client.Search(ctx, q, &opts)
	if err != nil {
		return false, err
	}

	for _, fm := range res.Files {
		// Unexpected commit searched, signal to retry.
		if fm.Version != string(indexed) {
			return false, nil
		}

		cms := make([]protocol.ChunkMatch, 0, len(fm.ChunkMatches))
		for _, l := range fm.LineMatches {
			if l.FileName {
				continue
			}

			for _, m := range l.LineFragments {
				runeOffset := utf8.RuneCount(l.Line[:m.LineOffset])
				runeLength := utf8.RuneCount(l.Line[m.LineOffset : m.LineOffset+m.MatchLength])

				cms = append(cms, protocol.ChunkMatch{
					Content: string(l.Line),
					// zoekt line numbers are 1-based rather than 0-based so subtract 1
					ContentStart: protocol.Location{
						Offset: int32(l.LineStart),
						Line:   int32(l.LineNumber - 1),
						Column: 0,
					},
					Ranges: []protocol.Range{{
						Start: protocol.Location{
							Offset: int32(m.Offset),
							Line:   int32(l.LineNumber - 1),
							Column: int32(runeOffset),
						},
						End: protocol.Location{
							Offset: int32(m.Offset) + int32(m.MatchLength),
							Line:   int32(l.LineNumber - 1),
							Column: int32(runeOffset + runeLength),
						},
					}},
				})
			}
		}

		for _, cm := range fm.ChunkMatches {
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
	} else {
		re, err := syntax.Parse(rg.re.String(), syntax.Perl)
		if err != nil {
			return nil, err
		}
		parts = append(parts, &zoektquery.Regexp{
			Regexp:        re,
			Content:       true,
			CaseSensitive: !rg.ignoreCase,
		})
	}

	for _, pat := range p.IncludePatterns {
		if !p.PathPatternsAreRegExps {
			return nil, errors.New("hybrid search expects PathPatternsAreRegExps")
		}
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
		if !p.PathPatternsAreRegExps {
			return nil, errors.New("hybrid search expects PathPatternsAreRegExps")
		}
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

// logWithTrace is a helper which returns l.WithTrace if there is a
// TraceContext associated with ctx.
func logWithTrace(ctx context.Context, l log.Logger) log.Logger {
	return l.WithTrace(trace.Context(ctx))
}
