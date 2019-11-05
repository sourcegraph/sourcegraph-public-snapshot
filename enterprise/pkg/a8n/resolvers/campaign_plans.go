package resolvers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	ee "github.com/sourcegraph/sourcegraph/enterprise/pkg/a8n"
	"github.com/sourcegraph/sourcegraph/internal/a8n"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

const campaignPlanIDKind = "CampaignPlan"

func marshalCampaignPlanID(id int64) graphql.ID {
	return relay.MarshalID(campaignPlanIDKind, id)
}

func unmarshalCampaignPlanID(id graphql.ID) (campaignPlanID int64, err error) {
	err = relay.UnmarshalSpec(id, &campaignPlanID)
	return
}

type campaignPlanResolver struct {
	store        *ee.Store
	campaignPlan *a8n.CampaignPlan
}

func (r *campaignPlanResolver) ID() graphql.ID {
	return marshalCampaignPlanID(r.campaignPlan.ID)
}

func (r *campaignPlanResolver) Type() string { return r.campaignPlan.CampaignType }
func (r *campaignPlanResolver) Arguments() (graphqlbackend.JSONCString, error) {
	return graphqlbackend.JSONCString(r.campaignPlan.Arguments), nil
}

func (r *campaignPlanResolver) Status() graphqlbackend.BackgroundProcessStatus {
	// TODO(a8n): Implement this
	return a8n.BackgroundProcessStatus{
		Completed:     0,
		Pending:       99,
		ProcessState:  a8n.BackgroundProcessStateErrored,
		ProcessErrors: []string{"this is just a skeleton api"},
	}
}

func (r *campaignPlanResolver) Changesets(
	ctx context.Context,
	args *graphqlutil.ConnectionArgs,
) graphqlbackend.ChangesetPlansConnectionResolver {
	return &campaignJobsConnectionResolver{
		store:        r.store,
		campaignPlan: r.campaignPlan,
	}
}

func (r *campaignPlanResolver) RepositoryDiffs(
	ctx context.Context,
	args *graphqlutil.ConnectionArgs,
) (graphqlbackend.ChangesetPlansConnectionResolver, error) {
	return &campaignJobsConnectionResolver{
		store:        r.store,
		campaignPlan: r.campaignPlan,
	}, nil
}

type campaignJobsConnectionResolver struct {
	store        *ee.Store
	campaignPlan *a8n.CampaignPlan

	// cache results because they are used by multiple fields
	once sync.Once
	jobs []*a8n.CampaignJob
	next int64
	err  error
}

func (r *campaignJobsConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.ChangesetPlanResolver, error) {
	jobs, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]graphqlbackend.ChangesetPlanResolver, 0, len(jobs))
	for _, j := range jobs {
		resolvers = append(resolvers, &campaignJobResolver{job: j})
	}
	return resolvers, nil
}

func (r *campaignJobsConnectionResolver) compute(ctx context.Context) ([]*a8n.CampaignJob, int64, error) {
	r.once.Do(func() {
		r.jobs, r.next, r.err = r.store.ListCampaignJobs(ctx, ee.ListCampaignJobsOpts{
			CampaignPlanID: r.campaignPlan.ID,
		})
		// TODO(a8n): To avoid n+1 queries, we could preload all repositories here
		// and save them
	})
	return r.jobs, r.next, r.err
}

func (r *campaignJobsConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	opts := ee.CountCampaignJobsOpts{CampaignPlanID: r.campaignPlan.ID}
	count, err := r.store.CountCampaignJobs(ctx, opts)
	return int32(count), err
}

func (r *campaignJobsConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, next, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	return graphqlutil.HasNextPage(next != 0), nil
}

type campaignJobResolver struct {
	job *a8n.CampaignJob

	// cache repo because it's called more than one time
	once   sync.Once
	repo   *graphqlbackend.RepositoryResolver
	commit *graphqlbackend.GitCommitResolver
	err    error
}

func (r *campaignJobResolver) Title() (string, error) { return "Title placeholder", nil }
func (r *campaignJobResolver) Body() (string, error)  { return "Body placeholder", nil }

func (r *campaignJobResolver) computeRepoCommit(ctx context.Context) (*graphqlbackend.RepositoryResolver, *graphqlbackend.GitCommitResolver, error) {
	r.once.Do(func() {
		r.repo, r.err = graphqlbackend.RepositoryByIDInt32(ctx, api.RepoID(r.job.RepoID))
		if r.err != nil {
			return
		}
		args := &graphqlbackend.RepositoryCommitArgs{Rev: string(r.job.Rev)}
		r.commit, r.err = r.repo.Commit(ctx, args)
	})
	return r.repo, r.commit, r.err
}

func (r *campaignJobResolver) Repository(ctx context.Context) (*graphqlbackend.RepositoryResolver, error) {
	repo, _, err := r.computeRepoCommit(ctx)
	return repo, err
}

