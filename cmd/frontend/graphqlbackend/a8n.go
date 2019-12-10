package graphqlbackend

import (
	"context"
	"database/sql"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/externallink"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/a8n"
)

// NewA8NResolver will be set by enterprise
var NewA8NResolver func(*sql.DB) A8NResolver

type AddChangesetsToCampaignArgs struct {
	Campaign   graphql.ID
	Changesets []graphql.ID
}

type CreateCampaignArgs struct {
	Input struct {
		Namespace   graphql.ID
		Name        string
		Description string
		Plan        *graphql.ID
	}
}

type UpdateCampaignArgs struct {
	Input struct {
		ID          graphql.ID
		Name        *string
		Description *string
	}
}

type PreviewCampaignPlanArgs struct {
	Specification struct {
		Type      string
		Arguments JSONCString
	}
	Wait bool
}

type CancelCampaignPlanArgs struct {
	Plan graphql.ID
}

type DeleteCampaignArgs struct {
	Campaign graphql.ID
}

type RetryCampaignArgs struct {
	Campaign graphql.ID
}

type CloseCampaignArgs struct {
	Campaign        graphql.ID
	CloseChangesets bool
}

type CreateChangesetsArgs struct {
	Input []struct {
		Repository graphql.ID
		ExternalID string
	}
}

type A8NResolver interface {
	CreateCampaign(ctx context.Context, args *CreateCampaignArgs) (CampaignResolver, error)
	UpdateCampaign(ctx context.Context, args *UpdateCampaignArgs) (CampaignResolver, error)
	CampaignByID(ctx context.Context, id graphql.ID) (CampaignResolver, error)
	Campaigns(ctx context.Context, args *graphqlutil.ConnectionArgs) (CampaignsConnectionResolver, error)
	DeleteCampaign(ctx context.Context, args *DeleteCampaignArgs) (*EmptyResponse, error)
	RetryCampaign(ctx context.Context, args *RetryCampaignArgs) (CampaignResolver, error)
	CloseCampaign(ctx context.Context, args *CloseCampaignArgs) (CampaignResolver, error)

	CreateChangesets(ctx context.Context, args *CreateChangesetsArgs) ([]ExternalChangesetResolver, error)
	ChangesetByID(ctx context.Context, id graphql.ID) (ExternalChangesetResolver, error)
	Changesets(ctx context.Context, args *graphqlutil.ConnectionArgs) (ExternalChangesetsConnectionResolver, error)

	AddChangesetsToCampaign(ctx context.Context, args *AddChangesetsToCampaignArgs) (CampaignResolver, error)

	PreviewCampaignPlan(ctx context.Context, args PreviewCampaignPlanArgs) (CampaignPlanResolver, error)
	CampaignPlanByID(ctx context.Context, id graphql.ID) (CampaignPlanResolver, error)
	CancelCampaignPlan(ctx context.Context, args CancelCampaignPlanArgs) (*EmptyResponse, error)
}

var a8nOnlyInEnterprise = errors.New("campaigns and changesets are only available in enterprise")

func (r *schemaResolver) AddChangesetsToCampaign(ctx context.Context, args *AddChangesetsToCampaignArgs) (CampaignResolver, error) {
	if EnterpriseResolvers.a8nResolver == nil {
		return nil, a8nOnlyInEnterprise
	}
	return EnterpriseResolvers.a8nResolver.AddChangesetsToCampaign(ctx, args)
}

func (r *schemaResolver) PreviewCampaignPlan(ctx context.Context, args PreviewCampaignPlanArgs) (CampaignPlanResolver, error) {
	if EnterpriseResolvers.a8nResolver == nil {
		return nil, a8nOnlyInEnterprise
	}
	return EnterpriseResolvers.a8nResolver.PreviewCampaignPlan(ctx, args)
}

func (r *schemaResolver) CancelCampaignPlan(ctx context.Context, args CancelCampaignPlanArgs) (*EmptyResponse, error) {
	if EnterpriseResolvers.a8nResolver == nil {
		return nil, a8nOnlyInEnterprise
	}
	return EnterpriseResolvers.a8nResolver.CancelCampaignPlan(ctx, args)
}

func (r *schemaResolver) CreateCampaign(ctx context.Context, args *CreateCampaignArgs) (CampaignResolver, error) {
	if EnterpriseResolvers.a8nResolver == nil {
		return nil, a8nOnlyInEnterprise
	}
	return EnterpriseResolvers.a8nResolver.CreateCampaign(ctx, args)
}

func (r *schemaResolver) UpdateCampaign(ctx context.Context, args *UpdateCampaignArgs) (CampaignResolver, error) {
	if EnterpriseResolvers.a8nResolver == nil {
		return nil, a8nOnlyInEnterprise
	}
	return EnterpriseResolvers.a8nResolver.UpdateCampaign(ctx, args)
}

func (r *schemaResolver) DeleteCampaign(ctx context.Context, args *DeleteCampaignArgs) (*EmptyResponse, error) {
	if EnterpriseResolvers.a8nResolver == nil {
		return nil, a8nOnlyInEnterprise
	}
	return EnterpriseResolvers.a8nResolver.DeleteCampaign(ctx, args)
}

func (r *schemaResolver) RetryCampaign(ctx context.Context, args *RetryCampaignArgs) (CampaignResolver, error) {
	if EnterpriseResolvers.a8nResolver == nil {
		return nil, a8nOnlyInEnterprise
	}
	return EnterpriseResolvers.a8nResolver.RetryCampaign(ctx, args)
}

