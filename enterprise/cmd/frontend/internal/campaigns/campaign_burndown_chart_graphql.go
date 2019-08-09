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

func (v *gqlCampaign) BurndownChart(ctx context.Context) (graphqlbackend.CampaignBurndownChart, error) {
	// Find the earliest creation date of any thread (for the starting point of the burndown chart).
	threadConnection, err := v.Threads(ctx, &graphqlbackend.ThreadConnectionArgs{})
	if err != nil {
		return nil, err
	}
	threadNodes, err := threadConnection.Nodes(ctx)
	if err != nil {
		return nil, err
	}
	earliestCreationDate := time.Now()
	for _, thread := range threadNodes {
		createdAt, err := thread.CreatedAt(ctx)
		if err != nil {
			return nil, err
		}
		if createdAt.Time.Before(earliestCreationDate) {
			earliestCreationDate = createdAt.Time
		}
	}

	// Compute data for every 24 hours since the first thread creation date.
	const (
		maxPeriods = 30
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

	data := make([]*burndownChartData, len(dates))
	for i, date := range dates {
		var err error
		data[i], err = computeBurndownChartData(ctx, v.db.ID, date)
		if err != nil {
			return nil, err
		}
	}

	return &gqlCampaignBurndownChart{
		dates: dates,
		data:  data,
	}, nil
}

func computeBurndownChartData(ctx context.Context, campaignID int64, date time.Time) (*burndownChartData, error) {
	// A thread is considered to be part of the campaign when it was created, not when it was added
	// to the campaign. This is so that you can use campaigns to track efforts that you started
	// before you created the campaign.
	// 	threads, err := threads.DBQuery(ctx, sqlf.Sprintf(`
	// SELECT `+threads.DBSelectColumns+` FROM threads
	// INNER JOIN campaigns_threads ct ON ct.thread_id=threads.id AND ct.campaign_id=%d
	// LEFT JOIN events ON events.thread_id=threads.id AND type='CreateThread'
	// WHERE events.created_at <= %v
	// `,
	// 		campaignID,
	// 		date,
	// 	))
	// 	if err != nil {
	// 		return nil, err
	// 	}

	var data burndownChartData
	ec, err := events.GetEventConnection(ctx,
		&graphqlbackend.EventConnectionCommonArgs{
			BeforeDate: &graphqlbackend.DateTime{date},
			Types:      &[]string{"CreateThread", "MergeThread", "CloseThread", "Review"},
		},
		events.Objects{Campaign: campaignID},
	)
	if err != nil {
		return nil, err
	}
	events, err := ec.Nodes(ctx)
	if err != nil {
		return nil, err
	}
	openApprovedThreads := map[graphql.ID]struct{}{}
	for _, event := range events {
		switch {
		case event.CreateThreadEvent != nil:
			data.openThreads++
		case event.MergeThreadEvent != nil:
			data.mergedThreads++
		case event.CloseThreadEvent != nil:
			data.closedThreads++
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
