package graphqlbackend

import (
	"context"
	"errors"
	"sort"
	"sync"

	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
	"github.com/sourcegraph/sourcegraph/pkg/registry"
	"github.com/sourcegraph/sourcegraph/schema"
)

type configuredExtensionConnectionArgs struct {
	connectionArgs
	Enabled  bool
	Disabled bool
	Invalid  bool
}

func (r *schemaResolver) ViewerConfiguredExtensions(ctx context.Context, args *configuredExtensionConnectionArgs) (*configuredExtensionConnectionResolver, error) {
	if conf.Platform() == nil {
		return nil, errors.New("platform disabled")
	}
	cascade, err := r.ViewerConfiguration(ctx)
	if err != nil {
		return nil, err
	}
	return &configuredExtensionConnectionResolver{args: *args, cascade: cascade}, nil
}

func (r *siteResolver) ConfiguredExtensions(ctx context.Context, args *configuredExtensionConnectionArgs) (*configuredExtensionConnectionResolver, error) {
	if conf.Platform() == nil {
		return nil, errors.New("platform disabled")
	}
	return &configuredExtensionConnectionResolver{
		args:    *args,
		cascade: &configurationCascadeResolver{subject: &configurationSubject{site: r}},
	}, nil
}

func (r *userResolver) ConfiguredExtensions(ctx context.Context, args *configuredExtensionConnectionArgs) (*configuredExtensionConnectionResolver, error) {
	if conf.Platform() == nil {
		return nil, errors.New("platform disabled")
	}
	return &configuredExtensionConnectionResolver{
		args:    *args,
		cascade: &configurationCascadeResolver{subject: &configurationSubject{user: r}},
	}, nil
}

func (r *orgResolver) ConfiguredExtensions(ctx context.Context, args *configuredExtensionConnectionArgs) (*configuredExtensionConnectionResolver, error) {
	if conf.Platform() == nil {
		return nil, errors.New("platform disabled")
	}
	return &configuredExtensionConnectionResolver{
		args:    *args,
		cascade: &configurationCascadeResolver{subject: &configurationSubject{org: r}},
	}, nil
}

func (r *extensionConfigurationSubject) ConfiguredExtensions(ctx context.Context, args *configuredExtensionConnectionArgs) (*configuredExtensionConnectionResolver, error) {
	if conf.Platform() == nil {
		return nil, errors.New("platform disabled")
	}
	return &configuredExtensionConnectionResolver{
		args:    *args,
		cascade: &configurationCascadeResolver{subject: r.subject},
	}, nil
}

type configuredExtensionConnectionResolver struct {
	args    configuredExtensionConnectionArgs
	cascade *configurationCascadeResolver // the config cascade for the config subject

	// cache results because they are used by multiple fields
	once    sync.Once
	results []configuredExtensionResult
	err     error
}

type configuredExtensionResult struct {
	extensionID string
	enabled     bool
}

func (r *configuredExtensionConnectionResolver) compute(ctx context.Context) ([]configuredExtensionResult, error) {
	r.once.Do(func() {
		configResolver, err := r.cascade.Merged(ctx)
		if err != nil {
			r.err = err
			return
		}

		var settings schema.Settings
		if err := conf.UnmarshalJSON(configResolver.contents, &settings); err != nil {
			r.err = err
			return
		}

		for extensionID, s := range settings.Extensions {
			if !r.args.Invalid {
				if _, err := getExtensionByExtensionID(ctx, extensionID); errcode.IsNotFound(err) || registry.IsRemoteRegistryError(err) {
					continue // omit
				} else if err != nil {
					r.err = err
					return
				}
			}

			enabled := !s.Disabled
			if (enabled && r.args.Enabled) || (!enabled && r.args.Disabled) {
				r.results = append(r.results, configuredExtensionResult{extensionID: extensionID, enabled: enabled})
			}
		}
		sort.Slice(r.results, func(i, j int) bool {
			return r.results[i].extensionID < r.results[j].extensionID
		})
	})
	return r.results, r.err
}

func (r *configuredExtensionConnectionResolver) Nodes(ctx context.Context) ([]*configuredExtensionResolver, error) {
	results, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if r.args.First != nil && len(results) > int(*r.args.First) {
		results = results[:int(*r.args.First)]
	}

	var l []*configuredExtensionResolver
	for _, result := range results {
		l = append(l, &configuredExtensionResolver{extensionID: result.extensionID, subject: r.cascade.subject, enabled: result.enabled})
	}
	return l, nil
}

func (r *configuredExtensionConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	results, err := r.compute(ctx)
	if err != nil {
		return 0, err
	}
	return int32(len(results)), nil
}

func (r *configuredExtensionConnectionResolver) PageInfo(ctx context.Context) (*pageInfo, error) {
	results, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	return &pageInfo{hasNextPage: r.args.First != nil && len(results) > int(*r.args.First)}, nil
}

func (r *configuredExtensionConnectionResolver) URL(ctx context.Context) *string {
	const urlSuffix = "/extensions"
	switch {
	case r.cascade.subject.user != nil:
		return strptr(r.cascade.subject.user.URL() + urlSuffix)
	case r.cascade.subject.org != nil:
		return strptr(r.cascade.subject.org.URL() + urlSuffix)
	default:
		// There is currently no URL for listing unauthenticated user or site-wide configured
		// extensions.
		return nil
	}
}