func (r *schemaResolver) CloseCampaign(ctx context.Context, args *CloseCampaignArgs) (CampaignResolver, error) {
	if EnterpriseResolvers.a8nResolver == nil {
		return nil, a8nOnlyInEnterprise
	}
	return EnterpriseResolvers.a8nResolver.CloseCampaign(ctx, args)
}

func (r *schemaResolver) Campaigns(ctx context.Context, args *graphqlutil.ConnectionArgs) (CampaignsConnectionResolver, error) {
	if EnterpriseResolvers.a8nResolver == nil {
		return nil, a8nOnlyInEnterprise
	}
	return EnterpriseResolvers.a8nResolver.Campaigns(ctx, args)
}

func (r *schemaResolver) CreateChangesets(ctx context.Context, args *CreateChangesetsArgs) ([]ExternalChangesetResolver, error) {
	if EnterpriseResolvers.a8nResolver == nil {
		return nil, a8nOnlyInEnterprise
	}
	return EnterpriseResolvers.a8nResolver.CreateChangesets(ctx, args)
}

func (r *schemaResolver) Changesets(ctx context.Context, args *graphqlutil.ConnectionArgs) (ExternalChangesetsConnectionResolver, error) {
	if EnterpriseResolvers.a8nResolver == nil {
		return nil, a8nOnlyInEnterprise
	}
	return EnterpriseResolvers.a8nResolver.Changesets(ctx, args)
}

type ChangesetCountsArgs struct {
	From *DateTime
	To   *DateTime
}

type CampaignResolver interface {
	ID() graphql.ID
	Name() string
	Description() string
	Author(ctx context.Context) (*UserResolver, error)
	URL(ctx context.Context) (string, error)
	Namespace(ctx context.Context) (n NamespaceResolver, err error)
	CreatedAt() DateTime
	UpdatedAt() DateTime
	Changesets(ctx context.Context, args struct{ graphqlutil.ConnectionArgs }) ExternalChangesetsConnectionResolver
	ChangesetCountsOverTime(ctx context.Context, args *ChangesetCountsArgs) ([]ChangesetCountsResolver, error)
	RepositoryDiffs(ctx context.Context, args *graphqlutil.ConnectionArgs) (RepositoryComparisonConnectionResolver, error)
	Plan(ctx context.Context) (CampaignPlanResolver, error)
	ChangesetCreationStatus(context.Context) (BackgroundProcessStatus, error)
	ClosedAt() *DateTime
}

type CampaignsConnectionResolver interface {
	Nodes(ctx context.Context) ([]CampaignResolver, error)
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

type ExternalChangesetsConnectionResolver interface {
	Nodes(ctx context.Context) ([]ExternalChangesetResolver, error)
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

type ExternalChangesetResolver interface {
	ID() graphql.ID
	CreatedAt() DateTime
	UpdatedAt() DateTime
	Title() (string, error)
	Body() (string, error)
	State() (a8n.ChangesetState, error)
	ExternalURL() (*externallink.Resolver, error)
	ReviewState(context.Context) (a8n.ChangesetReviewState, error)
	Repository(ctx context.Context) (*RepositoryResolver, error)
	Campaigns(ctx context.Context, args *struct{ graphqlutil.ConnectionArgs }) (CampaignsConnectionResolver, error)
	Events(ctx context.Context, args *struct{ graphqlutil.ConnectionArgs }) (ChangesetEventsConnectionResolver, error)
	Diff(ctx context.Context) (*RepositoryComparisonResolver, error)
	Head(ctx context.Context) (*GitRefResolver, error)
	Base(ctx context.Context) (*GitRefResolver, error)
}

type ChangesetPlansConnectionResolver interface {
	Nodes(ctx context.Context) ([]ChangesetPlanResolver, error)
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

type ChangesetPlanResolver interface {
	Repository(ctx context.Context) (*RepositoryResolver, error)
	BaseRepository(ctx context.Context) (*RepositoryResolver, error)
	Diff() ChangesetPlanResolver
	FileDiffs(ctx context.Context, args *graphqlutil.ConnectionArgs) (PreviewFileDiffConnection, error)
}

type ChangesetEventsConnectionResolver interface {
	Nodes(ctx context.Context) ([]ChangesetEventResolver, error)
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

type ChangesetEventResolver interface {
	ID() graphql.ID
	Changeset(ctx context.Context) (ExternalChangesetResolver, error)
	CreatedAt() DateTime
}

type ChangesetCountsResolver interface {
	Date() DateTime
	Total() int32
	Merged() int32
	Closed() int32
	Open() int32
	OpenApproved() int32
	OpenChangesRequested() int32
	OpenPending() int32
}

type CampaignPlanArgResolver interface {
	Name() string
	Value() string
}

type CampaignPlanSpecification interface {
	Type() string
	Arguments() string
}

type BackgroundProcessStatus interface {
	CompletedCount() int32
	PendingCount() int32

	State() a8n.BackgroundProcessState

	Errors() []string
}

type CampaignPlanResolver interface {
	ID() graphql.ID

	Type() string
	Arguments() (JSONCString, error)

	Status(ctx context.Context) (BackgroundProcessStatus, error)

	Changesets(ctx context.Context, args *graphqlutil.ConnectionArgs) ChangesetPlansConnectionResolver
}

type PreviewFileDiff interface {
	OldPath() *string
	NewPath() *string
	Hunks() []*DiffHunk
	Stat() *DiffStat
	OldFile() *GitTreeEntryResolver
	InternalID() string
}

type PreviewFileDiffConnection interface {
	Nodes(ctx context.Context) ([]PreviewFileDiff, error)
	TotalCount(ctx context.Context) (*int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
	DiffStat(ctx context.Context) (*DiffStat, error)
	RawDiff(ctx context.Context) (string, error)
}
