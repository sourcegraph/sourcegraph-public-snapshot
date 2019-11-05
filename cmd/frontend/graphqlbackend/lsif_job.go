package graphqlbackend

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/lsif"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

func (r *schemaResolver) LSIFJob(ctx context.Context, args *struct{ ID graphql.ID }) (*lsifJobResolver, error) {
	return lsifJobByGQLID(ctx, args.ID)
}

type lsifJobResolver struct {
	lsifJob *types.LSIFJob
}

func lsifJobByGQLID(ctx context.Context, id graphql.ID) (*lsifJobResolver, error) {
	jobID, err := unmarshalLSIFJobGQLID(id)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/jobs/%s", url.PathEscape(jobID))

	var lsifJob *types.LSIFJob
	if err := lsif.TraceRequestAndUnmarshalPayload(ctx, path, nil, &lsifJob); err != nil {
		return nil, err
	}

	return &lsifJobResolver{lsifJob: lsifJob}, nil
}

func (r *lsifJobResolver) ID() graphql.ID         { return marshalLSIFJobGQLID(r.lsifJob.ID) }
func (r *lsifJobResolver) Name() string           { return r.lsifJob.Name }
func (r *lsifJobResolver) Args() JSONValue        { return JSONValue{r.lsifJob.Args} }
func (r *lsifJobResolver) State() string          { return strings.ToUpper(r.lsifJob.State) }
func (r *lsifJobResolver) Progress() float64      { return r.lsifJob.Progress }
func (r *lsifJobResolver) FailedReason() *string  { return r.lsifJob.FailedReason }
func (r *lsifJobResolver) Stacktrace() *[]string  { return r.lsifJob.Stacktrace }
func (r *lsifJobResolver) Timestamp() DateTime    { return DateTime{Time: r.lsifJob.Timestamp} }
func (r *lsifJobResolver) ProcessedOn() *DateTime { return DateTimeOrNil(r.lsifJob.ProcessedOn) }
func (r *lsifJobResolver) FinishedOn() *DateTime  { return DateTimeOrNil(r.lsifJob.FinishedOn) }

func marshalLSIFJobGQLID(lsifJobID string) graphql.ID {
	return relay.MarshalID("LSIFJob", lsifJobID)
}

func unmarshalLSIFJobGQLID(id graphql.ID) (lsifJobID string, err error) {
	err = relay.UnmarshalSpec(id, &lsifJobID)
	return
}
