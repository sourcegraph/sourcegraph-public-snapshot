import type { GetInsightPreviewResult, LangStatsInsightContentResult } from '../../../graphql-operations'

/**
 * For code-stats insight we have 1 step just in time processing. This fixture provides mock
 * data for this single step.
 */
export const LANG_STATS_INSIGHT_DATA_FIXTURE: LangStatsInsightContentResult = {
    search: {
        results: { limitHit: false },
        stats: {
            languages: [
                { name: 'Go', totalLines: 4498 },
                { name: 'JavaScript', totalLines: 948 },
                { name: 'Markdown', totalLines: 557 },
                { name: 'CSS', totalLines: 338 },
                { name: 'HTML', totalLines: 48 },
                { name: 'Text', totalLines: 21 },
                { name: 'YAML', totalLines: 1 },
            ],
        },
    },
}

export const SEARCH_INSIGHT_LIVE_PREVIEW_FIXTURE: GetInsightPreviewResult = {
    searchInsightPreview: [
        {
            points: [
                {
                    dateTime: '2020-05-29T00:00:00Z',
                    value: 655,
                    __typename: 'InsightDataPoint',
                    pointInTimeQuery: 'type:diff',
                },
                {
                    dateTime: '2020-07-29T00:00:00Z',
                    value: 648,
                    __typename: 'InsightDataPoint',
                    pointInTimeQuery: 'type:diff',
                },
                {
                    dateTime: '2020-09-29T00:00:00Z',
                    value: 738,
                    __typename: 'InsightDataPoint',
                    pointInTimeQuery: 'type:diff',
                },
                {
                    dateTime: '2020-11-29T00:00:00Z',
                    value: 725,
                    __typename: 'InsightDataPoint',
                    pointInTimeQuery: 'type:diff',
                },
                {
                    dateTime: '2021-01-31T00:00:00Z',
                    value: 758,
                    __typename: 'InsightDataPoint',
                    pointInTimeQuery: 'type:diff',
                },
                {
                    dateTime: '2021-03-30T00:00:00Z',
                    value: 831,
                    __typename: 'InsightDataPoint',
                    pointInTimeQuery: 'type:diff',
                },
                {
                    dateTime: '2021-05-31T00:00:00Z',
                    value: 864,
                    __typename: 'InsightDataPoint',
                    pointInTimeQuery: 'type:diff',
                },
            ],
            label: 'test series #1 title',
            __typename: 'SearchInsightLivePreviewSeries',
        },
        {
            points: [
                {
                    dateTime: '2020-05-29T00:00:00Z',
                    value: 100,
                    __typename: 'InsightDataPoint',
                    pointInTimeQuery: 'type:diff',
                },
                {
                    dateTime: '2020-07-29T00:00:00Z',
                    value: 105,
                    __typename: 'InsightDataPoint',
                    pointInTimeQuery: 'type:diff',
                },
                {
                    dateTime: '2020-09-29T00:00:00Z',
                    value: 98,
                    __typename: 'InsightDataPoint',
                    pointInTimeQuery: 'type:diff',
                },
                {
                    dateTime: '2020-11-29T00:00:00Z',
                    value: 135,
                    __typename: 'InsightDataPoint',
                    pointInTimeQuery: 'type:diff',
                },
                {
                    dateTime: '2021-01-31T00:00:00Z',
                    value: 157,
                    __typename: 'InsightDataPoint',
                    pointInTimeQuery: 'type:diff',
                },
                {
                    dateTime: '2021-03-30T00:00:00Z',
                    value: 160,
                    __typename: 'InsightDataPoint',
                    pointInTimeQuery: 'type:diff',
                },
                {
                    dateTime: '2021-05-31T00:00:00Z',
                    value: 188,
                    __typename: 'InsightDataPoint',
                    pointInTimeQuery: 'type:diff',
                },
            ],
            label: 'test series #2 title',
            __typename: 'SearchInsightLivePreviewSeries',
        },
    ],
}
