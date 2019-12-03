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
	LSIFJobByID(ctx context.Context, id graphql.ID) (LSIFJobResolver, error)
	LSIFJobs(ctx context.Context, args *LSIFJobsQueryArgs) (LSIFJobConnectionResolver, error)
	LSIFJobStats(ctx context.Context) (LSIFJobStatsResolver, error)
	LSIFJobStatsByID(ctx context.Context, id graphql.ID) (LSIFJobStatsResolver, error)
	DeleteLSIFDump(ctx context.Context, id graphql.ID) (*EmptyResponse, error)
	DeleteLSIFJob(ctx context.Context, id graphql.ID) (*EmptyResponse, error)
	Definitions(ctx context.Context, args *LSIFFilePositionArgs) (LocationWithConfidenceConnectionResolver, error)
	References(ctx context.Context, args *LSIFPagedFilePositionArgs) (LocationWithConfidenceConnectionResolver, error)
	Hover(ctx context.Context, args *LSIFFilePositionArgs) (MarkdownWithConfidenceResolver, error)
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

type LSIFJobsQueryArgs struct {
	graphqlutil.ConnectionArgs
	State string
	Query *string
	After *string
}

type LSIFFilePositionArgs struct {
	Repository string
	Commit     string
	Path       string
	Line       int32
	Character  int32
}

type LSIFPagedFilePositionArgs struct {
	LSIFFilePositionArgs
	graphqlutil.ConnectionArgs
	After *string
}

type LSIFDumpResolver interface {
	ID() graphql.ID
	ProjectRoot(ctx context.Context) (*GitTreeEntryResolver, error)
	InputRevision() *GitObjectID
	IsLatestForRepo() bool
	UploadedAt() DateTime
	ProcessedAt() DateTime
}

type LSIFDumpConnectionResolver interface {
	Nodes(ctx context.Context) ([]LSIFDumpResolver, error)
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

type LocationWithConfidenceConnectionResolver interface {
	Nodes(ctx context.Context) ([]LocationWithConfidenceResolver, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

type LocationWithConfidenceResolver interface {
	Resource() *GitTreeEntryResolver
	Range() *RangeResolver
	URL(ctx context.Context) (string, error)
	CanonicalURL() (string, error)
	Placeholder() *string
}

type MarkdownWithConfidenceResolver interface {
	Text() string
	HTML() string
	Placeholder() *string
}

type LSIFJobStatsResolver interface {
	ID() graphql.ID
	ProcessingCount() int32
	ErroredCount() int32
	CompletedCount() int32
	QueuedCount() int32
	ScheduledCount() int32
}

type LSIFJobResolver interface {
	ID() graphql.ID
	Type() string
	Arguments() JSONValue
	State() string
	Failure() LSIFJobFailureReasonResolver
	QueuedAt() DateTime
	StartedAt() *DateTime
	CompletedOrErroredAt() *DateTime
}

type LSIFJobFailureReasonResolver interface {
	Summary() string
	Stacktraces() []string
}

type LSIFJobConnectionResolver interface {
	Nodes(ctx context.Context) ([]LSIFJobResolver, error)
	TotalCount(ctx context.Context) (*int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

var codeIntelOnlyInEnterprise = errors.New("lsif dumps and jobs are only available in enterprise")

func (r *schemaResolver) LSIFJobs(ctx context.Context, args *LSIFJobsQueryArgs) (LSIFJobConnectionResolver, error) {
	if EnterpriseResolvers.codeIntelResolver == nil {
		return nil, codeIntelOnlyInEnterprise
	}
	return EnterpriseResolvers.codeIntelResolver.LSIFJobs(ctx, args)
}

func (r *schemaResolver) LSIFJobStats(ctx context.Context) (LSIFJobStatsResolver, error) {
	if EnterpriseResolvers.codeIntelResolver == nil {
		return nil, codeIntelOnlyInEnterprise
	}
	return EnterpriseResolvers.codeIntelResolver.LSIFJobStats(ctx)
}

func (r *schemaResolver) DeleteLSIFDump(ctx context.Context, args *struct{ ID graphql.ID }) (*EmptyResponse, error) {
	if EnterpriseResolvers.codeIntelResolver == nil {
		return nil, codeIntelOnlyInEnterprise
	}
	return EnterpriseResolvers.codeIntelResolver.DeleteLSIFDump(ctx, args.ID)
}

func (r *schemaResolver) DeleteLSIFJob(ctx context.Context, args *struct{ ID graphql.ID }) (*EmptyResponse, error) {
	if EnterpriseResolvers.codeIntelResolver == nil {
		return nil, codeIntelOnlyInEnterprise
	}
	return EnterpriseResolvers.codeIntelResolver.DeleteLSIFJob(ctx, args.ID)
}
