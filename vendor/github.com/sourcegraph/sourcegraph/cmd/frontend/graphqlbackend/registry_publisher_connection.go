package graphqlbackend

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
)

func (r *extensionRegistryResolver) Publishers(ctx context.Context, args *struct {
	ConnectionArgs
}) (*registryPublisherConnectionResolver, error) {
	var opt db.RegistryPublishersListOptions
	args.ConnectionArgs.Set(&opt.LimitOffset)
	return &registryPublisherConnectionResolver{opt: opt}, nil
}

// registryPublisherConnectionResolver resolves a list of registry publishers.
type registryPublisherConnectionResolver struct {
	opt db.RegistryPublishersListOptions

	// cache results because they are used by multiple fields
	once               sync.Once
	registryPublishers []*db.RegistryPublisher
	err                error
}

func (r *registryPublisherConnectionResolver) compute(ctx context.Context) ([]*db.RegistryPublisher, error) {
	r.once.Do(func() {
		opt2 := r.opt
		if opt2.LimitOffset != nil {
			tmp := *opt2.LimitOffset
			opt2.LimitOffset = &tmp
			opt2.Limit++ // so we can detect if there is a next page
		}

		r.registryPublishers, r.err = db.RegistryExtensions.ListPublishers(ctx, opt2)
	})
	return r.registryPublishers, r.err
}

func (r *registryPublisherConnectionResolver) Nodes(ctx context.Context) ([]*registryPublisher, error) {
	publishers, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	var l []*registryPublisher
	for _, publisher := range publishers {
		p, err := getRegistryPublisher(ctx, *publisher)
		if err != nil {
			return nil, err
		}
		l = append(l, p)
	}
	return l, nil
}

func (r *registryPublisherConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	count, err := db.RegistryExtensions.CountPublishers(ctx, r.opt)
	return int32(count), err
}

func (r *registryPublisherConnectionResolver) PageInfo(ctx context.Context) (*PageInfo, error) {
	publishers, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	return &PageInfo{hasNextPage: r.opt.LimitOffset != nil && len(publishers) > r.opt.Limit}, nil
}
