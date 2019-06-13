package campaigns

import (
	"context"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/events"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

type burndownChartData struct {
	openThreads, mergedThreads, closedThreads, openApprovedThreads int32
}

func campaignBurndownChart(ctx context.Context, campaign interface {
	threadsGetter
	eventsGetter
}) (graphqlbackend.CampaignBurndownChart, error) {
	// Find the earliest creation date of any thread (for the starting point of the burndown chart).
	threads, err := campaign.getThreads(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	earliestCreationDate := now
	for _, thread := range threads {
		// The creation date of ThreadPreviews is now, so we ignore them for the purpose of
		// computing earliestCreationDate.
		if thread.Thread != nil {
			createdAt, err := thread.Thread.CreatedAt(ctx)
			if err != nil {
				return nil, err
			}
			if createdAt.Time.Before(earliestCreationDate) {
				earliestCreationDate = createdAt.Time
			}
		}
	}

	// Compute data for every 24 hours since the first thread creation date.
	const (
		maxPeriods = 9
		minPeriod  = 12 * time.Hour
	)
	sinceCreation := time.Since(earliestCreationDate)
	period := sinceCreation / maxPeriods
	if period < minPeriod {
		period = minPeriod
	}
	numPeriods := int(sinceCreation / period)
	dates := make([]time.Time, numPeriods+1)
	for i := 0; i < numPeriods; i++ {
		dates[i] = earliestCreationDate.Add(time.Duration(period * time.Duration(i)))
	}
	dates[numPeriods] = time.Now()

	// To compute the data for the chart, we look at all of the events that occurred before each
	// date. For campaign previews, we don't have the events persisted, so we need to fetch them on
	// the fly. TODO!(sqs): implement this for campaign previews
	data := make([]*burndownChartData, len(dates))
	for i, date := range dates {
		// TODO!(sqs): use constants for these event types
		events, err := campaign.getEvents(ctx, date, []events.Type{"CreateThread", "MergeThread", "CloseThread", "Review"})
		if err != nil {
			return nil, err
		}
		data[i], err = computeBurndownChartData(ctx, events)
		if err != nil {
			return nil, err
		}
	}

	return &gqlCampaignBurndownChart{
		dates: dates,
		data:  data,
	}, nil
}

func computeBurndownChartData(ctx context.Context, events []graphqlbackend.ToEvent) (*burndownChartData, error) {
	var data burndownChartData
	openApprovedThreads := map[graphql.ID]struct{}{}
	for _, event := range events {
		switch {
		case event.CreateThreadEvent != nil:
			data.openThreads++
		case event.MergeThreadEvent != nil:
			data.mergedThreads++
		case event.CloseThreadEvent != nil:
			data.closedThreads++
			data.openThreads--
		case event.ReviewEvent != nil:
			// TODO!(sqs): check if this is sufficient or if other reviews are required, or maybe
			// this reviewer went back and changed teir review to non-approved, etc.
			if event.ReviewEvent.State() == graphqlbackend.ReviewStateApproved {
				openApprovedThreads[event.ReviewEvent.Thread().ID()] = struct{}{}
			}
		}
	}
	data.openApprovedThreads = int32(len(openApprovedThreads))

	// Every merged thread was also closed, so remove double-counting.
	data.closedThreads -= data.mergedThreads

	return &data, nil
}

type gqlCampaignBurndownChart struct {
	dates []time.Time
	data  []*burndownChartData
}

func (v *gqlCampaignBurndownChart) series(f func(d *burndownChartData) int32) []int32 {
	series := make([]int32, len(v.data))
	for i, data := range v.data {
		series[i] = f(data)
	}
	return series
}

func (v *gqlCampaignBurndownChart) Dates() []graphqlbackend.DateTime {
	dates := make([]graphqlbackend.DateTime, len(v.dates))
	for i, d := range v.dates {
		dates[i] = graphqlbackend.DateTime{d}
	}
	return dates
}

func (v *gqlCampaignBurndownChart) OpenThreads() []int32 {
	return v.series(func(d *burndownChartData) int32 { return d.openThreads })
}

func (v *gqlCampaignBurndownChart) MergedThreads() []int32 {
	return v.series(func(d *burndownChartData) int32 { return d.mergedThreads })
}

func (v *gqlCampaignBurndownChart) ClosedThreads() []int32 {
	return v.series(func(d *burndownChartData) int32 { return d.closedThreads })
}

func (v *gqlCampaignBurndownChart) TotalThreads() []int32 {
	return v.series(func(d *burndownChartData) int32 { return d.openThreads + d.mergedThreads + d.closedThreads })
}

func (v *gqlCampaignBurndownChart) OpenApprovedThreads() []int32 {
	return v.series(func(d *burndownChartData) int32 { return d.openApprovedThreads })
}
