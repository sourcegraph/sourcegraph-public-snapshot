package resolvers

import (
	"context"
	"encoding/json"
	"errors"
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
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
	b, err := json.Marshal(r.campaignPlan.Arguments)
	if err != nil {
		return graphqlbackend.JSONCString(""), err
	}

	return graphqlbackend.JSONCString(string(b)), nil
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
) (graphqlbackend.PreviewRepositoryDiffConnectionResolver, error) {
	// TODO(a8n): Implement this. We need to fetch the CampaignJobs diffs
	// and return them as a `PreviewRepositoryDiffConnectionResolver`
	return nil, nil
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
		resolvers = append(resolvers, &campaignJobResolver{
			store:        r.store,
			job:          j,
			campaignPlan: r.campaignPlan,
		})
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
	store        *ee.Store
	job          *a8n.CampaignJob
	campaignPlan *a8n.CampaignPlan
}

func (r *campaignJobResolver) Title() (string, error) { return "Title placeholder", nil }
func (r *campaignJobResolver) Body() (string, error)  { return "Body placeholder", nil }
func (r *campaignJobResolver) Repository(ctx context.Context) (*graphqlbackend.RepositoryResolver, error) {
	return graphqlbackend.RepositoryByIDInt32(ctx, api.RepoID(r.job.RepoID))
}

func (r *campaignJobResolver) Diff(ctx context.Context) (graphqlbackend.PreviewRepositoryDiff, error) {
	return nil, errors.New("not implemented")
}
