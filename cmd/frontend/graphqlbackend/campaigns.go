package graphqlbackend

import (
	"context"
	"errors"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/externallink"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
)

type CreateCampaignArgs struct {
	CampaignSpec graphql.ID
}

type CreateBatchChangeArgs struct {
	BatchSpec graphql.ID
}

type ApplyCampaignArgs struct {
	CampaignSpec   graphql.ID
	EnsureCampaign *graphql.ID
}

type MoveCampaignArgs struct {
	Campaign     graphql.ID
	NewName      *string
	NewNamespace *graphql.ID
}

type ListCampaignsArgs struct {
	First               int32
	After               *string
	State               *string
	ViewerCanAdminister *bool

	Namespace *graphql.ID
}

type CloseCampaignArgs struct {
	Campaign        graphql.ID
	CloseChangesets bool
}

type DeleteCampaignArgs struct {
	Campaign graphql.ID
}

type SyncChangesetArgs struct {
	Changeset graphql.ID
}

type ReenqueueChangesetArgs struct {
	Changeset graphql.ID
}

type CreateChangesetSpecArgs struct {
	ChangesetSpec string
}

type CreateCampaignSpecArgs struct {
	Namespace graphql.ID

	CampaignSpec   string
	ChangesetSpecs []graphql.ID
}

type ChangesetSpecsConnectionArgs struct {
	First int32
	After *string
}

type ChangesetApplyPreviewConnectionArgs struct {
	First        int32
	After        *string
	Search       *string
	CurrentState *campaigns.ChangesetState
	Action       *campaigns.ReconcilerOperation
}

type BatchChangeArgs struct {
	Namespace string
	Name      string
}

type ChangesetEventsConnectionArgs struct {
	First int32
	After *string
}

type CreateCampaignsCredentialArgs struct {
	ExternalServiceKind string
	ExternalServiceURL  string
	User                graphql.ID
	Credential          string
}

type DeleteCampaignsCredentialArgs struct {
	CampaignsCredential graphql.ID
}

type ListCampaignsCodeHostsArgs struct {
	First  int32
	After  *string
	UserID int32
}

type ListViewerCampaignsCodeHostsArgs struct {
	First                 int32
	After                 *string
	OnlyWithoutCredential bool
}

type CampaignsResolver interface {
	//
	// MUTATIONS
	//
	// Deprecated:
	CreateCampaign(ctx context.Context, args *CreateCampaignArgs) (BatchChangeResolver, error)
	// TODO: To-be-deprecated (these need to be marked as deprecated and use
	// the code for the new implementations) and then moved up to "Deprecated:"
	ApplyCampaign(ctx context.Context, args *ApplyCampaignArgs) (BatchChangeResolver, error)
	MoveCampaign(ctx context.Context, args *MoveCampaignArgs) (BatchChangeResolver, error)
	CloseCampaign(ctx context.Context, args *CloseCampaignArgs) (BatchChangeResolver, error)
	DeleteCampaign(ctx context.Context, args *DeleteCampaignArgs) (*EmptyResponse, error)
	CreateCampaignSpec(ctx context.Context, args *CreateCampaignSpecArgs) (CampaignSpecResolver, error)
	CreateCampaignsCredential(ctx context.Context, args *CreateCampaignsCredentialArgs) (CampaignsCredentialResolver, error)
	DeleteCampaignsCredential(ctx context.Context, args *DeleteCampaignsCredentialArgs) (*EmptyResponse, error)
	// New:
	CreateBatchChange(ctx context.Context, args *CreateBatchChangeArgs) (BatchChangeResolver, error)

	CreateChangesetSpec(ctx context.Context, args *CreateChangesetSpecArgs) (ChangesetSpecResolver, error)
	SyncChangeset(ctx context.Context, args *SyncChangesetArgs) (*EmptyResponse, error)
	ReenqueueChangeset(ctx context.Context, args *ReenqueueChangesetArgs) (ChangesetResolver, error)

	// Queries

	// Deprecated:
	Campaign(ctx context.Context, args *BatchChangeArgs) (BatchChangeResolver, error)
	// TODO: To-be-deprecated (these need to be marked as deprecated and use
	// the code for the new implementations) and then moved up to "Deprecated:"
	Campaigns(ctx context.Context, args *ListCampaignsArgs) (CampaignsConnectionResolver, error)
	CampaignsCredentialByID(ctx context.Context, id graphql.ID) (CampaignsCredentialResolver, error)
	CampaignsCodeHosts(ctx context.Context, args *ListCampaignsCodeHostsArgs) (CampaignsCodeHostConnectionResolver, error)
	CampaignSpecByID(ctx context.Context, id graphql.ID) (CampaignSpecResolver, error)
	// New:
	BatchChange(ctx context.Context, args *BatchChangeArgs) (BatchChangeResolver, error)
	BatchChangeByID(ctx context.Context, id graphql.ID) (BatchChangeResolver, error)

	ChangesetByID(ctx context.Context, id graphql.ID) (ChangesetResolver, error)
	ChangesetSpecByID(ctx context.Context, id graphql.ID) (ChangesetSpecResolver, error)
}

