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
	LSIFDump(ctx context.Context, args *struct{ ID graphql.ID }) (LSIFDumpResolver, error)
	LSIFDumpByGQLID(ctx context.Context, id graphql.ID) (LSIFDumpResolver, error)
	LSIFDumps(ctx context.Context, args *LSIFDumpsQueryArgs) (LSIFDumpConnectionResolver, error)
	LSIFJob(ctx context.Context, args *struct{ ID graphql.ID }) (LSIFJobResolver, error)
	LSIFJobByGQLID(ctx context.Context, id graphql.ID) (LSIFJobResolver, error)
	LSIFJobs(ctx context.Context, args *LSIFJobsQueryArgs) (LSIFJobConnectionResolver, error)
	LSIFJobStats(ctx context.Context) (LSIFJobStatsResolver, error)
	LSIFJobStatsByGQLID(ctx context.Context, id graphql.ID) (LSIFJobStatsResolver, error)
}

type LSIFDumpsQueryArgs struct {
	graphqlutil.ConnectionArgs
	Repository      graphql.ID
	Query           *string
	IsLatestForRepo *bool
	After           *string
}

type LSIFJobsQueryArgs struct {
	graphqlutil.ConnectionArgs
	State string
	Query *string
	After *string
}

type LSIFDumpResolver interface {
	ID() graphql.ID
	ProjectRoot() (*GitTreeEntryResolver, error)
	IsLatestForRepo() bool
	UploadedAt() DateTime
	ProcessedAt() DateTime
}

type LSIFDumpConnectionResolver interface {
	Nodes(ctx context.Context) ([]LSIFDumpResolver, error)
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
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

func (r *schemaResolver) LSIFDump(ctx context.Context, args *struct{ ID graphql.ID }) (LSIFDumpResolver, error) {
	if r.codeIntelResolver == nil {
		return nil, codeIntelOnlyInEnterprise
	}
	return r.codeIntelResolver.LSIFDump(ctx, args)
}

func (r *schemaResolver) LSIFDumpByGQLID(ctx context.Context, id graphql.ID) (LSIFDumpResolver, error) {
	if r.codeIntelResolver == nil {
		return nil, codeIntelOnlyInEnterprise
	}
	return r.codeIntelResolver.LSIFDumpByGQLID(ctx, id)
}

func (r *schemaResolver) LSIFDumps(ctx context.Context, args *LSIFDumpsQueryArgs) (LSIFDumpConnectionResolver, error) {
	if r.codeIntelResolver == nil {
		return nil, codeIntelOnlyInEnterprise
	}
	return r.codeIntelResolver.LSIFDumps(ctx, args)
}

func (r *schemaResolver) LSIFJob(ctx context.Context, args *struct{ ID graphql.ID }) (LSIFJobResolver, error) {
	if r.codeIntelResolver == nil {
		return nil, codeIntelOnlyInEnterprise
	}
	return r.codeIntelResolver.LSIFJob(ctx, args)
}

func (r *schemaResolver) LSIFJobByGQLID(ctx context.Context, id graphql.ID) (LSIFJobResolver, error) {
	if r.codeIntelResolver == nil {
		return nil, codeIntelOnlyInEnterprise
	}
	return r.codeIntelResolver.LSIFJobByGQLID(ctx, id)
}

func (r *schemaResolver) LSIFJobs(ctx context.Context, args *LSIFJobsQueryArgs) (LSIFJobConnectionResolver, error) {
	if r.codeIntelResolver == nil {
		return nil, codeIntelOnlyInEnterprise
	}
	return r.codeIntelResolver.LSIFJobs(ctx, args)
}

func (r *schemaResolver) LSIFJobStats(ctx context.Context) (LSIFJobStatsResolver, error) {
	if r.codeIntelResolver == nil {
		return nil, codeIntelOnlyInEnterprise
	}
	return r.codeIntelResolver.LSIFJobStats(ctx)
}

func (r *schemaResolver) LSIFJobStatsByGQLID(ctx context.Context, id graphql.ID) (LSIFJobStatsResolver, error) {
	if r.codeIntelResolver == nil {
		return nil, codeIntelOnlyInEnterprise
	}
	return r.codeIntelResolver.LSIFJobStatsByGQLID(ctx, id)
}
