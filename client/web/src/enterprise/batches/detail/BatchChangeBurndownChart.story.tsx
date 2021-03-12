import { storiesOf } from '@storybook/react'
import React from 'react'
import { BatchChangeBurndownChart } from './BatchChangeBurndownChart'
import { of } from 'rxjs'
import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'
import { ChangesetCountsOverTimeFields } from '../../../graphql-operations'
import { addSeconds, isBefore } from 'date-fns'
import { useMemo } from '@storybook/addons'

const { add } = storiesOf('web/batches/BurndownChart', module).addDecorator(story => (
    <div className="p-3 container web-content">{story()}</div>
))

// tzs := state.GenerateTimestamps(start, end)
// 	wantCounts := make([]apitest.ChangesetCounts, 0, len(tzs))
// 	idx := 0
// 	for _, tz := range tzs {
// 		currentWant := wantEntries[idx]
// 		for len(wantEntries) > idx+1 && !tz.Before(wantEntries[idx+1].Time) {
// 			idx++
// 			currentWant = wantEntries[idx]
// 		}
// 		wantCounts = append(wantCounts, apitest.ChangesetCounts{
// 			Date:                 marshalDateTime(t, tz),
// 			Total:                currentWant.Total,
// 			Merged:               currentWant.Merged,
// 			Closed:               currentWant.Closed,
// 			Open:                 currentWant.Open,
// 			Draft:                currentWant.Draft,
// 			OpenApproved:         currentWant.OpenApproved,
// 			OpenChangesRequested: currentWant.OpenChangesRequested,
// 			OpenPending:          currentWant.OpenPending,
// 		})
// 	}

add('All states', () => {
    const changesetCounts = useMemo<ChangesetCountsOverTimeFields[]>(() => {
        const timeMarks = [
            {
                __typename: 'ChangesetCounts',
                date: '2019-11-13T12:00:00Z',
                closed: 0,
                merged: 0,
                openApproved: 0,
                openChangesRequested: 0,
                openPending: 0,
                total: 10,
                draft: 10,
                open: 0,
            },
            {
                __typename: 'ChangesetCounts',
                date: '2019-11-14T12:00:00Z',
                closed: 0,
                merged: 0,
                openApproved: 0,
                openChangesRequested: 0,
                openPending: 2,
                total: 10,
                draft: 5,
                open: 5,
            },
            {
                __typename: 'ChangesetCounts',
                date: '2019-11-15T12:00:00Z',
                closed: 0,
                merged: 1,
                openApproved: 1,
                openChangesRequested: 0,
                openPending: 3,
                total: 10,
                draft: 0,
                open: 8,
            },
            {
                __typename: 'ChangesetCounts',
                date: '2019-11-16T12:00:00Z',
                closed: 1,
                merged: 1,
                openApproved: 1,
                openChangesRequested: 0,
                openPending: 3,
                total: 10,
                draft: 0,
                open: 7,
            },
            {
                __typename: 'ChangesetCounts',
                date: '2019-11-17T12:00:00Z',
                closed: 1,
                merged: 2,
                openApproved: 1,
                openChangesRequested: 0,
                openPending: 5,
                total: 10,
                draft: 0,
                open: 6,
            },
            {
                __typename: 'ChangesetCounts',
                date: '2019-11-18T12:00:00Z',
                closed: 2,
                merged: 4,
                openApproved: 0,
                openChangesRequested: 2,
                openPending: 2,
                total: 10,
                draft: 0,
                open: 4,
            },
            {
                __typename: 'ChangesetCounts',
                date: '2019-11-19T12:00:00Z',
                closed: 2,
                merged: 4,
                openApproved: 0,
                openChangesRequested: 2,
                openPending: 2,
                total: 10,
                draft: 0,
                open: 4,
            },
            {
                __typename: 'ChangesetCounts',
                date: '2019-11-20T12:00:00Z',
                closed: 2,
                merged: 4,
                openApproved: 0,
                openChangesRequested: 2,
                openPending: 2,
                total: 10,
                draft: 0,
                open: 4,
            },
            {
                __typename: 'ChangesetCounts',
                date: '2019-11-21T12:00:00Z',
                closed: 2,
                merged: 5,
                openApproved: 3,
                openChangesRequested: 0,
                openPending: 0,
                total: 10,
                draft: 0,
                open: 3,
            },
            {
                __typename: 'ChangesetCounts',
                date: '2019-11-22T12:00:00Z',
                closed: 2,
                merged: 8,
                openApproved: 0,
                openChangesRequested: 0,
                openPending: 0,
                total: 10,
                draft: 0,
                open: 0,
            },
        ]
        let index_ = 0
        return new Array(150).fill(undefined).map((value, index) => {
            let currentMark = timeMarks[index_]
            const tz = addSeconds(
                new Date('2019-11-13T12:00:00Z'),
                // 10 days of data
                index * Math.round((10 * 24 * 60 * 60) / 150)
            )
            while (timeMarks.length > index_ + 1 && !isBefore(tz, new Date(timeMarks[index_ + 1].date))) {
                index_++
                currentMark = timeMarks[index_]
            }
            return {
                __typename: 'ChangesetCounts',
                date: tz.toISOString(),
                closed: currentMark.closed,
                draft: currentMark.draft,
                merged: currentMark.merged,
                openApproved: currentMark.openApproved,
                openChangesRequested: currentMark.openChangesRequested,
                openPending: currentMark.openPending,
                total: currentMark.total,
            }
        })
    }, [])
    return (
        <EnterpriseWebStory>
            {props => (
                <BatchChangeBurndownChart
                    {...props}
                    batchChangeID="123"
                    queryChangesetCountsOverTime={() => of(changesetCounts)}
                />
            )}
        </EnterpriseWebStory>
    )
})
