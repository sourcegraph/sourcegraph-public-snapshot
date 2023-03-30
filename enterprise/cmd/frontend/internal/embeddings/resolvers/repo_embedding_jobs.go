package resolvers

import (
	"context"
	"strconv"
	"strings"
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	repobg "github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/background/repo"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func NewRepoEmbeddingJobsResolver(
	db database.DB,
	gitserverClient gitserver.Client,
	repoEmbeddingJobsStore repobg.RepoEmbeddingJobsStore,
	args graphqlbackend.ListRepoEmbeddingJobsArgs,
) (*graphqlutil.ConnectionResolver[graphqlbackend.RepoEmbeddingJobResolver], error) {
	store := &repoEmbeddingJobsConnectionStore{
		db:              db,
		gitserverClient: gitserverClient,
		store:           repoEmbeddingJobsStore,
		args:            args,
	}
	return graphqlutil.NewConnectionResolver[graphqlbackend.RepoEmbeddingJobResolver](store, &args.ConnectionResolverArgs, nil)
}

type repoEmbeddingJobsConnectionStore struct {
	db              database.DB
	gitserverClient gitserver.Client
	store           repobg.RepoEmbeddingJobsStore
	args            graphqlbackend.ListRepoEmbeddingJobsArgs
}

func (s *repoEmbeddingJobsConnectionStore) ComputeTotal(ctx context.Context) (*int32, error) {
	count, err := s.store.CountRepoEmbeddingJobs(ctx)
	if err != nil {
		return nil, err
	}
	total := int32(count)
	return &total, nil
}

func (s *repoEmbeddingJobsConnectionStore) ComputeNodes(ctx context.Context, args *database.PaginationArgs) ([]graphqlbackend.RepoEmbeddingJobResolver, error) {
	jobs, err := s.store.ListRepoEmbeddingJobs(ctx, args)
	if err != nil {
		return nil, err
	}
	resolvers := make([]graphqlbackend.RepoEmbeddingJobResolver, 0, len(jobs))
	for _, job := range jobs {
		resolvers = append(resolvers, &repoEmbeddingJobResolver{db: s.db, gitserverClient: s.gitserverClient, job: job})
	}
	return resolvers, nil
}

func (s *repoEmbeddingJobsConnectionStore) MarshalCursor(node graphqlbackend.RepoEmbeddingJobResolver, _ database.OrderBy) (*string, error) {
	if node == nil {
		return nil, errors.New("node is nil")
	}
	cursor := string(node.ID())
	return &cursor, nil
}

func (s *repoEmbeddingJobsConnectionStore) UnmarshalCursor(cursor string, _ database.OrderBy) (*string, error) {
	nodeID, err := unmarshalRepoEmbeddingJobID(graphql.ID(cursor))
	if err != nil {
		return nil, err
	}
	id := strconv.Itoa(nodeID)
	return &id, nil
}

type repoEmbeddingJobResolver struct {
	db              database.DB
	gitserverClient gitserver.Client
	job             *repobg.RepoEmbeddingJob
	// cache results because they are used by multiple fields
	once         sync.Once
	repoResolver *graphqlbackend.RepositoryResolver
	err          error
}

func (r *repoEmbeddingJobResolver) ID() graphql.ID {
	return marshalRepoEmbeddingJobID(r.job.ID)
}

func (r *repoEmbeddingJobResolver) State() string {
	return strings.ToUpper(r.job.State)
}

func (r *repoEmbeddingJobResolver) FailureMessage() *string {
	return r.job.FailureMessage
}

func (r *repoEmbeddingJobResolver) QueuedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.job.QueuedAt}
}

func (r *repoEmbeddingJobResolver) StartedAt() *gqlutil.DateTime {
	if r.job.StartedAt == nil {
		return nil
	}
	return gqlutil.FromTime(*r.job.StartedAt)
}

func (r *repoEmbeddingJobResolver) FinishedAt() *gqlutil.DateTime {
	if r.job.FinishedAt == nil {
		return nil
	}
	return gqlutil.FromTime(*r.job.FinishedAt)
}

func (r *repoEmbeddingJobResolver) ProcessAfter() *gqlutil.DateTime {
	if r.job.ProcessAfter == nil {
		return nil
	}
	return gqlutil.FromTime(*r.job.ProcessAfter)
}

func (r *repoEmbeddingJobResolver) NumResets() int32 {
	return int32(r.job.NumResets)
}

func (r *repoEmbeddingJobResolver) NumFailures() int32 {
	return int32(r.job.NumFailures)
}

func (r *repoEmbeddingJobResolver) LastHeartbeatAt() *gqlutil.DateTime {
	if r.job.ProcessAfter == nil {
		return nil
	}
	return gqlutil.FromTime(*r.job.LastHeartbeatAt)
}

func (r *repoEmbeddingJobResolver) WorkerHostname() string {
	return r.job.WorkerHostname
}

func (r *repoEmbeddingJobResolver) Cancel() bool {
	return r.job.Cancel
}

func (r *repoEmbeddingJobResolver) compute(ctx context.Context) (*graphqlbackend.RepositoryResolver, error) {
	r.once.Do(func() {
		repo, err := r.db.Repos().Get(ctx, r.job.RepoID)
		if err != nil {
			if errcode.IsNotFound(err) {
				// Skip resolving repository if it does not exist.
				r.repoResolver, r.err = nil, nil
				return
			}
			r.repoResolver, r.err = nil, err
			return
		}
		r.repoResolver, r.err = graphqlbackend.NewRepositoryResolver(r.db, r.gitserverClient, repo), nil
	})
	return r.repoResolver, r.err
}

func (r *repoEmbeddingJobResolver) Repo(ctx context.Context) (*graphqlbackend.RepositoryResolver, error) {
	return r.compute(ctx)
}

func (r *repoEmbeddingJobResolver) Revision(ctx context.Context) (*graphqlbackend.GitCommitResolver, error) {
	repoResolver, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	if repoResolver == nil {
		return nil, nil
	}
	return graphqlbackend.NewGitCommitResolver(r.db, r.gitserverClient, repoResolver, r.job.Revision, nil), nil
}

func marshalRepoEmbeddingJobID(id int) graphql.ID {
	return relay.MarshalID("RepoEmbeddingJob", id)
}

func unmarshalRepoEmbeddingJobID(id graphql.ID) (jobID int, err error) {
	err = relay.UnmarshalSpec(id, &jobID)
	return
}
