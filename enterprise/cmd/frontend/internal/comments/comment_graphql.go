package comments

import (
	"context"
	"path"
	"sort"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

// 🚨 SECURITY: TODO!(sqs): there needs to be security checks everywhere here! there are none

// gqlComment implements the GraphQL type Comment.
type gqlComment struct{ db *dbComment }

// commentByID looks up and returns the Comment with the given GraphQL ID. If no such Comment exists, it
// returns a non-nil error.
func commentByID(ctx context.Context, id graphql.ID) (*gqlComment, error) {
	dbID, err := unmarshalCommentID(id)
	if err != nil {
		return nil, err
	}
	return commentByDBID(ctx, dbID)
}

func (GraphQLResolver) CommentByID(ctx context.Context, id graphql.ID) (graphqlbackend.Comment, error) {
	return commentByID(ctx, id)
}

// commentByDBID looks up and returns the Comment with the given database ID. If no such Comment exists,
// it returns a non-nil error.
func commentByDBID(ctx context.Context, dbID int64) (*gqlComment, error) {
	v, err := dbComments{}.GetByID(ctx, dbID)
	if err != nil {
		return nil, err
	}
	return &gqlComment{db: v}, nil
}

func (v *gqlComment) ID() graphql.ID {
	return marshalCommentID(v.db.ID)
}

func marshalCommentID(id int64) graphql.ID {
	return relay.MarshalID("Comment", id)
}

func unmarshalCommentID(id graphql.ID) (dbID int64, err error) {
	err = relay.UnmarshalSpec(id, &dbID)
	return
}

func (v *gqlComment) Namespace(ctx context.Context) (*graphqlbackend.NamespaceResolver, error) {
	return graphqlbackend.NamespaceByDBID(ctx, v.db.NamespaceUserID, v.db.NamespaceOrgID)
}

func (v *gqlComment) Name() string { return v.db.Name }

func (v *gqlComment) Description() *string { return v.db.Description }

func (v *gqlComment) IsPreview() bool { return v.db.IsPreview }

func (v *gqlComment) Rules() string {
	if v.db.Rules != "" {
		return v.db.Rules
	}
	return "[]"
}

func (v *gqlComment) URL(ctx context.Context) (string, error) {
	namespace, err := v.Namespace(ctx)
	if err != nil {
		return "", err
	}

	var preview string
	if v.db.IsPreview {
		preview = "preview"
	}

	return path.Join(namespace.URL(), "comments", preview, string(v.ID())), nil
	//
	// TODO!(sqs): use global url?
	// return path.Join("/comments", string(v.ID())), nil
}

func (v *gqlComment) ThreadOrIssueOrChangesets(ctx context.Context, arg *graphqlutil.ConnectionArgs) (graphqlbackend.ThreadOrIssueOrChangesetConnection, error) {
	opt := dbCommentsThreadsListOptions{CommentID: v.db.ID}
	arg.Set(&opt.LimitOffset)
	l, err := dbCommentsThreads{}.List(ctx, opt)
	if err != nil {
		return nil, err
	}

	threadlikeIDs := make([]int64, len(l))
	for i, e := range l {
		threadlikeIDs[i] = e.Thread
	}
	return threadlike.ThreadOrIssueOrChangesetsByIDs(ctx, threadlikeIDs, arg)
}

func (v *gqlComment) getChangesets(ctx context.Context) ([]graphqlbackend.Changeset, error) {
	connection, err := v.ThreadOrIssueOrChangesets(ctx, &graphqlutil.ConnectionArgs{})
	if err != nil {
		return nil, err
	}
	nodes, err := connection.Nodes(ctx)
	if err != nil {
		return nil, err
	}

	// TODO!(sqs): easier way to filter down to only changesets
	changesets := make([]graphqlbackend.Changeset, 0, len(nodes))
	for _, node := range nodes {
		if changeset, ok := node.ToChangeset(); ok {
			changesets = append(changesets, changeset)
		}
	}
	return changesets, nil
}

func (v *gqlComment) Repositories(ctx context.Context) ([]*graphqlbackend.RepositoryResolver, error) {
	threadNodes, err := v.getChangesets(ctx)
	if err != nil {
		return nil, err
	}

	byRepositoryDBID := map[api.RepoID]*graphqlbackend.RepositoryResolver{}
	for _, thread := range threadNodes {
		repo, err := thread.Repository(ctx)
		if err != nil {
			return nil, err
		}
		key := repo.DBID()
		if _, seen := byRepositoryDBID[key]; !seen {
			byRepositoryDBID[key] = repo
		}
	}

	repos := make([]*graphqlbackend.RepositoryResolver, 0, len(byRepositoryDBID))
	for _, repo := range byRepositoryDBID {
		repos = append(repos, repo)
	}
	sort.Slice(repos, func(i, j int) bool {
		return repos[i].DBID() < repos[j].DBID()
	})
	return repos, nil
}

func (v *gqlComment) Commits(ctx context.Context) ([]*graphqlbackend.GitCommitResolver, error) {
	rcs, err := v.RepositoryComparisons(ctx)
	if err != nil {
		return nil, err
	}

	var allCommits []*graphqlbackend.GitCommitResolver
	for _, rc := range rcs {
		cc := rc.Commits(&graphqlutil.ConnectionArgs{})
		commits, err := cc.Nodes(ctx)
		if err != nil {
			return nil, err
		}
		allCommits = append(allCommits, commits...)
	}
	return allCommits, nil
}

func (v *gqlComment) RepositoryComparisons(ctx context.Context) ([]*graphqlbackend.RepositoryComparisonResolver, error) {
	changesets, err := v.getChangesets(ctx)
	if err != nil {
		return nil, err
	}

	rcs := make([]*graphqlbackend.RepositoryComparisonResolver, len(changesets))
	for i, changeset := range changesets {
		rc, err := changeset.RepositoryComparison(ctx)
		if err != nil {
			return nil, err
		}
		rcs[i] = rc
	}
	return rcs, nil
}
