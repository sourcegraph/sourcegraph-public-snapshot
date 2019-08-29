package graphqlbackend

import (
	"context"
	"errors"
	"fmt"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

// Campaigns is the implementation of the GraphQL campaigns queries and mutations. If it is not set
// at runtime, a "not implemented" error is returned to API clients who invoke it.
//
// This is contributed by enterprise.
var Campaigns CampaignsResolver

const GQLTypeCampaign = "Campaign"

func MarshalCampaignID(id int64) graphql.ID {
	return relay.MarshalID(GQLTypeCampaign, id)
}

func UnmarshalCampaignID(id graphql.ID) (dbID int64, err error) {
	if typ := relay.UnmarshalKind(id); typ != GQLTypeCampaign {
		return 0, fmt.Errorf("campaign ID has unexpected type type %q", typ)
	}
	err = relay.UnmarshalSpec(id, &dbID)
	return
}

var errCampaignsNotImplemented = errors.New("campaigns is not implemented")

// CampaignByID is called to look up a Campaign given its GraphQL ID.
func CampaignByID(ctx context.Context, id graphql.ID) (Campaign, error) {
	if Campaigns == nil {
		return nil, errCampaignsNotImplemented
	}
	return Campaigns.CampaignByID(ctx, id)
}

// CampaignByDBID is called to look up a Campaign given its DB ID.
func CampaignByDBID(ctx context.Context, id int64) (Campaign, error) {
	if Campaigns == nil {
		return nil, errCampaignsNotImplemented
	}
	return Campaigns.CampaignByDBID(ctx, id)
}

// CampaignsInNamespace returns an instance of the GraphQL CampaignConnection type with the list of
// campaigns defined in a namespace.
func CampaignsInNamespace(ctx context.Context, namespace graphql.ID, arg *graphqlutil.ConnectionArgs) (CampaignConnection, error) {
	if Campaigns == nil {
		return nil, errCampaignsNotImplemented
	}
	return Campaigns.CampaignsInNamespace(ctx, namespace, arg)
}

// CampaignsWithObject returns an instance of the GraphQL CampaignConnection type with the list of
// campaigns that contain the object.
func CampaignsWithObject(ctx context.Context, object graphql.ID, arg *graphqlutil.ConnectionArgs) (CampaignConnection, error) {
	if Campaigns == nil {
		return nil, errCampaignsNotImplemented
	}
	return Campaigns.CampaignsWithObject(ctx, object, arg)
}

func (schemaResolver) Campaigns(ctx context.Context, arg *CampaignsArgs) (CampaignConnection, error) {
	if Campaigns == nil {
		return nil, errCampaignsNotImplemented
	}
	return Campaigns.Campaigns(ctx, arg)
}

func (r schemaResolver) CampaignPreview(ctx context.Context, arg *CampaignPreviewArgs) (CampaignPreview, error) {
	if Campaigns == nil {
		return nil, errCampaignsNotImplemented
	}
	return Campaigns.CampaignPreview(ctx, arg)
}

func (r schemaResolver) CampaignUpdatePreview(ctx context.Context, arg *CampaignUpdatePreviewArgs) (CampaignUpdatePreview, error) {
	if Campaigns == nil {
		return nil, errCampaignsNotImplemented
	}
	return Campaigns.CampaignUpdatePreview(ctx, arg)
}

func (r schemaResolver) CreateCampaign(ctx context.Context, arg *CreateCampaignArgs) (Campaign, error) {
	if Campaigns == nil {
		return nil, errCampaignsNotImplemented
	}
	return Campaigns.CreateCampaign(ctx, arg)
}

func (r schemaResolver) UpdateCampaign(ctx context.Context, arg *UpdateCampaignArgs) (Campaign, error) {
	if Campaigns == nil {
		return nil, errCampaignsNotImplemented
	}
	return Campaigns.UpdateCampaign(ctx, arg)
}

func (r schemaResolver) PublishDraftCampaign(ctx context.Context, arg *PublishDraftCampaignArgs) (Campaign, error) {
	if Campaigns == nil {
		return nil, errCampaignsNotImplemented
	}
	return Campaigns.PublishDraftCampaign(ctx, arg)
}

func (r schemaResolver) ForceRefreshCampaign(ctx context.Context, arg *ForceRefreshCampaignArgs) (Campaign, error) {
	if Campaigns == nil {
		return nil, errCampaignsNotImplemented
	}
	return Campaigns.ForceRefreshCampaign(ctx, arg)
}

func (r schemaResolver) DeleteCampaign(ctx context.Context, arg *DeleteCampaignArgs) (*EmptyResponse, error) {
	if Campaigns == nil {
		return nil, errCampaignsNotImplemented
	}
	return Campaigns.DeleteCampaign(ctx, arg)
}

func (r schemaResolver) AddThreadsToCampaign(ctx context.Context, arg *AddRemoveThreadsToFromCampaignArgs) (*EmptyResponse, error) {
	if Campaigns == nil {
		return nil, errCampaignsNotImplemented
	}
	return Campaigns.AddThreadsToCampaign(ctx, arg)
}

func (r schemaResolver) RemoveThreadsFromCampaign(ctx context.Context, arg *AddRemoveThreadsToFromCampaignArgs) (*EmptyResponse, error) {
	if Campaigns == nil {
		return nil, errCampaignsNotImplemented
	}
	return Campaigns.RemoveThreadsFromCampaign(ctx, arg)
}

// CampaignsResolver is the interface for the GraphQL campaigns queries and mutations.
type CampaignsResolver interface {
	// Queries
	Campaigns(context.Context, *CampaignsArgs) (CampaignConnection, error)
	CampaignPreview(context.Context, *CampaignPreviewArgs) (CampaignPreview, error)
	CampaignUpdatePreview(context.Context, *CampaignUpdatePreviewArgs) (CampaignUpdatePreview, error)

	// Mutations
	CreateCampaign(context.Context, *CreateCampaignArgs) (Campaign, error)
	UpdateCampaign(context.Context, *UpdateCampaignArgs) (Campaign, error)
	PublishDraftCampaign(context.Context, *PublishDraftCampaignArgs) (Campaign, error)
	ForceRefreshCampaign(context.Context, *ForceRefreshCampaignArgs) (Campaign, error)
	DeleteCampaign(context.Context, *DeleteCampaignArgs) (*EmptyResponse, error)
	AddThreadsToCampaign(context.Context, *AddRemoveThreadsToFromCampaignArgs) (*EmptyResponse, error)
	RemoveThreadsFromCampaign(context.Context, *AddRemoveThreadsToFromCampaignArgs) (*EmptyResponse, error)

	// CampaignByID is called by the CampaignByID func but is not in the GraphQL API.
	CampaignByID(context.Context, graphql.ID) (Campaign, error)

	// CampaignByDBID is called by the CampaignByDBID func but is not in the GraphQL API.
	CampaignByDBID(context.Context, int64) (Campaign, error)

	// CampaignsInNamespace is called by the CampaignsInNamespace func but is not in the GraphQL
	// API.
	CampaignsInNamespace(ctx context.Context, namespace graphql.ID, arg *graphqlutil.ConnectionArgs) (CampaignConnection, error)

	// CampaignsWithObject is called by the CampaignsWithObject func but is not in the GraphQL API.
	CampaignsWithObject(ctx context.Context, object graphql.ID, arg *graphqlutil.ConnectionArgs) (CampaignConnection, error)
}

type CampaignsArgs struct {
	graphqlutil.ConnectionArgs
	Object *graphql.ID
}

type CampaignPreviewArgs struct {
	Input CampaignPreviewInput
}

type CampaignUpdatePreviewArgs struct {
	Input CampaignUpdatePreviewInput
}

type CampaignExtensionData struct {
	RawDiagnostics []string
	RawFileDiffs   []string
}

type CampaignPreviewInput struct {
	Campaign CreateCampaignInput
}

type CreateCampaignInput struct {
	Namespace     graphql.ID
	Name          string
	Body          *string
	Template      *CampaignTemplateInput
	Draft         *bool
	StartDate     *DateTime
	DueDate       *DateTime
	Rules         *[]NewRuleInput
	ExtensionData CampaignExtensionData
}

type CreateCampaignArgs struct {
	Input CreateCampaignInput
}

type CampaignUpdatePreviewInput struct {
	Campaign graphql.ID
	Update   UpdateCampaignInput
}

type UpdateCampaignInput struct {
	ID             graphql.ID
	Name           *string
	Body           *string
	Template       *CampaignTemplateInput
	ClearTemplate  *bool
	StartDate      *DateTime
	ClearStartDate *bool
	DueDate        *DateTime
	ClearDueDate   *bool
	Rules          *[]NewRuleInput
	ExtensionData  *CampaignExtensionData
}

type UpdateCampaignArgs struct {
	Input UpdateCampaignInput
}

type PublishDraftCampaignArgs struct {
	Campaign graphql.ID
}

type ForceRefreshCampaignArgs struct {
	Campaign      graphql.ID
	ExtensionData CampaignExtensionData
}

type DeleteCampaignArgs struct {
	Campaign graphql.ID
}

type AddRemoveThreadsToFromCampaignArgs struct {
	Campaign graphql.ID
	Threads  []graphql.ID
}

// Campaign is the interface for the GraphQL type Campaign.
type Campaign interface {
	PartialComment
	ID() graphql.ID
	Namespace(context.Context) (*NamespaceResolver, error)
	Name() string
	Template() *CampaignTemplateInstance
	Updatable
	commentable
	IsDraft() bool
	StartDate() *DateTime
	DueDate() *DateTime
	ruleContainer
	URL(context.Context) (string, error)
	Threads(context.Context, *ThreadConnectionArgs) (ThreadOrThreadPreviewConnection, error)
	Repositories(context.Context) ([]*RepositoryResolver, error)
	Commits(context.Context) ([]*GitCommitResolver, error)
	RepositoryComparisons(context.Context) ([]RepositoryComparison, error)
	Diagnostics(context.Context, *ThreadDiagnosticConnectionArgs) (ThreadDiagnosticConnection, error)
	BurndownChart(context.Context) (CampaignBurndownChart, error)
	TimelineItems(context.Context, *EventConnectionCommonArgs) (EventConnection, error)
	hasParticipants
}

// CampaignNode is the interface for the GraphQL interface CampaignNode.
type CampaignNode interface {
	Campaigns(context.Context, *graphqlutil.ConnectionArgs) (CampaignConnection, error)
}

// CampaignConnection is the interface for the GraphQL type CampaignConnection.
type CampaignConnection interface {
	Nodes(context.Context) ([]Campaign, error)
	TotalCount(context.Context) (int32, error)
	PageInfo(context.Context) (*graphqlutil.PageInfo, error)
}

type AddRemoveThreadToFromCampaignEvent struct {
	EventCommon
	Campaign_ Campaign
	Thread_   Thread
}

func (v AddRemoveThreadToFromCampaignEvent) Campaign() Campaign { return v.Campaign_ }
func (v AddRemoveThreadToFromCampaignEvent) Thread() Thread     { return v.Thread_ }

// CampaignBurndownChart is the interface for the GraphQL type CampaignBurndownChart.
type CampaignBurndownChart interface {
	Dates() []DateTime
	OpenThreads() []int32
	ClosedThreads() []int32
	MergedThreads() []int32
	TotalThreads() []int32
	OpenApprovedThreads() []int32
}
