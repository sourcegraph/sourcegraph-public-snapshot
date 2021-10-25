package resolvers

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/batches/webhooks"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type externalServicesWithoutWebhooksResolver struct {
	store         *store.Store
	batchChangeID int64
	args          *graphqlbackend.ListExternalServicesArgs

	once                     sync.Once
	externalServiceResolvers []*graphqlbackend.ExternalServiceResolver
	next                     int64
	err                      error
}

var _ graphqlbackend.ExternalServiceConnectionResolver = &externalServicesWithoutWebhooksResolver{}

func (r *externalServicesWithoutWebhooksResolver) compute(ctx context.Context) ([]*graphqlbackend.ExternalServiceResolver, int64, error) {
	r.once.Do(func() {
		// We have to do our own pagination here, as the underlying paged query
		// doesn't know how to filter by webhook status.
		var (
			cursor int64 = 0
			first  int   = int(r.args.First) + 1
		)
		if r.args.After != nil {
			cursor, r.err = strconv.ParseInt(*r.args.After, 10, 64)
			if r.err != nil {
				return
			}
		}

		externalServices := []*types.ExternalService{}
		for {
			page, next, err := r.store.ListExternalServicesForBatchChange(ctx, store.ListExternalServicesForBatchChangeOpts{
				BatchChangeID: r.batchChangeID,
				Cursor:        cursor,
				LimitOpts:     store.LimitOpts{Limit: first},
			})
			if err != nil {
				r.err = errors.Wrap(err, "listing external services")
				return
			}

			for _, es := range page {
				cfg, err := es.Configuration()
				if err != nil {
					r.err = errors.Wrapf(err, "retrieving configuration for external service %d", es.ID)
					return
				}

				if hasWebhooks, err := webhooks.ConfigurationHasWebhooks(cfg); err != nil {
					r.err = errors.Wrapf(err, "checking webhook configuration for external service %d", es.ID)
				} else if !hasWebhooks {
					externalServices = append(externalServices, es)
				}
			}

			if len(page) == 0 || next == 0 {
				break
			}
			cursor = next
		}

		if len(externalServices) > int(r.args.First) {
			// The cursor is the ID of the first excess external service
			// resolver.
			r.next = externalServices[int(r.args.First)].ID
			externalServices = externalServices[0:int(r.args.First)]
		} else {
			r.next = 0
		}

		r.externalServiceResolvers = make([]*graphqlbackend.ExternalServiceResolver, len(externalServices))
		for i, es := range externalServices {
			r.externalServiceResolvers[i] = graphqlbackend.NewExternalServiceResolver(r.store.DB(), es)
		}
	})
	return r.externalServiceResolvers, r.next, r.err
}

func (r *externalServicesWithoutWebhooksResolver) Nodes(ctx context.Context) ([]*graphqlbackend.ExternalServiceResolver, error) {
	externalServiceResolvers, _, err := r.compute(ctx)
	return externalServiceResolvers, err
}

func (r *externalServicesWithoutWebhooksResolver) TotalCount(ctx context.Context) (int32, error) {
	count, err := r.store.CountExternalServicesForBatchChange(ctx, r.batchChangeID)
	return int32(count), err
}

func (r *externalServicesWithoutWebhooksResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, next, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if next == 0 {
		return graphqlutil.HasNextPage(false), nil
	}
	return graphqlutil.NextPageCursor(fmt.Sprint(next)), nil
}
