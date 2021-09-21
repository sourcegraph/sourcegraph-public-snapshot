package commit

import (
	"bufio"
	"context"
	"regexp"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	gitprotocol "github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

func searchInReposNew(ctx context.Context, db dbutil.DB, textParams *search.TextParametersForCommitParameters, params searchCommitsInReposParameters) error {
	g, ctx := errgroup.WithContext(ctx)
	for _, repoRev := range textParams.Repos {
		// Skip the repo if no revisions were resolved for it
		if len(repoRev.Revs) == 0 {
			continue
		}

		rr := repoRev
		query := params.CommitParams.Query
		diff := params.CommitParams.Diff
		limit := int(textParams.PatternInfo.FileMatchLimit)

		args := &protocol.SearchRequest{
			Repo:        rr.Repo.Name,
			Revisions:   searchRevsToGitserverRevs(rr.Revs),
			Query:       &gitprotocol.And{queryNodesToPredicates(query, query.IsCaseSensitive(), diff)},
			IncludeDiff: diff,
			Limit:       limit,
		}

		onMatches := func(in []protocol.CommitMatch) {
			res := make([]result.Match, 0, len(in))
			for _, protocolMatch := range in {
				res = append(res, protocolMatchToCommitMatch(rr.Repo, diff, protocolMatch))
			}
			params.ResultChannel.Send(streaming.SearchEvent{
				Results: res,
			})
		}

		g.Go(func() error {
			limitHit, err := gitserver.DefaultClient.Search(ctx, args, onMatches)
			params.ResultChannel.Send(streaming.SearchEvent{
				Stats: streaming.Stats{
					IsLimitHit: limitHit,
				},
			})
			return err
		})
	}

	return g.Wait()
}

func searchRevsToGitserverRevs(in []search.RevisionSpecifier) []gitprotocol.RevisionSpecifier {
	out := make([]gitprotocol.RevisionSpecifier, 0, len(in))
	for _, rev := range in {
		out = append(out, gitprotocol.RevisionSpecifier{
			RevSpec:        rev.RevSpec,
			RefGlob:        rev.RefGlob,
			ExcludeRefGlob: rev.ExcludeRefGlob,
		})
	}
	return out
}

func queryNodesToPredicates(nodes []query.Node, caseSensitive, diff bool) []gitprotocol.SearchQuery {
	res := make([]gitprotocol.SearchQuery, 0, len(nodes))
	for _, node := range nodes {
		var newPred gitprotocol.SearchQuery
		switch v := node.(type) {
		case query.Operator:
			newPred = queryOperatorToPredicate(v, caseSensitive, diff)
		case query.Pattern:
			newPred = queryPatternToPredicate(v, caseSensitive, diff)
		case query.Parameter:
			newPred = queryParameterToPredicate(v, caseSensitive, diff)
		}
		if newPred != nil {
			res = append(res, newPred)
		}
	}
	return res
}

func queryOperatorToPredicate(op query.Operator, caseSensitive, diff bool) gitprotocol.SearchQuery {
	switch op.Kind {
	case query.And:
		return &gitprotocol.And{queryNodesToPredicates(op.Operands, caseSensitive, diff)}
	case query.Or:
		return &gitprotocol.Or{queryNodesToPredicates(op.Operands, caseSensitive, diff)}
	default:
		// I don't think we should have concats at this point, but ignore it if we do
		return nil
	}
}

func queryPatternToPredicate(pattern query.Pattern, caseSensitive, diff bool) gitprotocol.SearchQuery {
	patString := pattern.Value
	if pattern.Annotation.Labels.IsSet(query.Literal) {
		patString = regexp.QuoteMeta(pattern.Value)
	}

	var newPred gitprotocol.SearchQuery
	if diff {
		newPred = &gitprotocol.DiffMatches{gitprotocol.Regexp{regexp.MustCompile(wrapCaseSensitive(patString, caseSensitive))}}
	} else {
		newPred = &gitprotocol.MessageMatches{gitprotocol.Regexp{regexp.MustCompile(wrapCaseSensitive(patString, caseSensitive))}}
	}

	if pattern.Negated {
		return &gitprotocol.Not{newPred}
	}
	return newPred
}

func queryParameterToPredicate(parameter query.Parameter, caseSensitive, diff bool) gitprotocol.SearchQuery {
	var newPred gitprotocol.SearchQuery
	switch parameter.Field {
	case query.FieldAuthor:
		// TODO look up emails
		newPred = &gitprotocol.AuthorMatches{gitprotocol.Regexp{regexp.MustCompile(wrapCaseSensitive(parameter.Value, caseSensitive))}}
	case query.FieldCommitter:
		newPred = &gitprotocol.CommitterMatches{gitprotocol.Regexp{regexp.MustCompile(wrapCaseSensitive(parameter.Value, caseSensitive))}}
	case query.FieldBefore:
		newPred = &gitprotocol.CommitBefore{time.Now()} // TODO parse the time in with go-naturaldate
	case query.FieldAfter:
		newPred = &gitprotocol.CommitAfter{time.Now()}
	case query.FieldMessage:
		newPred = &gitprotocol.MessageMatches{gitprotocol.Regexp{regexp.MustCompile(wrapCaseSensitive(parameter.Value, caseSensitive))}}
	case query.FieldContent:
		if diff {
			newPred = &gitprotocol.DiffMatches{gitprotocol.Regexp{regexp.MustCompile(wrapCaseSensitive(parameter.Value, caseSensitive))}}
		} else {
			newPred = &gitprotocol.MessageMatches{gitprotocol.Regexp{regexp.MustCompile(wrapCaseSensitive(parameter.Value, caseSensitive))}}
		}
	case query.FieldFile:
		newPred = &gitprotocol.DiffModifiesFile{gitprotocol.Regexp{regexp.MustCompile(wrapCaseSensitive(parameter.Value, caseSensitive))}}
	case query.FieldLang:
		newPred = &gitprotocol.DiffModifiesFile{gitprotocol.Regexp{regexp.MustCompile(search.LangToFileRegexp(parameter.Value))}}
	}

	if parameter.Negated {
		return &gitprotocol.Not{newPred}
	}
	return newPred
}

func wrapCaseSensitive(pattern string, caseSensitive bool) string {
	if caseSensitive {
		return pattern
	}
	return "(?i:" + pattern + ")"
}

func protocolMatchToCommitMatch(repo types.RepoName, diff bool, in protocol.CommitMatch) *result.CommitMatch {
	var (
		matchBody       string
		matchHighlights []result.HighlightedRange
		diffPreview     *result.HighlightedString
	)

	if diff {
		matchBody = "```diff\n" + in.Diff.Content + "\n```"
		matchHighlights = searchRangesToHighlights(in.Diff.Content, in.Diff.Highlights.Add(gitprotocol.Location{Line: 1}))
		diffPreview = &result.HighlightedString{
			Value:      in.Message.Content,
			Highlights: searchRangesToHighlights(in.Diff.Content, in.Diff.Highlights),
		}
	} else {
		matchBody = "```COMMIT_EDITMSG\n" + in.Message.Content + "\n```"
		matchHighlights = searchRangesToHighlights(in.Message.Content, in.Message.Highlights.Add(gitprotocol.Location{Line: 1}))
	}

	return &result.CommitMatch{
		Commit: git.Commit{
			ID: in.Oid,
			Author: git.Signature{
				Name:  in.Author.Name,
				Email: in.Author.Email,
				Date:  in.Author.Date,
			},
			Committer: &git.Signature{
				Name:  in.Committer.Name,
				Email: in.Committer.Email,
				Date:  in.Committer.Date,
			},
			Message: git.Message(in.Message.Content),
			Parents: in.Parents,
		},
		Repo: repo,
		MessagePreview: &result.HighlightedString{
			Value:      in.Message.Content,
			Highlights: searchRangesToHighlights(in.Message.Content, in.Message.Highlights),
		},
		DiffPreview: diffPreview,
		Body: result.HighlightedString{
			Value:      matchBody,
			Highlights: matchHighlights,
		},
	}
}

func searchRangesToHighlights(s string, ranges []gitprotocol.Range) []result.HighlightedRange {
	res := make([]result.HighlightedRange, 0, len(ranges))
	for _, r := range ranges {
		res = append(res, searchRangeToHighlights(s, r)...)
	}
	return res
}

// searchRangeToHighlight converts a Range (which can cross multiple lines)
// into HighlightedRange, which is scoped to one line. In order to do this
// correctly, we need the string that is being highlighted in order to identify
// line-end boundaries within multi-line ranges.
// TODO(camdencheek): push the Range format up the stack so we can be smarter about multi-line highlights.
func searchRangeToHighlights(s string, r gitprotocol.Range) []result.HighlightedRange {
	var res []result.HighlightedRange

	// Use a scanner to handle \r?\n
	scanner := bufio.NewScanner(strings.NewReader(s[r.Start.Offset:r.End.Offset]))
	lineNum := r.Start.Line
	for scanner.Scan() {
		line := scanner.Text()

		character := 0
		if lineNum == r.Start.Line {
			character = r.Start.Column
		}

		length := len(line)
		if lineNum == r.End.Line {
			length = r.End.Column - character
		}

		if length > 0 {
			res = append(res, result.HighlightedRange{
				Line:      int32(lineNum),
				Character: int32(character),
				Length:    int32(length),
			})
		}

		lineNum++
	}

	return res
}