type CampaignSpecResolver interface {
	ID() graphql.ID

	OriginalInput() (string, error)
	ParsedInput() (JSONValue, error)
	ChangesetSpecs(ctx context.Context, args *ChangesetSpecsConnectionArgs) (ChangesetSpecConnectionResolver, error)
	ApplyPreview(ctx context.Context, args *ChangesetApplyPreviewConnectionArgs) (ChangesetApplyPreviewConnectionResolver, error)

	Description() CampaignDescriptionResolver

	Creator(context.Context) (*UserResolver, error)
	CreatedAt() DateTime
	Namespace(context.Context) (*NamespaceResolver, error)

	ExpiresAt() *DateTime

	ApplyURL(ctx context.Context) (string, error)

	ViewerCanAdminister(context.Context) (bool, error)

	DiffStat(ctx context.Context) (*DiffStat, error)

	AppliesToCampaign(ctx context.Context) (BatchChangeResolver, error)

	SupersedingCampaignSpec(context.Context) (CampaignSpecResolver, error)

	ViewerCampaignsCodeHosts(ctx context.Context, args *ListViewerCampaignsCodeHostsArgs) (CampaignsCodeHostConnectionResolver, error)
}

type CampaignDescriptionResolver interface {
	Name() string
	Description() string
}

type ChangesetApplyPreviewResolver interface {
	ToVisibleChangesetApplyPreview() (VisibleChangesetApplyPreviewResolver, bool)
	ToHiddenChangesetApplyPreview() (HiddenChangesetApplyPreviewResolver, bool)
}

type VisibleChangesetApplyPreviewResolver interface {
	Operations(ctx context.Context) ([]campaigns.ReconcilerOperation, error)
	Delta(ctx context.Context) (ChangesetSpecDeltaResolver, error)
	Targets() VisibleApplyPreviewTargetsResolver
}

type HiddenChangesetApplyPreviewResolver interface {
	Operations(ctx context.Context) ([]campaigns.ReconcilerOperation, error)
	Delta(ctx context.Context) (ChangesetSpecDeltaResolver, error)
	Targets() HiddenApplyPreviewTargetsResolver
}

type VisibleApplyPreviewTargetsResolver interface {
	ToVisibleApplyPreviewTargetsAttach() (VisibleApplyPreviewTargetsAttachResolver, bool)
	ToVisibleApplyPreviewTargetsUpdate() (VisibleApplyPreviewTargetsUpdateResolver, bool)
	ToVisibleApplyPreviewTargetsDetach() (VisibleApplyPreviewTargetsDetachResolver, bool)
}

type VisibleApplyPreviewTargetsAttachResolver interface {
	ChangesetSpec(ctx context.Context) (VisibleChangesetSpecResolver, error)
}
type VisibleApplyPreviewTargetsUpdateResolver interface {
	ChangesetSpec(ctx context.Context) (VisibleChangesetSpecResolver, error)
	Changeset(ctx context.Context) (ExternalChangesetResolver, error)
}
type VisibleApplyPreviewTargetsDetachResolver interface {
	Changeset(ctx context.Context) (ExternalChangesetResolver, error)
}

