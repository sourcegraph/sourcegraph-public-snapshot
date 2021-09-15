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
	gitsearch "github.com/sourcegraph/sourcegraph/internal/gitserver/search"
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
			Predicate:   &gitsearch.And{queryNodesToPredicates(query, query.IsCaseSensitive(), diff)},
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

func queryNodesToPredicates(nodes []query.Node, caseSensitive, diff bool) []gitsearch.CommitPredicate {
	res := make([]gitsearch.CommitPredicate, 0, len(nodes))
	for _, node := range nodes {
		var newPred gitsearch.CommitPredicate
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

func queryOperatorToPredicate(op query.Operator, caseSensitive, diff bool) gitsearch.CommitPredicate {
	switch op.Kind {
	case query.And:
		return &gitsearch.And{queryNodesToPredicates(op.Operands, caseSensitive, diff)}
	case query.Or:
		return &gitsearch.Or{queryNodesToPredicates(op.Operands, caseSensitive, diff)}
	default:
		// I don't think we should have concats at this point, but ignore it if we do
		return nil
	}
}

func queryPatternToPredicate(pattern query.Pattern, caseSensitive, diff bool) gitsearch.CommitPredicate {
	patString := pattern.Value
	if pattern.Annotation.Labels.IsSet(query.Literal) {
		patString = regexp.QuoteMeta(pattern.Value)
	}

	var newPred gitsearch.CommitPredicate
	if diff {
		newPred = &gitsearch.DiffMatches{gitsearch.Regexp{regexp.MustCompile(wrapCaseSensitive(patString, caseSensitive))}}
	} else {
		newPred = &gitsearch.MessageMatches{gitsearch.Regexp{regexp.MustCompile(wrapCaseSensitive(patString, caseSensitive))}}
	}

	if pattern.Negated {
		return &gitsearch.Not{newPred}
	}
	return newPred
}

func queryParameterToPredicate(parameter query.Parameter, caseSensitive, diff bool) gitsearch.CommitPredicate {
	var newPred gitsearch.CommitPredicate
	switch parameter.Field {
	case query.FieldAuthor:
		// TODO look up emails
		newPred = &gitsearch.AuthorMatches{gitsearch.Regexp{regexp.MustCompile(wrapCaseSensitive(parameter.Value, caseSensitive))}}
	case query.FieldCommitter:
		newPred = &gitsearch.CommitterMatches{gitsearch.Regexp{regexp.MustCompile(wrapCaseSensitive(parameter.Value, caseSensitive))}}
	case query.FieldBefore:
		newPred = &gitsearch.CommitBefore{time.Now()} // TODO parse the time in with go-naturaldate
	case query.FieldAfter:
		newPred = &gitsearch.CommitAfter{time.Now()}
	case query.FieldMessage:
		newPred = &gitsearch.MessageMatches{gitsearch.Regexp{regexp.MustCompile(wrapCaseSensitive(parameter.Value, caseSensitive))}}
	case query.FieldContent:
		if diff {
			newPred = &gitsearch.DiffMatches{gitsearch.Regexp{regexp.MustCompile(wrapCaseSensitive(parameter.Value, caseSensitive))}}
		} else {
			newPred = &gitsearch.MessageMatches{gitsearch.Regexp{regexp.MustCompile(wrapCaseSensitive(parameter.Value, caseSensitive))}}
		}
	case query.FieldFile:
		newPred = &gitsearch.DiffModifiesFile{gitsearch.Regexp{regexp.MustCompile(wrapCaseSensitive(parameter.Value, caseSensitive))}}
	case query.FieldLang:
		// TODO(camdencheek)
		return nil
	}

	if parameter.Negated {
		return &gitsearch.Not{newPred}
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
		matchHighlights = searchRangesToHighlights(in.Diff.Content, in.Diff.Highlights.Shift(gitsearch.Location{Line: 1}))
		diffPreview = &result.HighlightedString{
			Value:      in.Message.Content,
			Highlights: searchRangesToHighlights(in.Diff.Content, in.Diff.Highlights),
		}
	} else {
		matchBody = "```COMMIT_EDITMSG\n" + in.Message.Content + "\n```"
		matchHighlights = searchRangesToHighlights(in.Message.Content, in.Message.Highlights.Shift(gitsearch.Location{Line: 1}))
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

func searchRangesToHighlights(s string, ranges []gitsearch.Range) []result.HighlightedRange {
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
func searchRangeToHighlights(s string, r gitsearch.Range) []result.HighlightedRange {
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
