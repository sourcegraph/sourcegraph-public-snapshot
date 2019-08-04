package campaigns

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func (v *gqlCampaign) BurndownChart(ctx context.Context) (graphqlbackend.CampaignBurndownChart, error) {
	openThreads := []int32{2071, 1918, 1231, 1121, 1018, 1003, 980, 979, 930, 945, 715, 331}
	maxOpenBefore := func(i int) (max int32) {
		if i == 0 && len(openThreads) > 0 {
			return openThreads[0]
		}
		for _, v := range openThreads[:i] {
			if v > max {
				max = v
			}
		}
		return max
	}
	mapOpenThreads := func(f func(v int32, i int) int32) []int32 {
		vs := make([]int32, len(openThreads))
		for i, v := range openThreads {
			vs[i] = f(v, i)
		}
		return vs
	}
	dates := make([]time.Time, len(openThreads))
	for i := range openThreads {
		dates[i] = time.Now().Add(time.Duration(-1*24*(len(openThreads)-i)) * time.Hour)
	}
	return &gqlCampaignBurndownChart{
		dates:               dates,
		openThreads:         openThreads,
		mergedThreads:       mapOpenThreads(func(v int32, i int) int32 { return maxOpenBefore(i) - v + maxOpenBefore(i)/int32(len(openThreads)-i+4) }),
		closedThreads:       mapOpenThreads(func(v int32, i int) int32 { return maxOpenBefore(i) / int32(len(openThreads)-i+5) }),
		openApprovedThreads: mapOpenThreads(func(v int32, i int) int32 { return v / int32(len(openThreads)-i+2) }),
	}, nil
}

type gqlCampaignBurndownChart struct {
	dates                                                          []time.Time
	openThreads, mergedThreads, closedThreads, openApprovedThreads []int32
}

func (v *gqlCampaignBurndownChart) Dates() []graphqlbackend.DateTime {
	dates := make([]graphqlbackend.DateTime, len(v.dates))
	for i, d := range v.dates {
		dates[i] = graphqlbackend.DateTime{d}
	}
	return dates
}

func (v *gqlCampaignBurndownChart) OpenThreads() []int32 { return v.openThreads }

func (v *gqlCampaignBurndownChart) MergedThreads() []int32 { return v.mergedThreads }

func (v *gqlCampaignBurndownChart) ClosedThreads() []int32 { return v.closedThreads }

func (v *gqlCampaignBurndownChart) TotalThreads() []int32 {
	totalThreads := make([]int32, len(v.dates))
	for i := range v.dates {
		totalThreads[i] = v.openThreads[i] + v.mergedThreads[i] + v.closedThreads[i]
	}
	return totalThreads
}

func (v *gqlCampaignBurndownChart) OpenApprovedThreads() []int32 { return v.openApprovedThreads }