type HiddenApplyPreviewTargetsResolver interface {
	ToHiddenApplyPreviewTargetsAttach() (HiddenApplyPreviewTargetsAttachResolver, bool)
	ToHiddenApplyPreviewTargetsUpdate() (HiddenApplyPreviewTargetsUpdateResolver, bool)
	ToHiddenApplyPreviewTargetsDetach() (HiddenApplyPreviewTargetsDetachResolver, bool)
}

type HiddenApplyPreviewTargetsAttachResolver interface {
	ChangesetSpec(ctx context.Context) (HiddenChangesetSpecResolver, error)
}
type HiddenApplyPreviewTargetsUpdateResolver interface {
	ChangesetSpec(ctx context.Context) (HiddenChangesetSpecResolver, error)
	Changeset(ctx context.Context) (HiddenExternalChangesetResolver, error)
}
type HiddenApplyPreviewTargetsDetachResolver interface {
	Changeset(ctx context.Context) (HiddenExternalChangesetResolver, error)
}

type ChangesetApplyPreviewConnectionStatsResolver interface {
	Push() int32
	Update() int32
	Undraft() int32
	Publish() int32
	PublishDraft() int32
	Sync() int32
	Import() int32
	Close() int32
	Reopen() int32
	Sleep() int32
	Detach() int32

	Added() int32
	Modified() int32
	Removed() int32
}

type ChangesetApplyPreviewConnectionResolver interface {
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
	Nodes(ctx context.Context) ([]ChangesetApplyPreviewResolver, error)
	Stats(ctx context.Context) (ChangesetApplyPreviewConnectionStatsResolver, error)
}

type ChangesetSpecConnectionResolver interface {
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
	Nodes(ctx context.Context) ([]ChangesetSpecResolver, error)
}

type ChangesetSpecResolver interface {
	ID() graphql.ID
	Type() campaigns.ChangesetSpecDescriptionType
	ExpiresAt() *DateTime

	ToHiddenChangesetSpec() (HiddenChangesetSpecResolver, bool)
	ToVisibleChangesetSpec() (VisibleChangesetSpecResolver, bool)
}

type HiddenChangesetSpecResolver interface {
	ChangesetSpecResolver
}

type VisibleChangesetSpecResolver interface {
	ChangesetSpecResolver

	Description(ctx context.Context) (ChangesetDescription, error)
}

type ChangesetSpecDeltaResolver interface {
	TitleChanged() bool
	BodyChanged() bool
	Undraft() bool
	BaseRefChanged() bool
	DiffChanged() bool
	CommitMessageChanged() bool
	AuthorNameChanged() bool
	AuthorEmailChanged() bool
}

type ChangesetDescription interface {
	ToExistingChangesetReference() (ExistingChangesetReferenceResolver, bool)
	ToGitBranchChangesetDescription() (GitBranchChangesetDescriptionResolver, bool)
}

type ExistingChangesetReferenceResolver interface {
	BaseRepository() *RepositoryResolver
	ExternalID() string
}

type GitBranchChangesetDescriptionResolver interface {
	BaseRepository() *RepositoryResolver
	BaseRef() string
	BaseRev() string

	HeadRepository() *RepositoryResolver
	HeadRef() string

	Title() string
	Body() string

	Diff(ctx context.Context) (PreviewRepositoryComparisonResolver, error)
	DiffStat() *DiffStat

	Commits() []GitCommitDescriptionResolver

	Published() campaigns.PublishedValue
}

type GitCommitDescriptionResolver interface {
	Message() string
	Subject() string
	Body() *string
	Author() *PersonResolver
	Diff() string
}

