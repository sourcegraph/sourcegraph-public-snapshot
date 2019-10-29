package resolvers

import (
	"context"
	"encoding/json"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	ee "github.com/sourcegraph/sourcegraph/enterprise/pkg/a8n"
	"github.com/sourcegraph/sourcegraph/internal/a8n"
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
	// TODO(a8n): Implement this. We need to fetch the CampaignJobs and their
	// diffs/errors and return them as a `campaignJobConnectionResolver` that
	// implements the `ChangesetsConnectionResolver` interface.
	return &changesetPlansConnectionResolver{store: r.store, campaignPlan: r.campaignPlan}
}

func (r *campaignPlanResolver) RepositoryDiffs(
	ctx context.Context,
	args *graphqlutil.ConnectionArgs,
) (graphqlbackend.PreviewRepositoryDiffConnectionResolver, error) {
	// TODO(a8n): Implement this. We need to fetch the CampaignJobs diffs
	// and return them as a `PreviewRepositoryDiffConnectionResolver`
	return nil, nil
}

type changesetPlansConnectionResolver struct {
	store        *ee.Store
	campaignPlan *a8n.CampaignPlan
}

func (r *changesetPlansConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.ChangesetPlanResolver, error) {
	return []graphqlbackend.ChangesetPlanResolver{}, nil
}

func (r *changesetPlansConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	return 0, nil
}

func (r *changesetPlansConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(false), nil
}
