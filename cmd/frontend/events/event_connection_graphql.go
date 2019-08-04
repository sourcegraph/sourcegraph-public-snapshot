package events

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

func GetEventConnection(ctx context.Context, arg *graphqlbackend.EventConnectionCommonArgs, objects Objects) (graphqlbackend.EventConnection, error) {
	opt := dbEventsListOptions{Objects: objects}
	if arg.AfterDate != nil {
		opt.AfterDate = arg.AfterDate.Time
	}
	if arg.BeforeDate != nil {
		opt.BeforeDate = arg.BeforeDate.Time
	}
	arg.ConnectionArgs.Set(&opt.LimitOffset)
	return &eventConnection{opt: opt}, nil
}

type eventConnection struct {
	opt dbEventsListOptions

	once   sync.Once
	events []*dbEvent
	err    error
}

func (r *eventConnection) compute(ctx context.Context) ([]*dbEvent, error) {
	r.once.Do(func() {
		opt2 := r.opt
		if opt2.LimitOffset != nil {
			tmp := *opt2.LimitOffset
			opt2.LimitOffset = &tmp
			opt2.Limit++ // so we can detect if there is a next page
		}

		r.events, r.err = dbEvents{}.List(ctx, opt2)
	})
	return r.events, r.err
}

func (r *eventConnection) Nodes(ctx context.Context) ([]graphqlbackend.ToEvent, error) {
	events, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	if r.opt.LimitOffset != nil && len(events) > r.opt.LimitOffset.Limit {
		events = events[:r.opt.LimitOffset.Limit]
	}

	toEvents := make([]graphqlbackend.ToEvent, len(events))
	for i, e := range events {
		var err error
		toEvents[i], err = toRegisteredEventType(ctx, e)
		if err != nil {
			return nil, err
		}
	}
	return toEvents, nil
}

func (r *eventConnection) TotalCount(ctx context.Context) (int32, error) {
	count, err := dbEvents{}.Count(ctx, r.opt)
	return int32(count), err
}

func (r *eventConnection) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	events, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	return graphqlutil.HasNextPage(r.opt.LimitOffset != nil && len(events) > r.opt.Limit), nil
}
