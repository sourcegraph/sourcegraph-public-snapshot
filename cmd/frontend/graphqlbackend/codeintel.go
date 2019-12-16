package graphqlbackend

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

// NewCodeIntelResolver will be set by enterprise
var NewCodeIntelResolver func() CodeIntelResolver

type CodeIntelResolver interface {
	LSIFDumpByID(ctx context.Context, id graphql.ID) (LSIFDumpResolver, error)
	LSIFDumps(ctx context.Context, args *LSIFRepositoryDumpsQueryArgs) (LSIFDumpConnectionResolver, error)
	LSIFUploadByID(ctx context.Context, id graphql.ID) (LSIFUploadResolver, error)
	LSIFUploads(ctx context.Context, args *LSIFUploadsQueryArgs) (LSIFUploadConnectionResolver, error)
	LSIFUploadStats(ctx context.Context) (LSIFUploadStatsResolver, error)
	LSIFUploadStatsByID(ctx context.Context, id graphql.ID) (LSIFUploadStatsResolver, error)
	DeleteLSIFDump(ctx context.Context, id graphql.ID) (*EmptyResponse, error)
	DeleteLSIFUpload(ctx context.Context, id graphql.ID) (*EmptyResponse, error)
	LSIF(ctx context.Context, args *LSIFQueryArgs) (LSIFQueryResolver, error)
}

type LSIFDumpsQueryArgs struct {
	graphqlutil.ConnectionArgs
	Query           *string
	IsLatestForRepo *bool
	After           *string
}

type LSIFRepositoryDumpsQueryArgs struct {
	*LSIFDumpsQueryArgs
	RepositoryID graphql.ID
}

type LSIFUploadsQueryArgs struct {
	graphqlutil.ConnectionArgs
	State string
	Query *string
	After *string
}

type LSIFDumpResolver interface {
	ID() graphql.ID
	ProjectRoot(ctx context.Context) (*GitTreeEntryResolver, error)
	IsLatestForRepo() bool
	UploadedAt() DateTime
	ProcessedAt() DateTime
}

type LSIFDumpConnectionResolver interface {
	Nodes(ctx context.Context) ([]LSIFDumpResolver, error)
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

type LSIFUploadStatsResolver interface {
	ID() graphql.ID
	ProcessingCount() int32
	ErroredCount() int32
	CompletedCount() int32
	QueuedCount() int32
}

type LSIFUploadResolver interface {
	ID() graphql.ID
	ProjectRoot(ctx context.Context) (*GitTreeEntryResolver, error)
	State() string
	Failure() LSIFUploadFailureReasonResolver
	UploadedAt() DateTime
	StartedAt() *DateTime
	FinishedAt() *DateTime
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

type LSIFQueryResolver interface {
	Commit(ctx context.Context) (*GitCommitResolver, error)
	Definitions(ctx context.Context, args *LSIFQueryPositionArgs) (LocationConnectionResolver, error)
	References(ctx context.Context, args *LSIFPagedQueryPositionArgs) (LocationConnectionResolver, error)
	Hover(ctx context.Context, args *LSIFQueryPositionArgs) (HoverResolver, error)
}

type LSIFQueryArgs struct {
	RepoName string
	Commit   GitObjectID
	Path     string
	DumpID   int64
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

var codeIntelOnlyInEnterprise = errors.New("lsif dumps, uploads, and queries are only available in enterprise")

func (r *schemaResolver) LSIFUploads(ctx context.Context, args *LSIFUploadsQueryArgs) (LSIFUploadConnectionResolver, error) {
	if EnterpriseResolvers.codeIntelResolver == nil {
		return nil, codeIntelOnlyInEnterprise
	}
	return EnterpriseResolvers.codeIntelResolver.LSIFUploads(ctx, args)
}

func (r *schemaResolver) LSIFUploadStats(ctx context.Context) (LSIFUploadStatsResolver, error) {
	if EnterpriseResolvers.codeIntelResolver == nil {
		return nil, codeIntelOnlyInEnterprise
	}
	return EnterpriseResolvers.codeIntelResolver.LSIFUploadStats(ctx)
}

func (r *schemaResolver) DeleteLSIFDump(ctx context.Context, args *struct{ ID graphql.ID }) (*EmptyResponse, error) {
	if EnterpriseResolvers.codeIntelResolver == nil {
		return nil, codeIntelOnlyInEnterprise
	}
	return EnterpriseResolvers.codeIntelResolver.DeleteLSIFDump(ctx, args.ID)
}

func (r *schemaResolver) DeleteLSIFUpload(ctx context.Context, args *struct{ ID graphql.ID }) (*EmptyResponse, error) {
	if EnterpriseResolvers.codeIntelResolver == nil {
		return nil, codeIntelOnlyInEnterprise
	}
	return EnterpriseResolvers.codeIntelResolver.DeleteLSIFUpload(ctx, args.ID)
}
