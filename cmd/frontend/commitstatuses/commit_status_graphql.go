package commitstatuses

import (
	"context"
	"sort"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

func (GraphQLResolver) CommitStatusForCommit(ctx context.Context, repositoryID graphql.ID, commitID api.CommitID) (graphqlbackend.CommitStatus, error) {
	// Check existence.
	repository, err := graphqlbackend.RepositoryByID(ctx, repositoryID)
	if err != nil {
		return nil, err
	}

	commit, err := graphqlbackend.GetGitCommit(ctx, repository, graphqlbackend.GitObjectID(commitID))
	if err != nil {
		return nil, err
	}

	list, err := dbCommitStatusContexts{}.List(ctx, dbCommitStatusContextsListOptions{
		RepositoryID: repository.DBID(),
		CommitOID:    commitID,
	})
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, nil // no status
	}
	return &gqlCommitStatus{commit: commit, db: list}, nil
}

// gqlCommitStatus implements the GraphQL type CommitStatus.
type gqlCommitStatus struct {
	commit *graphqlbackend.GitCommitResolver
	db     []*dbCommitStatusContext
}

func (v *gqlCommitStatus) Repository(ctx context.Context) (*graphqlbackend.RepositoryResolver, error) {
	return v.commit.Repository(), nil
}

func (v *gqlCommitStatus) Commit(ctx context.Context) (*graphqlbackend.GitCommitResolver, error) {
	return v.commit, nil
}

// getLatestPerContext returns a list of commit status contexts that are each the newest of all
// others with the same context. For example, if there are 2 commit status contexts with context
// "foo", only the newest will be in the returned slice.
func (v *gqlCommitStatus) getNewestContexts() []*dbCommitStatusContext {
	newest := map[string]*dbCommitStatusContext{}
	for _, c := range v.db {
		if existing, ok := newest[c.Context]; !ok || c.CreatedAt.After(existing.CreatedAt) {
			newest[c.Context] = c
		}
	}

	cs := make([]*dbCommitStatusContext, 0, len(newest))
	for _, c := range cs {
		cs = append(cs, c)
	}
	sort.Slice(cs, func(i, j int) bool { return cs[i].Context < cs[j].Context })
	return cs
}

func (v *gqlCommitStatus) Contexts(ctx context.Context) ([]graphqlbackend.CommitStatusContext, error) {
	cs := v.getNewestContexts()
	gqls := make([]graphqlbackend.CommitStatusContext, len(cs))
	for i, c := range cs {
		gqls[i] = &gqlCommitStatusContext{*c}
	}
	return gqls, nil
}

func (v *gqlCommitStatus) State() graphqlbackend.CommitStatusState {
	if len(v.db) == 0 {
		// The combined state is undefined when there are no contexts. This should never happen.
		panic("unexpected empty commit status contexts list")
	}

	// The combined state is the "worst" state of all contexts' states, in order of enum definition
	// from worst to best (except that ERROR trumps EXPECTED because that is probably more
	// helpful/expected to the user).
	var combinedState graphqlbackend.CommitStatusState
	for _, c := range v.getNewestContexts() {
		switch c.State {
		case string(graphqlbackend.CommitStatusStateExpected):
			if combinedState == "" || combinedState == graphqlbackend.CommitStatusStateSuccess || combinedState == graphqlbackend.CommitStatusStatePending || combinedState == graphqlbackend.CommitStatusStateFailure {
				combinedState = graphqlbackend.CommitStatusState(c.State)
			}
		case string(graphqlbackend.CommitStatusStateError):
			if combinedState == "" || combinedState == graphqlbackend.CommitStatusStateSuccess || combinedState == graphqlbackend.CommitStatusStatePending || combinedState == graphqlbackend.CommitStatusStateFailure || combinedState == graphqlbackend.CommitStatusStateExpected {
				combinedState = graphqlbackend.CommitStatusState(c.State)
			}
		case string(graphqlbackend.CommitStatusStateFailure):
			if combinedState == "" || combinedState == graphqlbackend.CommitStatusStateSuccess || combinedState == graphqlbackend.CommitStatusStatePending {
				combinedState = graphqlbackend.CommitStatusState(c.State)
			}
		case string(graphqlbackend.CommitStatusStatePending):
			if combinedState == "" || combinedState == graphqlbackend.CommitStatusStateSuccess {
				combinedState = graphqlbackend.CommitStatusState(c.State)
			}
		case string(graphqlbackend.CommitStatusStateSuccess):
			if combinedState == "" {
				combinedState = graphqlbackend.CommitStatusState(c.State)
			}
		default:
			combinedState = graphqlbackend.CommitStatusStateError
		}
	}
	return combinedState
}
