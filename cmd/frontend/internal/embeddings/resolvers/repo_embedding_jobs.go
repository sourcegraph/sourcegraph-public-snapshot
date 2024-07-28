package resolvers

import (
	"context"
	"strings"
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	repobg "github.com/sourcegraph/sourcegraph/internal/embeddings/background/repo"
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
) (*gqlutil.ConnectionResolver[graphqlbackend.RepoEmbeddingJobResolver], error) {
	store := &repoEmbeddingJobsConnectionStore{
		db:              db,
		gitserverClient: gitserverClient,
		store:           repoEmbeddingJobsStore,
		args:            args,
	}
	opts := &gqlutil.ConnectionResolverOptions{
		OrderBy: database.OrderBy{
			{Field: "repo_embedding_jobs.id"},
		},
	}
	return gqlutil.NewConnectionResolver[graphqlbackend.RepoEmbeddingJobResolver](store, &args.ConnectionResolverArgs, opts)
}

type repoEmbeddingJobsConnectionStore struct {
	db              database.DB
	gitserverClient gitserver.Client
	store           repobg.RepoEmbeddingJobsStore
	args            graphqlbackend.ListRepoEmbeddingJobsArgs
}

func withRepoID(o *repobg.ListOpts, id graphql.ID) error {
	var repoID api.RepoID
	if err := relay.UnmarshalSpec(id, &repoID); err != nil {
		return err
	}
	o.Repo = &repoID
	return nil
}

func (s *repoEmbeddingJobsConnectionStore) ComputeTotal(ctx context.Context) (int32, error) {
	opts := repobg.ListOpts{Query: s.args.Query, State: s.args.State}
	if s.args.Repo != nil {
		err := withRepoID(&opts, *s.args.Repo)
		if err != nil {
			return 0, err
		}
	}

	count, err := s.store.CountRepoEmbeddingJobs(ctx, opts)
	if err != nil {
		return 0, err
	}

	return int32(count), nil
}

func (s *repoEmbeddingJobsConnectionStore) ComputeNodes(ctx context.Context, args *database.PaginationArgs) ([]graphqlbackend.RepoEmbeddingJobResolver, error) {
	opts := repobg.ListOpts{PaginationArgs: args, Query: s.args.Query, State: s.args.State}
	if s.args.Repo != nil {
		err := withRepoID(&opts, *s.args.Repo)
		if err != nil {
			return nil, err
		}
	}

	jobs, err := s.store.ListRepoEmbeddingJobs(ctx, opts)
	if err != nil {
		return nil, err
	}
	resolvers := make([]graphqlbackend.RepoEmbeddingJobResolver, 0, len(jobs))
	for _, job := range jobs {
		resolvers = append(resolvers, &repoEmbeddingJobResolver{
			db:              s.db,
			gitserverClient: s.gitserverClient,
			job:             job,
		})
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

func (s *repoEmbeddingJobsConnectionStore) UnmarshalCursor(cursor string, _ database.OrderBy) ([]any, error) {
	nodeID, err := unmarshalRepoEmbeddingJobID(graphql.ID(cursor))
	if err != nil {
		return nil, err
	}

	return []any{nodeID}, nil
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

func (r *repoEmbeddingJobResolver) Stats(ctx context.Context) (graphqlbackend.RepoEmbeddingJobStatsResolver, error) {
	store := repobg.NewRepoEmbeddingJobsStore(r.db)
	stats, err := store.GetRepoEmbeddingJobStats(ctx, r.job.ID)
	if err != nil {
		return nil, err
	}
	return &repoEmbeddingJobStatsResolver{stats}, nil
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

	// An empty revision value can accompany a valid repository if gitserver cannot resolve the default branch or latest revision during job scheduling.
	// The job will always fail in this case and must be displayed in site admin despite the gitserver error.
	// Site admin will only provide the job's failure_message in this case.
	invalidRevision := r.job.Revision == ""

	if repoResolver == nil || invalidRevision {
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

type repoEmbeddingJobStatsResolver struct {
	stats repobg.EmbedRepoStats
}

func (r *repoEmbeddingJobStatsResolver) FilesScheduled() int32 {
	return int32(r.stats.CodeIndexStats.FilesScheduled + r.stats.TextIndexStats.FilesScheduled)
}

func (r *repoEmbeddingJobStatsResolver) FilesEmbedded() int32 {
	return int32(r.stats.CodeIndexStats.FilesEmbedded + r.stats.TextIndexStats.FilesEmbedded)
}

func (r *repoEmbeddingJobStatsResolver) FilesSkipped() int32 {
	skipped := 0
	for _, count := range r.stats.CodeIndexStats.FilesSkipped {
		skipped += count
	}
	for _, count := range r.stats.TextIndexStats.FilesSkipped {
		skipped += count
	}
	return int32(skipped)
}
