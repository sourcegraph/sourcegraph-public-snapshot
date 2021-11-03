package graphqlbackend

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// webhookLogArgs are the arguments common to the two queries that provide
// access to webhook logs: the webhookLogs method on the top level query, and on
// the ExternalService type.
type webhookLogsArgs struct {
	First      *int
	After      *string
	OnlyErrors *bool
	Since      *time.Time
	Until      *time.Time
}

// toListOpts transforms the GraphQL webhookLogsArgs into options that can be
// provided to the WebhookLogStore's Count and List methods.
func (args *webhookLogsArgs) toListOpts(externalServiceID int64) (database.WebhookLogListOpts, error) {
	opts := database.WebhookLogListOpts{
		ExternalServiceID: &externalServiceID,
		Since:             args.Since,
		Until:             args.Until,
	}

	if args.First != nil {
		opts.Limit = *args.First
	} else {
		opts.Limit = 50
	}

	if args.After != nil {
		var err error
		opts.Cursor, err = strconv.ParseInt(*args.After, 10, 64)
		if err != nil {
			return opts, errors.Wrap(err, "parsing the after cursor")
		}
	}

	if args.OnlyErrors != nil && *args.OnlyErrors {
		opts.OnlyErrors = true
	}

	return opts, nil
}

// WebhookLogs is the top level query used to return webhook logs that weren't
// resolved to a specific external service.
func (r *schemaResolver) WebhookLogs(ctx context.Context, args *webhookLogsArgs) (*webhookLogConnectionResolver, error) {
	return newWebhookLogConnectionResolver(ctx, r.db, args, 0)
}

type webhookLogConnectionResolver struct {
	args              *webhookLogsArgs
	externalServiceID int64
	store             database.WebhookLogStore

	once sync.Once
	logs []*types.WebhookLog
	next int64
	err  error
}

func newWebhookLogConnectionResolver(ctx context.Context, db database.DB, args *webhookLogsArgs, externalServiceID int64) (*webhookLogConnectionResolver, error) {
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, db); err != nil {
		return nil, err
	}

	return &webhookLogConnectionResolver{
		args:              args,
		externalServiceID: externalServiceID,
		store:             database.WebhookLogs(db, keyring.Default().WebhookLogKey),
	}, nil
}

func (r *webhookLogConnectionResolver) Nodes(ctx context.Context) ([]*webhookLogResolver, error) {
	logs, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	nodes := make([]*webhookLogResolver, len(logs))
	db := database.NewDB(r.store.Handle().DB())
	for i, log := range logs {
		nodes[i] = &webhookLogResolver{
			db:  db,
			log: log,
		}
	}

	return nodes, nil
}

func (r *webhookLogConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	opts, err := r.args.toListOpts(r.externalServiceID)
	if err != nil {
		return 0, err
	}

	count, err := r.store.Count(ctx, opts)
	return int32(count), err
}

func (r *webhookLogConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, next, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if next == 0 {
		return graphqlutil.HasNextPage(false), nil
	}
	return graphqlutil.NextPageCursor(fmt.Sprint(next)), nil
}

func (r *webhookLogConnectionResolver) compute(ctx context.Context) ([]*types.WebhookLog, int64, error) {
	r.once.Do(func() {
		r.err = func() error {
			opts, err := r.args.toListOpts(r.externalServiceID)
			if err != nil {
				return err
			}

			r.logs, r.next, err = r.store.List(ctx, opts)
			return err
		}()
	})

	return r.logs, r.next, r.err
}

type webhookLogResolver struct {
	db  database.DB
	log *types.WebhookLog
}

func marshalWebhookLogID(id int64) graphql.ID {
	return relay.MarshalID("WebhookLog", id)
}

func unmarshalWebhookLogID(id graphql.ID) (logID int64, err error) {
	err = relay.UnmarshalSpec(id, &logID)
	return
}

func webhookLogByID(ctx context.Context, db database.DB, gqlID graphql.ID) (*webhookLogResolver, error) {
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, db); err != nil {
		return nil, err
	}

	id, err := unmarshalWebhookLogID(gqlID)
	if err != nil {
		return nil, err
	}

	log, err := database.WebhookLogs(db, keyring.Default().WebhookLogKey).GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &webhookLogResolver{db: db, log: log}, nil
}

func (r *webhookLogResolver) ID() graphql.ID {
	return marshalWebhookLogID(r.log.ID)
}

func (r *webhookLogResolver) ReceivedAt() DateTime {
	return DateTime{Time: r.log.ReceivedAt}
}

func (r *webhookLogResolver) ExternalService(ctx context.Context) (*externalServiceResolver, error) {
	if r.log.ExternalServiceID == nil {
		return nil, errors.New("no external service attached to webhook log")
	}

	return externalServiceByID(ctx, r.db, marshalExternalServiceID(*r.log.ExternalServiceID))
}

func (r *webhookLogResolver) StatusCode() int32 {
	return int32(r.log.StatusCode)
}

func (r *webhookLogResolver) Request() *webhookLogMessageResolver {
	return &webhookLogMessageResolver{message: &r.log.Request}
}

func (r *webhookLogResolver) Response() *webhookLogMessageResolver {
	return &webhookLogMessageResolver{message: &r.log.Response}
}

type webhookLogMessageResolver struct {
	message *types.WebhookLogMessage
}

func (r *webhookLogMessageResolver) Headers() []*webhookLogHeaderResolver {
	headers := make([]*webhookLogHeaderResolver, 0, len(r.message.Header))
	for k, v := range r.message.Header {
		headers = append(headers, &webhookLogHeaderResolver{
			name:   k,
			values: v,
		})
	}

	return headers
}

func (r *webhookLogMessageResolver) Body() string {
	return string(r.message.Body)
}

type webhookLogHeaderResolver struct {
	name   string
	values []string
}

func (r *webhookLogHeaderResolver) Name() string {
	return r.name
}

func (r *webhookLogHeaderResolver) Values() []string {
	return r.values
}