func (r *campaignJobResolver) BaseRepository(ctx context.Context) (*graphqlbackend.RepositoryResolver, error) {
	repo, _, err := r.computeRepoCommit(ctx)
	return repo, err
}

func (r *campaignJobResolver) Diff(ctx context.Context) graphqlbackend.ChangesetPlanResolver {
	return r
}

func (r *campaignJobResolver) FileDiffs(ctx context.Context, args *graphqlutil.ConnectionArgs) (graphqlbackend.PreviewFileDiffConnection, error) {
	_, commit, err := r.computeRepoCommit(ctx)
	if err != nil {
		return nil, err
	}
	return &previewFileDiffConnectionResolver{
		job:    r.job,
		commit: commit,
		first:  args.First,
	}, nil
}

type previewFileDiffConnectionResolver struct {
	job    *a8n.CampaignJob
	commit *graphqlbackend.GitCommitResolver
	first  *int32

	// cache result because it is used by multiple fields
	once        sync.Once
	fileDiffs   []*diff.FileDiff
	hasNextPage bool
	err         error
}

func (r *previewFileDiffConnectionResolver) compute(ctx context.Context) ([]*diff.FileDiff, error) {
	r.once.Do(func() {
		r.fileDiffs, r.err = diff.ParseMultiFileDiff([]byte(r.job.Diff))
		if r.err != nil {
			return
		}

		if r.first != nil && len(r.fileDiffs) > int(*r.first) {
			r.hasNextPage = true
		}
	})
	return r.fileDiffs, r.err
}

func (r *previewFileDiffConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.PreviewFileDiff, error) {
	fileDiffs, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if r.first != nil && len(fileDiffs) > int(*r.first) {
		fileDiffs = fileDiffs[:*r.first]
	}

	resolvers := make([]graphqlbackend.PreviewFileDiff, len(fileDiffs))
	for i, fileDiff := range fileDiffs {
		resolvers[i] = &previewFileDiffResolver{
			fileDiff: fileDiff,
			commit:   r.commit,
		}
	}
	return resolvers, nil
}

func (r *previewFileDiffConnectionResolver) TotalCount(ctx context.Context) (*int32, error) {
	fileDiffs, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	if r.first == nil || (len(fileDiffs) > int(*r.first)) {
		n := int32(len(fileDiffs))
		return &n, nil
	}
	// This is taken from fileDiffConnectionResolver.TotalCount
	return nil, nil
}

func (r *previewFileDiffConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	if _, err := r.compute(ctx); err != nil {
		return nil, err
	}
	return graphqlutil.HasNextPage(r.hasNextPage), nil
}

func (r *previewFileDiffConnectionResolver) DiffStat(ctx context.Context) (*graphqlbackend.DiffStat, error) {
	fileDiffs, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	stat := &graphqlbackend.DiffStat{}
	for _, fileDiff := range fileDiffs {
		s := fileDiff.Stat()
		stat.AddStat(s)
	}
	return stat, nil
}
func (r *previewFileDiffConnectionResolver) RawDiff(ctx context.Context) (string, error) {
	fileDiffs, err := r.compute(ctx)
	if err != nil {
		return "", err
	}
	b, err := diff.PrintMultiFileDiff(fileDiffs)
	return string(b), err
}

type previewFileDiffResolver struct {
	fileDiff *diff.FileDiff
	commit   *graphqlbackend.GitCommitResolver
}

func (r *previewFileDiffResolver) OldPath() *string { return diffPathOrNull(r.fileDiff.OrigName) }
func (r *previewFileDiffResolver) NewPath() *string { return diffPathOrNull(r.fileDiff.NewName) }

func (r *previewFileDiffResolver) Hunks() []*graphqlbackend.DiffHunk {
	hunks := make([]*graphqlbackend.DiffHunk, len(r.fileDiff.Hunks))
	for i, hunk := range r.fileDiff.Hunks {
		hunks[i] = graphqlbackend.NewDiffHunk(hunk)
	}
	return hunks
}

func (r *previewFileDiffResolver) Stat() *graphqlbackend.DiffStat {
	stat := r.fileDiff.Stat()
	return graphqlbackend.NewDiffStat(stat)
}

func (r *previewFileDiffResolver) OldFile() *graphqlbackend.GitTreeEntryResolver {
	fileStat := graphqlbackend.CreateFileInfo(r.fileDiff.OrigName, false)
	return graphqlbackend.NewGitTreeEntryResolver(r.commit, fileStat)
}

func (r *previewFileDiffResolver) InternalID() string {
	b := sha256.Sum256([]byte(fmt.Sprintf("%d:%s:%s", len(r.fileDiff.OrigName), r.fileDiff.OrigName, r.fileDiff.NewName)))
	return hex.EncodeToString(b[:])[:32]
}

func diffPathOrNull(path string) *string {
	if path == "/dev/null" || path == "" {
		return nil
	}
	return &path
}
