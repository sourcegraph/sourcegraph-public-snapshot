package graphqlbackend

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

// NewCodeIntelResolver will be set by enterprise.
var NewCodeIntelResolver func() CodeIntelResolver

type CodeIntelResolver interface {
	LSIFUploadByID(ctx context.Context, id graphql.ID) (LSIFUploadResolver, error)
	LSIFUploads(ctx context.Context, args *LSIFRepositoryUploadsQueryArgs) (LSIFUploadConnectionResolver, error)
	DeleteLSIFUpload(ctx context.Context, id graphql.ID) (*EmptyResponse, error)
	LSIFIndexByID(ctx context.Context, id graphql.ID) (LSIFIndexResolver, error)
	LSIFIndexes(ctx context.Context, args *LSIFRepositoryIndexesQueryArgs) (LSIFIndexConnectionResolver, error)
	DeleteLSIFIndex(ctx context.Context, id graphql.ID) (*EmptyResponse, error)
	LSIF(ctx context.Context, args *LSIFQueryArgs) (LSIFQueryResolver, error)
}

var codeIntelOnlyInEnterprise = errors.New("lsif uploads and queries are only available in enterprise")

type defaultCodeIntelResolver struct{}

func (defaultCodeIntelResolver) LSIFUploadByID(ctx context.Context, id graphql.ID) (LSIFUploadResolver, error) {
	return nil, codeIntelOnlyInEnterprise
}

func (defaultCodeIntelResolver) LSIFUploads(ctx context.Context, args *LSIFRepositoryUploadsQueryArgs) (LSIFUploadConnectionResolver, error) {
	return nil, codeIntelOnlyInEnterprise
}

func (defaultCodeIntelResolver) DeleteLSIFUpload(ctx context.Context, id graphql.ID) (*EmptyResponse, error) {
	return nil, codeIntelOnlyInEnterprise
}

func (defaultCodeIntelResolver) LSIFIndexByID(ctx context.Context, id graphql.ID) (LSIFIndexResolver, error) {
	return nil, codeIntelOnlyInEnterprise
}

func (defaultCodeIntelResolver) LSIFIndexes(ctx context.Context, args *LSIFRepositoryIndexesQueryArgs) (LSIFIndexConnectionResolver, error) {
	return nil, codeIntelOnlyInEnterprise
}

func (defaultCodeIntelResolver) DeleteLSIFIndex(ctx context.Context, id graphql.ID) (*EmptyResponse, error) {
	return nil, codeIntelOnlyInEnterprise
}

func (defaultCodeIntelResolver) LSIF(ctx context.Context, args *LSIFQueryArgs) (LSIFQueryResolver, error) {
	return nil, codeIntelOnlyInEnterprise
}

func (r *schemaResolver) DeleteLSIFUpload(ctx context.Context, args *struct{ ID graphql.ID }) (*EmptyResponse, error) {
	// We need to override the embedded method here as it takes slightly different arguments
	return r.CodeIntelResolver.DeleteLSIFUpload(ctx, args.ID)
}

func (r *schemaResolver) DeleteLSIFIndex(ctx context.Context, args *struct{ ID graphql.ID }) (*EmptyResponse, error) {
	// We need to override the embedded method here as it takes slightly different arguments
	return r.CodeIntelResolver.DeleteLSIFIndex(ctx, args.ID)
}

type LSIFUploadsQueryArgs struct {
	graphqlutil.ConnectionArgs
	Query           *string
	State           *string
	IsLatestForRepo *bool
	After           *string
}

type LSIFRepositoryUploadsQueryArgs struct {
	*LSIFUploadsQueryArgs
	RepositoryID graphql.ID
}

type LSIFUploadResolver interface {
	ID() graphql.ID
	ProjectRoot(ctx context.Context) (*GitTreeEntryResolver, error)
	InputCommit() string
	InputRoot() string
	InputIndexer() string
	State() string
	UploadedAt() DateTime
	StartedAt() *DateTime
	FinishedAt() *DateTime
	Failure() LSIFUploadFailureReasonResolver
	IsLatestForRepo() bool
	PlaceInQueue() *int32
}

type LSIFUploadFailureReasonResolver interface {
	Summary() string
	Stacktrace() string
}

type LSIFUploadConnectionResolver interface {
	Nodes(ctx context.Context) ([]LSIFUploadResolver, error)
	TotalCount(ctx context.Context) (*int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

type LSIFIndexesQueryArgs struct {
	graphqlutil.ConnectionArgs
	Query *string
	State *string
	After *string
}

type LSIFRepositoryIndexesQueryArgs struct {
	*LSIFIndexesQueryArgs
	RepositoryID graphql.ID
}

type LSIFIndexResolver interface {
	ID() graphql.ID
	ProjectRoot(ctx context.Context) (*GitTreeEntryResolver, error)
	InputCommit() string
	State() string
	QueuedAt() DateTime
	StartedAt() *DateTime
	FinishedAt() *DateTime
	Failure() LSIFIndexFailureReasonResolver
	PlaceInQueue() *int32
}

type LSIFIndexFailureReasonResolver interface {
	Summary() string
	Stacktrace() string
}

type LSIFIndexConnectionResolver interface {
	Nodes(ctx context.Context) ([]LSIFIndexResolver, error)
	TotalCount(ctx context.Context) (*int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

type LSIFQueryResolver interface {
	Definitions(ctx context.Context, args *LSIFQueryPositionArgs) (LocationConnectionResolver, error)
	References(ctx context.Context, args *LSIFPagedQueryPositionArgs) (LocationConnectionResolver, error)
	Hover(ctx context.Context, args *LSIFQueryPositionArgs) (HoverResolver, error)
}

type LSIFQueryArgs struct {
	Repository *RepositoryResolver
	Commit     api.CommitID
	Path       string
	UploadID   int64
}

type LSIFQueryPositionArgs struct {
	Line      int32
	Character int32
}

type LSIFPagedQueryPositionArgs struct {
	LSIFQueryPositionArgs
	graphqlutil.ConnectionArgs
	After *string
}

type LocationConnectionResolver interface {
	Nodes(ctx context.Context) ([]LocationResolver, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

type HoverResolver interface {
	Markdown() MarkdownResolver
	Range() RangeResolver
}
