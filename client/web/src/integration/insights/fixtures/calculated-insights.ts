import type { InsightDataNode } from '../../../graphql-operations'

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
                    diffQuery: 'type:diff',
                },
                {
                    dateTime: '2021-01-27T00:00:00Z',
                    value: 90,
                    diffQuery: 'type:diff',
                },
                {
                    dateTime: '2021-01-12T00:00:00Z',
                    value: 85,
                    diffQuery: 'type:diff',
                },
                {
                    dateTime: '2020-12-28T00:00:00Z',
                    value: 45,
                    diffQuery: 'type:diff',
                },
                {
                    dateTime: '2020-12-13T00:00:00Z',
                    value: 36,
                    diffQuery: 'type:diff',
                },
                {
                    dateTime: '2020-11-28T00:00:00Z',
                    value: 20,
                    diffQuery: 'type:diff',
                },
                {
                    dateTime: '2020-11-13T00:00:00Z',
                    value: 15,
                    diffQuery: 'type:diff',
                },
                {
                    dateTime: '2020-10-29T00:00:00Z',
                    value: 8,
                    diffQuery: 'type:diff',
                },
                {
                    dateTime: '2020-10-14T00:00:00Z',
                    value: 7,
                    diffQuery: 'type:diff',
                },
                {
                    dateTime: '2020-09-29T00:00:00Z',
                    value: 1,
                    diffQuery: 'type:diff',
                },
            ],
            status: {
                __typename: 'InsightSeriesStatus',
                isLoadingData: false,
                incompleteDatapoints: [],
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
                    diffQuery: 'type:diff',
                },
                {
                    dateTime: '2021-01-27T00:00:00Z',
                    value: 0,
                    diffQuery: 'type:diff',
                },
                {
                    dateTime: '2021-01-12T00:00:00Z',
                    value: 10,
                    diffQuery: 'type:diff',
                },
                {
                    dateTime: '2020-12-28T00:00:00Z',
                    value: 45,
                    diffQuery: 'type:diff',
                },
                {
                    dateTime: '2020-12-13T00:00:00Z',
                    value: 60,
                    diffQuery: 'type:diff',
                },
                {
                    dateTime: '2020-11-28T00:00:00Z',
                    value: 65,
                    diffQuery: 'type:diff',
                },
                {
                    dateTime: '2020-11-13T00:00:00Z',
                    value: 65,
                    diffQuery: 'type:diff',
                },
                {
                    dateTime: '2020-10-29T00:00:00Z',
                    value: 88,
                    diffQuery: 'type:diff',
                },
                {
                    dateTime: '2020-10-14T00:00:00Z',
                    value: 96,
                    diffQuery: 'type:diff',
                },
                {
                    dateTime: '2020-09-29T00:00:00Z',
                    value: 99,
                    diffQuery: 'type:diff',
                },
            ],
            status: {
                __typename: 'InsightSeriesStatus',
                isLoadingData: false,
                incompleteDatapoints: [],
            },
        },
    ],
}
