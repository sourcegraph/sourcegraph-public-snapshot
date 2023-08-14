import type { MockedResponse } from '@apollo/client/testing'
import { addSeconds, isBefore } from 'date-fns'

import { getDocumentNode } from '@sourcegraph/http-client'

import type { ChangesetCountsOverTimeFields, ChangesetCountsOverTimeResult } from '../../../graphql-operations'

import { CHANGESET_COUNTS_OVER_TIME } from './backend'

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
let timeMarkIndex = 0
const changesetCounts: ChangesetCountsOverTimeFields[] = new Array(150).fill(undefined).map((value, index) => {
    let currentMark = timeMarks[timeMarkIndex]
    const currentDate = addSeconds(
        new Date('2019-11-13T12:00:00Z'),
        // 10 days of data
        index * Math.round((10 * 24 * 60 * 60) / 150)
    )
    while (
        timeMarks.length > timeMarkIndex + 1 &&
        !isBefore(currentDate, new Date(timeMarks[timeMarkIndex + 1].date))
    ) {
        timeMarkIndex++
        currentMark = timeMarks[timeMarkIndex]
    }
    return {
        __typename: 'ChangesetCounts',
        date: currentDate.toISOString(),
        closed: currentMark.closed,
        draft: currentMark.draft,
        merged: currentMark.merged,
        openApproved: currentMark.openApproved,
        openChangesRequested: currentMark.openChangesRequested,
        openPending: currentMark.openPending,
        total: currentMark.total,
    }
})

export const CHANGESET_COUNTS_OVER_TIME_MOCK: MockedResponse<ChangesetCountsOverTimeResult> = {
    request: {
        query: getDocumentNode(CHANGESET_COUNTS_OVER_TIME),
        variables: {
            includeArchived: false,
            batchChange: 'specid',
        },
    },
    result: {
        data: {
            node: {
                __typename: 'BatchChange',
                changesetCountsOverTime: changesetCounts,
            },
        },
    },
}
