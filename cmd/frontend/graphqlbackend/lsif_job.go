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

func (r *lsifJobResolver) ID() graphql.ID       { return marshalLSIFJobGQLID(r.lsifJob.ID) }
func (r *lsifJobResolver) JobType() string      { return r.lsifJob.JobType }
func (r *lsifJobResolver) Arguments() JSONValue { return JSONValue{r.lsifJob.Argumentss} }
func (r *lsifJobResolver) State() string        { return strings.ToUpper(r.lsifJob.State) }
func (r *lsifJobResolver) Failure() *lsifFailureReasonResolver {
	return &lsifFailureReasonResolver{r.lsifJob.Failure}
}
func (r *lsifJobResolver) QueuedAt() DateTime   { return DateTime{Time: r.lsifJob.QueuedAt} }
func (r *lsifJobResolver) StartedAt() *DateTime { return DateTimeOrNil(r.lsifJob.StartedAt) }
func (r *lsifJobResolver) CompletedOrErroredAt() *DateTime {
	return DateTimeOrNil(r.lsifJob.CompletedOrErroredAt)
}

type lsifFailureReasonResolver struct {
	lsifJobFailure *types.LSIFJobFailure
}

func (r *lsifFailureReasonResolver) Summary() string       { return r.lsifJobFailure.Summary }
func (r *lsifFailureReasonResolver) Stacktraces() []string { return r.lsifJobFailure.Stacktraces }

func marshalLSIFJobGQLID(lsifJobID string) graphql.ID {
	return relay.MarshalID("LSIFJob", lsifJobID)
}

func unmarshalLSIFJobGQLID(id graphql.ID) (lsifJobID string, err error) {
	err = relay.UnmarshalSpec(id, &lsifJobID)
	return
}
