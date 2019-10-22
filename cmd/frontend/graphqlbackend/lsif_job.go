package graphqlbackend

import (
	"context"
	"fmt"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

func (r *schemaResolver) LsifJob(ctx context.Context, args *struct{ ID graphql.ID }) (*lsifJobResolver, error) {
	return lsifJobByGQLID(ctx, args.ID)
}

type lsifJobResolver struct {
	lsifJob *types.LsifJob
}

func lsifJobByGQLID(ctx context.Context, id graphql.ID) (*lsifJobResolver, error) {
	jobID, err := unmarshalLsifJobGQLID(id)
	if err != nil {
		return nil, err
	}

	return lsifJobByStringID(ctx, jobID)
}

func lsifJobByStringID(ctx context.Context, id string) (*lsifJobResolver, error) {
	var lsifJob *types.LsifJob
	if err := lsifRequest(ctx, fmt.Sprintf("jobs/%s", id), nil, &lsifJob); err != nil {
		return nil, err
	}

	return &lsifJobResolver{lsifJob: lsifJob}, nil
}

func (r *lsifJobResolver) ID() graphql.ID         { return marshalLsifJobGQLID(r.lsifJob.ID) }
func (r *lsifJobResolver) Name() string           { return r.lsifJob.Name }
func (r *lsifJobResolver) Args() JSONValue        { return JSONValue{r.lsifJob.Args} }
func (r *lsifJobResolver) Status() string         { return r.lsifJob.Status }
func (r *lsifJobResolver) Progress() float64      { return r.lsifJob.Progress }
func (r *lsifJobResolver) FailedReason() *string  { return r.lsifJob.FailedReason }
func (r *lsifJobResolver) Stacktrace() *[]string  { return r.lsifJob.Stacktrace }
func (r *lsifJobResolver) Timestamp() DateTime    { return DateTime{Time: r.lsifJob.Timestamp} }
func (r *lsifJobResolver) ProcessedOn() *DateTime { return DateTimeOrNil(r.lsifJob.ProcessedOn) }
func (r *lsifJobResolver) FinishedOn() *DateTime  { return DateTimeOrNil(r.lsifJob.FinishedOn) }

func marshalLsifJobGQLID(lsifJobID string) graphql.ID {
	return relay.MarshalID("LsifJob", lsifJobID)
}

func unmarshalLsifJobGQLID(id graphql.ID) (lsifJobID string, err error) {
	err = relay.UnmarshalSpec(id, &lsifJobID)
	return
}
