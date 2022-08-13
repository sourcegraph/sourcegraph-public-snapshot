package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/conf"
)

type ExecutorsResolver interface {
	Executors(context.Context, *ExecutorsArgs) (ExecutorConnectionResolver, error)
	AreExecutorsConfigured() bool
	NodeResolvers() map[string]NodeByIDFunc
}

type ExecutorsArgs struct {
	Query  *string
	Active *bool
	First  *int32
	After  *string
}

type ExecutorConnectionResolver interface {
	Nodes() []ExecutorResolver
	TotalCount() int32
	PageInfo() *graphqlutil.PageInfo
}

type ExecutorResolver interface {
	ID() graphql.ID
	Hostname() string
	QueueName() string
	Active() bool
	Os() string
	Architecture() string
	DockerVersion() string
	ExecutorVersion() string
	GitVersion() string
	IgniteVersion() string
	SrcCliVersion() string
	FirstSeenAt() DateTime
	LastSeenAt() DateTime
	ActiveJobs(context.Context, ExecutorActiveJobArgs) (ExecutorJobConnectionResolver, error)
}

type ExecutorActiveJobArgs struct {
	First *int32
	After *string
}

type ExecutorJobConnectionResolver interface {
	Nodes() []ExecutorJobResolver
	TotalCount() int32
	PageInfo() *graphqlutil.PageInfo
}

type ExecutorJobResolver interface {
	ToLSIFIndex() (LSIFIndexResolver, error)
}

func (r *schemaResolver) AreExecutorsConfigured() bool {
	return conf.Get().ExecutorsAccessToken != ""
}
