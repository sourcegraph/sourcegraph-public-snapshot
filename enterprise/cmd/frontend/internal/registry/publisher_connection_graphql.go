package registry

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	frontendregistry "github.com/sourcegraph/sourcegraph/cmd/frontend/registry"
)

func init() {
	frontendregistry.ExtensionRegistry.PublishersFunc = extensionRegistryPublishers
}

func extensionRegistryPublishers(ctx context.Context, args *graphqlutil.ConnectionArgs) (graphqlbackend.RegistryPublisherConnection, error) {
	var opt dbPublishersListOptions
	args.Set(&opt.LimitOffset)
	return &registryPublisherConnection{opt: opt}, nil
}

// registryPublisherConnection resolves a list of registry publishers.
type registryPublisherConnection struct {
	opt dbPublishersListOptions

	// cache results because they are used by multiple fields
	once               sync.Once
	registryPublishers []*dbPublisher
	err                error
}

func (r *registryPublisherConnection) compute(ctx context.Context) ([]*dbPublisher, error) {
	r.once.Do(func() {
		opt2 := r.opt
		if opt2.LimitOffset != nil {
			tmp := *opt2.LimitOffset
			opt2.LimitOffset = &tmp
			opt2.Limit++ // so we can detect if there is a next page
		}

		r.registryPublishers, r.err = dbExtensions{}.ListPublishers(ctx, opt2)
	})
	return r.registryPublishers, r.err
}

func (r *registryPublisherConnection) Nodes(ctx context.Context) ([]graphqlbackend.RegistryPublisher, error) {
	publishers, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	var l []graphqlbackend.RegistryPublisher
	for _, publisher := range publishers {
		p, err := getRegistryPublisher(ctx, *publisher)
		if err != nil {
			return nil, err
		}
		l = append(l, p)
	}
	return l, nil
}

func (r *registryPublisherConnection) TotalCount(ctx context.Context) (int32, error) {
	count, err := dbExtensions{}.CountPublishers(ctx, r.opt)
	return int32(count), err
}

func (r *registryPublisherConnection) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	publishers, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	return graphqlutil.HasNextPage(r.opt.LimitOffset != nil && len(publishers) > r.opt.Limit), nil
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_688(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