type CampaignsCodeHostConnectionResolver interface {
	Nodes(ctx context.Context) ([]CampaignsCodeHostResolver, error)
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

type CampaignsCodeHostResolver interface {
	ExternalServiceKind() string
	ExternalServiceURL() string
	RequiresSSH() bool
	Credential() CampaignsCredentialResolver
}

type CampaignsCredentialResolver interface {
	ID() graphql.ID
	ExternalServiceKind() string
	ExternalServiceURL() string
	SSHPublicKey() *string
	CreatedAt() DateTime
}

type ChangesetCountsArgs struct {
	From *DateTime
	To   *DateTime
}

type ListChangesetsArgs struct {
	First            int32
	After            *string
	PublicationState *campaigns.ChangesetPublicationState
	ReconcilerState  *[]campaigns.ReconcilerState
	ExternalState    *campaigns.ChangesetExternalState
	State            *campaigns.ChangesetState
	ReviewState      *campaigns.ChangesetReviewState
	CheckState       *campaigns.ChangesetCheckState
	// old
	OnlyPublishedByThisCampaign *bool
	//new
	OnlyPublishedByThisBatchChange *bool
	Search                         *string
}

type BatchChangeResolver interface {
	ID() graphql.ID
	Name() string
	Description() *string
	InitialApplier(ctx context.Context) (*UserResolver, error)
	LastApplier(ctx context.Context) (*UserResolver, error)
	LastAppliedAt() DateTime
	SpecCreator(ctx context.Context) (*UserResolver, error)
	ViewerCanAdminister(ctx context.Context) (bool, error)
	URL(ctx context.Context) (string, error)
	Namespace(ctx context.Context) (n NamespaceResolver, err error)
	CreatedAt() DateTime
	UpdatedAt() DateTime
	ChangesetsStats(ctx context.Context) (ChangesetsStatsResolver, error)
	Changesets(ctx context.Context, args *ListChangesetsArgs) (ChangesetsConnectionResolver, error)
	ChangesetCountsOverTime(ctx context.Context, args *ChangesetCountsArgs) ([]ChangesetCountsResolver, error)
	ClosedAt() *DateTime
	DiffStat(ctx context.Context) (*DiffStat, error)
	CurrentSpec(ctx context.Context) (CampaignSpecResolver, error)

	// NOTE: This should be removed once we remove campaigns.
	// It's here so that in the NodeResolver we can have the same resolver,
	// BatchChangeResolver, act as a Campaign and a BatchChange.
	ActAsCampaign() bool
}

type CampaignsConnectionResolver interface {
	Nodes(ctx context.Context) ([]BatchChangeResolver, error)
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

type ChangesetsStatsResolver interface {
	Retrying() int32
	Failed() int32
	Processing() int32
	Unpublished() int32
	Draft() int32
	Open() int32
	Merged() int32
	Closed() int32
	Deleted() int32
	Total() int32
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
	PublicationState() campaigns.ChangesetPublicationState
	ReconcilerState() campaigns.ReconcilerState
	ExternalState() *campaigns.ChangesetExternalState
	State() (campaigns.ChangesetState, error)
	Campaigns(ctx context.Context, args *ListCampaignsArgs) (CampaignsConnectionResolver, error)

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

	ExternalID() *string
	Title(context.Context) (*string, error)
	Body(context.Context) (*string, error)
	Author() (*PersonResolver, error)
	ExternalURL() (*externallink.Resolver, error)
	ReviewState(context.Context) *campaigns.ChangesetReviewState
	CheckState() *campaigns.ChangesetCheckState
	Repository(ctx context.Context) *RepositoryResolver

	Events(ctx context.Context, args *ChangesetEventsConnectionArgs) (ChangesetEventsConnectionResolver, error)
	Diff(ctx context.Context) (RepositoryComparisonInterface, error)
	DiffStat(ctx context.Context) (*DiffStat, error)
	Labels(ctx context.Context) ([]ChangesetLabelResolver, error)

	Error() *string
	SyncerError() *string

	CurrentSpec(ctx context.Context) (VisibleChangesetSpecResolver, error)
}

type ChangesetEventsConnectionResolver interface {
	Nodes(ctx context.Context) ([]ChangesetEventResolver, error)
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

type ChangesetEventResolver interface {
	ID() graphql.ID
	Changeset() ExternalChangesetResolver
	CreatedAt() DateTime
}

type ChangesetCountsResolver interface {
	Date() DateTime
	Total() int32
	Merged() int32
	Closed() int32
	Draft() int32
	Open() int32
	OpenApproved() int32
	OpenChangesRequested() int32
	OpenPending() int32
}

var campaignsOnlyInEnterprise = errors.New("campaigns and changesets are only available in enterprise")

type defaultCampaignsResolver struct{}

var DefaultCampaignsResolver CampaignsResolver = defaultCampaignsResolver{}

// Mutations
func (defaultCampaignsResolver) CreateCampaign(ctx context.Context, args *CreateCampaignArgs) (BatchChangeResolver, error) {
	return nil, campaignsOnlyInEnterprise
}

func (defaultCampaignsResolver) CreateBatchChange(ctx context.Context, args *CreateBatchChangeArgs) (BatchChangeResolver, error) {
	return nil, campaignsOnlyInEnterprise
}

func (defaultCampaignsResolver) ApplyCampaign(ctx context.Context, args *ApplyCampaignArgs) (BatchChangeResolver, error) {
	return nil, campaignsOnlyInEnterprise
}

func (defaultCampaignsResolver) CreateChangesetSpec(ctx context.Context, args *CreateChangesetSpecArgs) (ChangesetSpecResolver, error) {
	return nil, campaignsOnlyInEnterprise
}

func (defaultCampaignsResolver) CreateCampaignSpec(ctx context.Context, args *CreateCampaignSpecArgs) (CampaignSpecResolver, error) {
	return nil, campaignsOnlyInEnterprise
}

func (defaultCampaignsResolver) MoveCampaign(ctx context.Context, args *MoveCampaignArgs) (BatchChangeResolver, error) {
	return nil, campaignsOnlyInEnterprise
}

func (defaultCampaignsResolver) CloseCampaign(ctx context.Context, args *CloseCampaignArgs) (BatchChangeResolver, error) {
	return nil, campaignsOnlyInEnterprise
}

func (defaultCampaignsResolver) SyncChangeset(ctx context.Context, args *SyncChangesetArgs) (*EmptyResponse, error) {
	return nil, campaignsOnlyInEnterprise
}

func (defaultCampaignsResolver) ReenqueueChangeset(ctx context.Context, args *ReenqueueChangesetArgs) (ChangesetResolver, error) {
	return nil, campaignsOnlyInEnterprise
}

func (defaultCampaignsResolver) DeleteCampaign(ctx context.Context, args *DeleteCampaignArgs) (*EmptyResponse, error) {
	return nil, campaignsOnlyInEnterprise
}

func (defaultCampaignsResolver) CreateCampaignsCredential(ctx context.Context, args *CreateCampaignsCredentialArgs) (CampaignsCredentialResolver, error) {
	return nil, campaignsOnlyInEnterprise
}

func (defaultCampaignsResolver) DeleteCampaignsCredential(ctx context.Context, args *DeleteCampaignsCredentialArgs) (*EmptyResponse, error) {
	return nil, campaignsOnlyInEnterprise
}

// Queries
func (defaultCampaignsResolver) BatchChangeByID(ctx context.Context, id graphql.ID) (BatchChangeResolver, error) {
	return nil, campaignsOnlyInEnterprise
}

func (defaultCampaignsResolver) Campaign(ctx context.Context, args *BatchChangeArgs) (BatchChangeResolver, error) {
	return nil, campaignsOnlyInEnterprise
}

func (defaultCampaignsResolver) BatchChange(ctx context.Context, args *BatchChangeArgs) (BatchChangeResolver, error) {
	return nil, campaignsOnlyInEnterprise
}

func (defaultCampaignsResolver) Campaigns(ctx context.Context, args *ListCampaignsArgs) (CampaignsConnectionResolver, error) {
	return nil, campaignsOnlyInEnterprise
}

func (defaultCampaignsResolver) ChangesetByID(ctx context.Context, id graphql.ID) (ChangesetResolver, error) {
	return nil, campaignsOnlyInEnterprise
}

func (defaultCampaignsResolver) CampaignSpecByID(ctx context.Context, id graphql.ID) (CampaignSpecResolver, error) {
	return nil, campaignsOnlyInEnterprise
}

func (defaultCampaignsResolver) ChangesetSpecByID(ctx context.Context, id graphql.ID) (ChangesetSpecResolver, error) {
	return nil, campaignsOnlyInEnterprise
}

func (defaultCampaignsResolver) CampaignsCredentialByID(ctx context.Context, id graphql.ID) (CampaignsCredentialResolver, error) {
	return nil, campaignsOnlyInEnterprise
}

func (defaultCampaignsResolver) CampaignsCodeHosts(ctx context.Context, args *ListCampaignsCodeHostsArgs) (CampaignsCodeHostConnectionResolver, error) {
	return nil, campaignsOnlyInEnterprise
}
