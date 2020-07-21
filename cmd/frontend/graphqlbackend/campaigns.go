package graphqlbackend

import (
	"context"
	"errors"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/externallink"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
)

type AddChangesetsToCampaignArgs struct {
	Campaign   graphql.ID
	Changesets []graphql.ID
}

type CreateCampaignArgs struct {
	Input struct {
		Namespace   graphql.ID
		Name        string
		Description *string
		Branch      *string
		PatchSet    *graphql.ID
	}
}

type UpdateCampaignArgs struct {
	Input struct {
		ID          graphql.ID
		Name        *string
		Description *string
		Branch      *string
		PatchSet    *graphql.ID
	}
}

type CreatePatchSetFromPatchesArgs struct {
	Patches []PatchInput
}

type PatchInput struct {
	Repository   graphql.ID
	BaseRevision api.CommitID
	BaseRef      string
	Patch        string
}

type ListCampaignArgs struct {
	First               *int32
	State               *string
	HasPatchSet         *bool
	ViewerCanAdminister *bool
}

type DeleteCampaignArgs struct {
	Campaign        graphql.ID
	CloseChangesets bool
}

type RetryCampaignChangesetsArgs struct {
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

type PublishCampaignChangesetsArgs struct {
	Campaign graphql.ID
}

type PublishChangesetArgs struct {
	Patch graphql.ID
}

type SyncChangesetArgs struct {
	Changeset graphql.ID
}

type FileDiffsConnectionArgs struct {
	First *int32
	After *string
}

type CampaignsResolver interface {
	CreateCampaign(ctx context.Context, args *CreateCampaignArgs) (CampaignResolver, error)
	UpdateCampaign(ctx context.Context, args *UpdateCampaignArgs) (CampaignResolver, error)
	CampaignByID(ctx context.Context, id graphql.ID) (CampaignResolver, error)
	Campaigns(ctx context.Context, args *ListCampaignArgs) (CampaignsConnectionResolver, error)
	DeleteCampaign(ctx context.Context, args *DeleteCampaignArgs) (*EmptyResponse, error)
	RetryCampaignChangesets(ctx context.Context, args *RetryCampaignChangesetsArgs) (CampaignResolver, error)
	CloseCampaign(ctx context.Context, args *CloseCampaignArgs) (CampaignResolver, error)
	PublishCampaignChangesets(ctx context.Context, args *PublishCampaignChangesetsArgs) (CampaignResolver, error)
	PublishChangeset(ctx context.Context, args *PublishChangesetArgs) (*EmptyResponse, error)
	SyncChangeset(ctx context.Context, args *SyncChangesetArgs) (*EmptyResponse, error)

	CreateChangesets(ctx context.Context, args *CreateChangesetsArgs) ([]ExternalChangesetResolver, error)
	ChangesetByID(ctx context.Context, id graphql.ID) (ChangesetResolver, error)

	AddChangesetsToCampaign(ctx context.Context, args *AddChangesetsToCampaignArgs) (CampaignResolver, error)

	CreatePatchSetFromPatches(ctx context.Context, args CreatePatchSetFromPatchesArgs) (PatchSetResolver, error)
	PatchSetByID(ctx context.Context, id graphql.ID) (PatchSetResolver, error)

	PatchByID(ctx context.Context, id graphql.ID) (PatchInterfaceResolver, error)
}

var campaignsOnlyInEnterprise = errors.New("campaigns and changesets are only available in enterprise")

type defaultCampaignsResolver struct{}

var DefaultCampaignsResolver CampaignsResolver = defaultCampaignsResolver{}

func (defaultCampaignsResolver) CreateCampaign(ctx context.Context, args *CreateCampaignArgs) (CampaignResolver, error) {
	return nil, campaignsOnlyInEnterprise
}

func (defaultCampaignsResolver) UpdateCampaign(ctx context.Context, args *UpdateCampaignArgs) (CampaignResolver, error) {
	return nil, campaignsOnlyInEnterprise
}

func (defaultCampaignsResolver) CampaignByID(ctx context.Context, id graphql.ID) (CampaignResolver, error) {
	return nil, campaignsOnlyInEnterprise
}

func (defaultCampaignsResolver) Campaigns(ctx context.Context, args *ListCampaignArgs) (CampaignsConnectionResolver, error) {
	return nil, campaignsOnlyInEnterprise
}

func (defaultCampaignsResolver) DeleteCampaign(ctx context.Context, args *DeleteCampaignArgs) (*EmptyResponse, error) {
	return nil, campaignsOnlyInEnterprise
}

func (defaultCampaignsResolver) RetryCampaignChangesets(ctx context.Context, args *RetryCampaignChangesetsArgs) (CampaignResolver, error) {
	return nil, campaignsOnlyInEnterprise
}

func (defaultCampaignsResolver) CloseCampaign(ctx context.Context, args *CloseCampaignArgs) (CampaignResolver, error) {
	return nil, campaignsOnlyInEnterprise
}

func (defaultCampaignsResolver) PublishCampaignChangesets(ctx context.Context, args *PublishCampaignChangesetsArgs) (CampaignResolver, error) {
	return nil, campaignsOnlyInEnterprise
}

func (defaultCampaignsResolver) PublishChangeset(ctx context.Context, args *PublishChangesetArgs) (*EmptyResponse, error) {
	return nil, campaignsOnlyInEnterprise
}

func (defaultCampaignsResolver) SyncChangeset(ctx context.Context, args *SyncChangesetArgs) (*EmptyResponse, error) {
	return nil, campaignsOnlyInEnterprise
}

func (defaultCampaignsResolver) CreateChangesets(ctx context.Context, args *CreateChangesetsArgs) ([]ExternalChangesetResolver, error) {
	return nil, campaignsOnlyInEnterprise
}

func (defaultCampaignsResolver) ChangesetByID(ctx context.Context, id graphql.ID) (ChangesetResolver, error) {
	return nil, campaignsOnlyInEnterprise
}

func (defaultCampaignsResolver) AddChangesetsToCampaign(ctx context.Context, args *AddChangesetsToCampaignArgs) (CampaignResolver, error) {
	return nil, campaignsOnlyInEnterprise
}

func (defaultCampaignsResolver) CreatePatchSetFromPatches(ctx context.Context, args CreatePatchSetFromPatchesArgs) (PatchSetResolver, error) {
	return nil, campaignsOnlyInEnterprise
}

func (defaultCampaignsResolver) PatchSetByID(ctx context.Context, id graphql.ID) (PatchSetResolver, error) {
	return nil, campaignsOnlyInEnterprise
}

func (defaultCampaignsResolver) PatchByID(ctx context.Context, id graphql.ID) (PatchInterfaceResolver, error) {
	return nil, campaignsOnlyInEnterprise
}

type ChangesetCountsArgs struct {
	From *DateTime
	To   *DateTime
}

type ListChangesetsArgs struct {
	First       *int32
	State       *campaigns.ChangesetState
	ReviewState *campaigns.ChangesetReviewState
	CheckState  *campaigns.ChangesetCheckState
}

type CampaignResolver interface {
	ID() graphql.ID
	Name() string
	Description() *string
	Branch() *string
	Author(ctx context.Context) (*UserResolver, error)
	ViewerCanAdminister(ctx context.Context) (bool, error)
	URL(ctx context.Context) (string, error)
	Namespace(ctx context.Context) (n NamespaceResolver, err error)
	CreatedAt() DateTime
	UpdatedAt() DateTime
	Changesets(ctx context.Context, args *ListChangesetsArgs) (ChangesetsConnectionResolver, error)
	OpenChangesets(ctx context.Context) (ChangesetsConnectionResolver, error)
	ChangesetCountsOverTime(ctx context.Context, args *ChangesetCountsArgs) ([]ChangesetCountsResolver, error)
	RepositoryDiffs(ctx context.Context, args *graphqlutil.ConnectionArgs) (RepositoryComparisonConnectionResolver, error)
	PatchSet(ctx context.Context) (PatchSetResolver, error)
	Status(context.Context) (BackgroundProcessStatus, error)
	ClosedAt() *DateTime
	Patches(ctx context.Context, args *graphqlutil.ConnectionArgs) PatchConnectionResolver
	HasUnpublishedPatches(ctx context.Context) (bool, error)
	DiffStat(ctx context.Context) (*DiffStat, error)
}

type CampaignsConnectionResolver interface {
	Nodes(ctx context.Context) ([]CampaignResolver, error)
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

type ChangesetsConnectionResolver interface {
	Nodes(ctx context.Context) ([]ChangesetResolver, error)
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

type ChangesetLabelResolver interface {
	Text() string
	Color() string
	Description() *string
}

// ChangesetResolver is the "interface Changeset" in the GraphQL schema and is
// implemented by ExternalChangesetResolver and HiddenExternalChangesetResolver.
type ChangesetResolver interface {
	ID() graphql.ID

	CreatedAt() DateTime
	UpdatedAt() DateTime
	NextSyncAt(ctx context.Context) (*DateTime, error)
	State() campaigns.ChangesetState
	Campaigns(ctx context.Context, args *ListCampaignArgs) (CampaignsConnectionResolver, error)

	ToExternalChangeset() (ExternalChangesetResolver, bool)
	ToHiddenExternalChangeset() (HiddenExternalChangesetResolver, bool)
}

// HiddenExternalChangesetResolver implements only the common interface,
// ChangesetResolver, to not reveal information to unauthorized users.
//
// Theoretically this type is not necessary, but it's easier to understand the
// implementation of the GraphQL schema if we have a mapping between GraphQL
// types and Go types.
type HiddenExternalChangesetResolver interface {
	ChangesetResolver
}

// ExternalChangesetResolver implements the ChangesetResolver interface and
// additional data.
type ExternalChangesetResolver interface {
	ChangesetResolver

	ExternalID() string
	Title() (string, error)
	Body() (string, error)
	ExternalURL() (*externallink.Resolver, error)
	ReviewState(context.Context) campaigns.ChangesetReviewState
	CheckState() *campaigns.ChangesetCheckState
	Repository(ctx context.Context) (*RepositoryResolver, error)

	Events(ctx context.Context, args *struct{ graphqlutil.ConnectionArgs }) (ChangesetEventsConnectionResolver, error)
	Diff(ctx context.Context) (*RepositoryComparisonResolver, error)
	DiffStat(ctx context.Context) (*DiffStat, error)
	Head(ctx context.Context) (*GitRefResolver, error)
	Base(ctx context.Context) (*GitRefResolver, error)
	Labels(ctx context.Context) ([]ChangesetLabelResolver, error)
}

type PatchConnectionResolver interface {
	Nodes(ctx context.Context) ([]PatchInterfaceResolver, error)
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

// PatchInterfaceResolver is the "interface PatchInterface" in the GraphQL
// schema and is implemented by Patch and HiddenPatch.
type PatchInterfaceResolver interface {
	ID() graphql.ID

	ToPatch() (PatchResolver, bool)
	ToHiddenPatch() (HiddenPatchResolver, bool)
}

// HiddenPatchResolver implements only the common interface,
// PatchInterfaceResolver, to not reveal information to unauthorized users.
//
// Theoretically this type is not necessary, but it's easier to understand the
// implementation of the GraphQL schema if we have a mapping between GraphQL
// types and Go types.
type HiddenPatchResolver interface {
	PatchInterfaceResolver
}

// PatchResolver implements the PatchInterfaceResolver interface and additional
// data.
type PatchResolver interface {
	PatchInterfaceResolver

	Repository(ctx context.Context) (*RepositoryResolver, error)
	BaseRepository(ctx context.Context) (*RepositoryResolver, error)
	Diff() PatchResolver
	FileDiffs(ctx context.Context, args *FileDiffsConnectionArgs) (FileDiffConnection, error)
	PublicationEnqueued(ctx context.Context) (bool, error)
	Publishable(ctx context.Context) (bool, error)
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

type BackgroundProcessStatus interface {
	CompletedCount() int32
	PendingCount() int32

	State() campaigns.BackgroundProcessState

	Errors() []string
}

type PatchSetResolver interface {
	ID() graphql.ID

	Patches(ctx context.Context, args *graphqlutil.ConnectionArgs) PatchConnectionResolver

	PreviewURL() string
	DiffStat(ctx context.Context) (*DiffStat, error)
}
