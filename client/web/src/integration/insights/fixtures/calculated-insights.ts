import { InsightDataNode } from '../../../graphql-operations'

export const MIGRATION_TO_GQL_INSIGHT_DATA_FIXTURE: InsightDataNode = {
    __typename: 'InsightView',
    id: '001',
    dataSeries: [
        {
            __typename: 'InsightsSeries',
            seriesId: '001',
            label: 'Imports of old GQL.* types',
            points: [
                {
                    dateTime: '2021-02-11T00:00:00Z',
                    value: 100,
                },
                {
                    dateTime: '2021-01-27T00:00:00Z',
                    value: 90,
                },
                {
                    dateTime: '2021-01-12T00:00:00Z',
                    value: 85,
                },
                {
                    dateTime: '2020-12-28T00:00:00Z',
                    value: 45,
                },
                {
                    dateTime: '2020-12-13T00:00:00Z',
                    value: 36,
                },
                {
                    dateTime: '2020-11-28T00:00:00Z',
                    value: 20,
                },
                {
                    dateTime: '2020-11-13T00:00:00Z',
                    value: 15,
                },
                {
                    dateTime: '2020-10-29T00:00:00Z',
                    value: 8,
                },
                {
                    dateTime: '2020-10-14T00:00:00Z',
                    value: 7,
                },
                {
                    dateTime: '2020-09-29T00:00:00Z',
                    value: 1,
                },
            ],
            status: {
                __typename: 'InsightSeriesStatus',
                backfillQueuedAt: '2022-01-01',
                completedJobs: 100,
                pendingJobs: 0,
                failedJobs: 0,
            },
        },
        {
            __typename: 'InsightsSeries',
            seriesId: '002',
            label: 'Imports of new graphql-operations types',
            points: [
                {
                    dateTime: '2021-02-11T00:00:00Z',
                    value: 0,
                },
                {
                    dateTime: '2021-01-27T00:00:00Z',
                    value: 0,
                },
                {
                    dateTime: '2021-01-12T00:00:00Z',
                    value: 10,
                },
                {
                    dateTime: '2020-12-28T00:00:00Z',
                    value: 45,
                },
                {
                    dateTime: '2020-12-13T00:00:00Z',
                    value: 60,
                },
                {
                    dateTime: '2020-11-28T00:00:00Z',
                    value: 65,
                },
                {
                    dateTime: '2020-11-13T00:00:00Z',
                    value: 65,
                },
                {
                    dateTime: '2020-10-29T00:00:00Z',
                    value: 88,
                },
                {
                    dateTime: '2020-10-14T00:00:00Z',
                    value: 96,
                },
                {
                    dateTime: '2020-09-29T00:00:00Z',
                    value: 99,
                },
            ],
            status: {
                __typename: 'InsightSeriesStatus',
                backfillQueuedAt: '2022-01-01',
                completedJobs: 100,
                pendingJobs: 0,
                failedJobs: 0,
            },
        },
    ],
}
