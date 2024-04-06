package opencodegraph

import (
	"context"
	"sort"

	"github.com/sourcegraph/sourcegraph/schema"
	"golang.org/x/sync/errgroup"
)

// multiProvider implements Provider by calling multiple providers and combining their results.
type multiProvider struct {
	providers *[]Provider
}

func (mp *multiProvider) Name() string { return "multi" }

func (mp *multiProvider) Capabilities(ctx context.Context, params schema.CapabilitiesParams) (*schema.CapabilitiesResult, error) {
	results := make([]*schema.CapabilitiesResult, len(*mp.providers))
	g, ctx := errgroup.WithContext(ctx)
	for i, p := range *mp.providers {
		i := i
		p := p
		g.Go(func() (err error) {
			results[i], err = p.Capabilities(ctx, params)
			return
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	var merged schema.CapabilitiesResult
	for _, r := range results {
		merged.Selector = append(merged.Selector, r.Selector...)
	}
	return &merged, nil
}

func (mp *multiProvider) Annotations(ctx context.Context, params schema.AnnotationsParams) (*schema.AnnotationsResult, error) {
	results := make([]*schema.AnnotationsResult, len(*mp.providers))
	g, ctx := errgroup.WithContext(ctx)
	for i, p := range *mp.providers {
		i := i
		p := p
		g.Go(func() (err error) {
			results[i], err = p.Annotations(ctx, params)
			return
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	var merged schema.AnnotationsResult
	for i, res := range results {
		providerName := (*mp.providers)[i].Name()
		for _, item := range res.Items {
			item.Id = providerName + ":" + item.Id
		}
		for _, ann := range res.Annotations {
			ann.Item.Id = providerName + ":" + ann.Item.Id
		}

		merged.Items = append(merged.Items, res.Items...)
		merged.Annotations = append(merged.Annotations, res.Annotations...)
	}

	sort.Slice(merged.Annotations, func(i, j int) bool {
		return merged.Annotations[i].Range.Start.Line < merged.Annotations[j].Range.Start.Line
	})

	return &merged, nil
}
