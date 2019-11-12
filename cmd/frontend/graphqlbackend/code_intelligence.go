package graphqlbackend

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

// NewCodeIntelligenceResolver will be set by enterprise
var NewCodeIntelligenceResolver func() CodeIntelligenceResolver

type CodeIntelligenceResolver interface {
	LSIFDumpByGQLID(ctx context.Context, id graphql.ID) (LSIFDumpResolver, error)
	LSIFJobStatsByGQLID(ctx context.Context, id graphql.ID) (LSIFJobStatsResolver, error)
	LSIFJobByGQLID(ctx context.Context, id graphql.ID) (LSIFJobResolver, error)

	LSIFDump(ctx context.Context, args *struct{ ID graphql.ID }) (LSIFDumpResolver, error)
	LSIFDumps(args *struct {
		graphqlutil.ConnectionArgs
		Repository      graphql.ID
		Query           *string
		IsLatestForRepo *bool
		After           *string
	}) (LSIFDumpConnectionResolver, error)
	LSIFJobStats(ctx context.Context) (LSIFJobStatsResolver, error)
	LSIFJob(ctx context.Context, args *struct{ ID graphql.ID }) (LSIFJobResolver, error)
	LSIFJobs(args *struct {
		graphqlutil.ConnectionArgs
		State string
		Query *string
		After *string
	}) (LSIFJobConnectionResolver, error)
}

type LSIFDumpResolver interface {
	ID() graphql.ID
	ProjectRoot() *GitTreeEntryResolver
	IsLatestForRepo() bool
	UploadedAt() DateTime
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
	Name() string
	Args() JSONValue
	State() string
	Progress() float64
	FailedReason() *string
	Stacktrace() *[]string
	Timestamp() DateTime
	ProcessedOn() *DateTime
	FinishedOn() *DateTime
}

type LSIFJobConnectionResolver interface {
	Nodes(ctx context.Context) ([]LSIFJobResolver, error)
	TotalCount(ctx context.Context) (*int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

var codeIntelligenceOnlyInEnterprise = errors.New("lsif dumps and jobs are only available in enterprise")

func (r *schemaResolver) LSIFDumpByGQLID(ctx context.Context, id graphql.ID) (LSIFDumpResolver, error) {
	if r.codeIntelligenceResolver == nil {
		return nil, codeIntelligenceOnlyInEnterprise
	}
	return r.codeIntelligenceResolver.LSIFDumpByGQLID(ctx, id)
}

func (r *schemaResolver) LSIFJobStatsByGQLID(ctx context.Context, id graphql.ID) (LSIFJobStatsResolver, error) {
	if r.codeIntelligenceResolver == nil {
		return nil, codeIntelligenceOnlyInEnterprise
	}
	return r.codeIntelligenceResolver.LSIFJobStatsByGQLID(ctx, id)
}

func (r *schemaResolver) LSIFJobByGQLID(ctx context.Context, id graphql.ID) (LSIFJobResolver, error) {
	if r.codeIntelligenceResolver == nil {
		return nil, codeIntelligenceOnlyInEnterprise
	}
	return r.codeIntelligenceResolver.LSIFJobByGQLID(ctx, id)
}

func (r *schemaResolver) LSIFDump(ctx context.Context, args *struct{ ID graphql.ID }) (LSIFDumpResolver, error) {
	if r.codeIntelligenceResolver == nil {
		return nil, codeIntelligenceOnlyInEnterprise
	}
	return r.codeIntelligenceResolver.LSIFDump(ctx, args)
}

// TODO - add ctx
func (r *schemaResolver) LSIFDumps(args *struct {
	graphqlutil.ConnectionArgs
	Repository      graphql.ID
	Query           *string
	IsLatestForRepo *bool
	After           *string
}) (LSIFDumpConnectionResolver, error) {
	if r.codeIntelligenceResolver == nil {
		return nil, codeIntelligenceOnlyInEnterprise
	}
	return r.codeIntelligenceResolver.LSIFDumps(args)
}

func (r *schemaResolver) LSIFJobStats(ctx context.Context) (LSIFJobStatsResolver, error) {
	if r.codeIntelligenceResolver == nil {
		return nil, codeIntelligenceOnlyInEnterprise
	}
	return r.codeIntelligenceResolver.LSIFJobStats(ctx)
}

func (r *schemaResolver) LSIFJob(ctx context.Context, args *struct{ ID graphql.ID }) (LSIFJobResolver, error) {
	if r.codeIntelligenceResolver == nil {
		return nil, codeIntelligenceOnlyInEnterprise
	}
	return r.codeIntelligenceResolver.LSIFJob(ctx, args)
}

// TODO - add ctx
func (r *schemaResolver) LSIFJobs(args *struct {
	graphqlutil.ConnectionArgs
	State string
	Query *string
	After *string
}) (LSIFJobConnectionResolver, error) {
	if r.codeIntelligenceResolver == nil {
		return nil, codeIntelligenceOnlyInEnterprise
	}
	return r.codeIntelligenceResolver.LSIFJobs(args)
}
