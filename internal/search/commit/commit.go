package commit

import (
	"context"
	"strings"
	"time"

	"github.com/grafana/regexp"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	gitprotocol "github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/limits"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	searchrepos "github.com/sourcegraph/sourcegraph/internal/search/repos"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type CommitSearch struct {
	Query                gitprotocol.Node
	RepoOpts             search.RepoOptions
	Diff                 bool
	HasTimeFilter        bool
	Limit                int
	CodeMonitorID        *int64
	IncludeModifiedFiles bool

	// CodeMonitorSearchWrapper, if set, will wrap the commit search with extra logic specific to code monitors.
	CodeMonitorSearchWrapper CodeMonitorHook `json:"-"`
}

type DoSearchFunc func(*gitprotocol.SearchRequest) error
type CodeMonitorHook func(context.Context, database.DB, GitserverClient, *gitprotocol.SearchRequest, api.RepoID, DoSearchFunc) error

type GitserverClient interface {
	Search(_ context.Context, _ *protocol.SearchRequest, onMatches func([]protocol.CommitMatch)) (limitHit bool, _ error)
	ResolveRevisions(context.Context, api.RepoName, []gitprotocol.RevisionSpecifier) ([]string, error)
}

func (j *CommitSearch) Run(ctx context.Context, clients job.RuntimeClients, stream streaming.Sender) (alert *search.Alert, err error) {
	tr, ctx, stream, finish := job.StartSpan(ctx, stream, j)
	defer func() { finish(alert, err) }()
	tr.TagFields(trace.LazyFields(j.Tags))

	if err := j.ExpandUsernames(ctx, clients.DB); err != nil {
		return nil, err
	}

	opts := j.RepoOpts
	if opts.Limit == 0 {
		opts.Limit = reposLimit(j.HasTimeFilter)
	}

	resultType := "commit"
	if j.Diff {
		resultType = "diff"
	}

	var repoRevs []*search.RepositoryRevisions
	repos := searchrepos.Resolver{DB: clients.DB, Opts: opts}
	err = repos.Paginate(ctx, func(page *searchrepos.Resolved) error {
		if repoRevs = page.RepoRevs; page.Next != nil {
			return newReposLimitError(opts.Limit, j.HasTimeFilter, resultType)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	bounded := goroutine.NewBounded(8)
	for _, repoRev := range repoRevs {
		repoRev := repoRev // we close over repoRev in onMatches

		// Skip the repo if no revisions were resolved for it
		if len(repoRev.Revs) == 0 {
			continue
		}

		args := &protocol.SearchRequest{
			Repo:                 repoRev.Repo.Name,
			Revisions:            searchRevsToGitserverRevs(repoRev.Revs),
			Query:                j.Query,
			IncludeDiff:          j.Diff,
			Limit:                j.Limit,
			IncludeModifiedFiles: j.IncludeModifiedFiles,
		}

		onMatches := func(in []protocol.CommitMatch) {
			res := make([]result.Match, 0, len(in))
			for _, protocolMatch := range in {
				res = append(res, protocolMatchToCommitMatch(repoRev.Repo, j.Diff, protocolMatch))
			}
			stream.Send(streaming.SearchEvent{
				Results: res,
			})
		}

		doSearch := func(args *gitprotocol.SearchRequest) error {
			limitHit, err := clients.Gitserver.Search(ctx, args, onMatches)
			stream.Send(streaming.SearchEvent{
				Stats: streaming.Stats{
					IsLimitHit: limitHit,
				},
			})
			return err
		}

		bounded.Go(func() error {
			if j.CodeMonitorSearchWrapper != nil {
				return j.CodeMonitorSearchWrapper(ctx, clients.DB, clients.Gitserver, args, repoRev.Repo.ID, doSearch)
			}
			return doSearch(args)
		})
	}

	return nil, bounded.Wait()
}

func (j CommitSearch) Name() string {
	if j.Diff {
		return "Diff"
	}
	return "Commit"
}

func (j *CommitSearch) Tags() []log.Field {
	return []log.Field{
		trace.Stringer("query", j.Query),
		trace.Stringer("repoOpts", &j.RepoOpts),
		log.Bool("diff", j.Diff),
		log.Bool("hasTimeFilter", j.HasTimeFilter),
		log.Int("limit", j.Limit),
		log.Bool("includeModifiedFiles", j.IncludeModifiedFiles),
	}
}

func (j *CommitSearch) ExpandUsernames(ctx context.Context, db database.DB) (err error) {
	protocol.ReduceWith(j.Query, func(n protocol.Node) protocol.Node {
		if err != nil {
			return n
		}

		var expr *string
		switch v := n.(type) {
		case *protocol.AuthorMatches:
			expr = &v.Expr
		case *protocol.CommitterMatches:
			expr = &v.Expr
		default:
			return n
		}

		var expanded []string
		expanded, err = expandUsernamesToEmails(ctx, db, []string{*expr})
		if err != nil {
			return n
		}

		*expr = "(" + strings.Join(expanded, ")|(") + ")"
		return n
	})
	return err
}

// expandUsernamesToEmails expands references to usernames to mention all possible (known and
// verified) email addresses for the user.
//
// For example, given a list ["foo", "@alice"] where the user "alice" has 2 email addresses
// "alice@example.com" and "alice@example.org", it would return ["foo", "alice@example\\.com",
// "alice@example\\.org"].
func expandUsernamesToEmails(ctx context.Context, db database.DB, values []string) (expandedValues []string, err error) {
	expandOne := func(ctx context.Context, value string) ([]string, error) {
		if isPossibleUsernameReference := strings.HasPrefix(value, "@"); !isPossibleUsernameReference {
			return nil, nil
		}

		user, err := db.Users().GetByUsername(ctx, strings.TrimPrefix(value, "@"))
		if errcode.IsNotFound(err) {
			return nil, nil
		} else if err != nil {
			return nil, err
		}
		emails, err := db.UserEmails().ListByUser(ctx, database.UserEmailsListOptions{
			UserID: user.ID,
		})
		if err != nil {
			return nil, err
		}
		values := make([]string, 0, len(emails))
		for _, email := range emails {
			if email.VerifiedAt != nil {
				values = append(values, regexp.QuoteMeta(email.Email))
			}
		}
		return values, nil
	}

	expandedValues = make([]string, 0, len(values))
	for _, v := range values {
		x, err := expandOne(ctx, v)
		if err != nil {
			return nil, err
		}
		if x == nil {
			expandedValues = append(expandedValues, v) // not a username or couldn't expand
		} else {
			expandedValues = append(expandedValues, x...)
		}
	}
	return expandedValues, nil
}

func QueryToGitQuery(q query.Q, diff bool) gitprotocol.Node {
	return gitprotocol.Reduce(gitprotocol.NewAnd(queryNodesToPredicates(q, q.IsCaseSensitive(), diff)...))
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

func queryNodesToPredicates(nodes []query.Node, caseSensitive, diff bool) []gitprotocol.Node {
	res := make([]gitprotocol.Node, 0, len(nodes))
	for _, node := range nodes {
		var newPred gitprotocol.Node
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

func queryOperatorToPredicate(op query.Operator, caseSensitive, diff bool) gitprotocol.Node {
	switch op.Kind {
	case query.And:
		return gitprotocol.NewAnd(queryNodesToPredicates(op.Operands, caseSensitive, diff)...)
	case query.Or:
		return gitprotocol.NewOr(queryNodesToPredicates(op.Operands, caseSensitive, diff)...)
	default:
		// I don't think we should have concats at this point, but ignore it if we do
		return nil
	}
}

func queryPatternToPredicate(pattern query.Pattern, caseSensitive, diff bool) gitprotocol.Node {
	patString := pattern.Value
	if pattern.Annotation.Labels.IsSet(query.Literal) {
		patString = regexp.QuoteMeta(pattern.Value)
	}

	var newPred gitprotocol.Node
	if diff {
		newPred = &gitprotocol.DiffMatches{Expr: patString, IgnoreCase: !caseSensitive}
	} else {
		newPred = &gitprotocol.MessageMatches{Expr: patString, IgnoreCase: !caseSensitive}
	}

	if pattern.Negated {
		return gitprotocol.NewNot(newPred)
	}
	return newPred
}

func queryParameterToPredicate(parameter query.Parameter, caseSensitive, diff bool) gitprotocol.Node {
	var newPred gitprotocol.Node
	switch parameter.Field {
	case query.FieldAuthor:
		// TODO(@camdencheek) look up emails (issue #25180)
		newPred = &gitprotocol.AuthorMatches{Expr: parameter.Value, IgnoreCase: !caseSensitive}
	case query.FieldCommitter:
		newPred = &gitprotocol.CommitterMatches{Expr: parameter.Value, IgnoreCase: !caseSensitive}
	case query.FieldBefore:
		t, _ := query.ParseGitDate(parameter.Value, time.Now) // field already validated
		newPred = &gitprotocol.CommitBefore{Time: t}
	case query.FieldAfter:
		t, _ := query.ParseGitDate(parameter.Value, time.Now) // field already validated
		newPred = &gitprotocol.CommitAfter{Time: t}
	case query.FieldMessage:
		newPred = &gitprotocol.MessageMatches{Expr: parameter.Value, IgnoreCase: !caseSensitive}
	case query.FieldContent:
		if diff {
			newPred = &gitprotocol.DiffMatches{Expr: parameter.Value, IgnoreCase: !caseSensitive}
		} else {
			newPred = &gitprotocol.MessageMatches{Expr: parameter.Value, IgnoreCase: !caseSensitive}
		}
	case query.FieldFile:
		newPred = &gitprotocol.DiffModifiesFile{Expr: parameter.Value, IgnoreCase: !caseSensitive}
	case query.FieldLang:
		newPred = &gitprotocol.DiffModifiesFile{Expr: query.LangToFileRegexp(parameter.Value), IgnoreCase: true}
	}

	if parameter.Negated && newPred != nil {
		return gitprotocol.NewNot(newPred)
	}
	return newPred
}

func protocolMatchToCommitMatch(repo types.MinimalRepo, diff bool, in protocol.CommitMatch) *result.CommitMatch {
	var diffPreview, messagePreview *result.MatchedString
	if diff {
		diffPreview = &in.Diff
	} else {
		messagePreview = &in.Message
	}

	return &result.CommitMatch{
		Commit: gitdomain.Commit{
			ID: in.Oid,
			Author: gitdomain.Signature{
				Name:  in.Author.Name,
				Email: in.Author.Email,
				Date:  in.Author.Date,
			},
			Committer: &gitdomain.Signature{
				Name:  in.Committer.Name,
				Email: in.Committer.Email,
				Date:  in.Committer.Date,
			},
			Message: gitdomain.Message(in.Message.Content),
			Parents: in.Parents,
		},
		Repo:           repo,
		DiffPreview:    diffPreview,
		MessagePreview: messagePreview,
		ModifiedFiles:  in.ModifiedFiles,
	}
}

func newReposLimitError(limit int, hasTimeFilter bool, resultType string) error {
	if hasTimeFilter {
		return &TimeLimitError{ResultType: resultType, Max: limit}
	}
	return &RepoLimitError{ResultType: resultType, Max: limit}
}

func reposLimit(hasTimeFilter bool) int {
	searchLimits := limits.SearchLimits(conf.Get())
	if hasTimeFilter {
		return searchLimits.CommitDiffWithTimeFilterMaxRepos
	}
	return searchLimits.CommitDiffMaxRepos
}

type DiffCommitError struct {
	ResultType string
	Max        int
}

type (
	RepoLimitError DiffCommitError
	TimeLimitError DiffCommitError
)

func (*RepoLimitError) Error() string {
	return "repo limit error"
}

func (*TimeLimitError) Error() string {
	return "time limit error"
}
