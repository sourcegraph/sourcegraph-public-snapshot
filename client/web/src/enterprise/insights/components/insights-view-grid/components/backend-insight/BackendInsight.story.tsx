import React from 'react'

import { MockedResponse } from '@apollo/client/testing'
import { Meta, Story } from '@storybook/react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import { H2 } from '@sourcegraph/wildcard'

import { WebStory } from '../../../../../../components/WebStory'
import { SearchBasedInsight, SeriesChartContent } from '../../../../core'
import { GET_INSIGHT_VIEW_GQL } from '../../../../core/backend/gql-backend'
import { InsightInProcessError } from '../../../../core/backend/utils/errors'
import { InsightExecutionType, InsightType } from '../../../../core/types'

import { BackendInsightView } from './BackendInsight'

const defaultStory: Meta = {
    title: 'web/insights/BackendInsight',
    decorators: [story => <WebStory>{() => story()}</WebStory>],
}

export default defaultStory

const INSIGHT_CONFIGURATION_MOCK: SearchBasedInsight = {
    id: 'searchInsights.insight.mock_backend_insight_id',
    title: 'Backend Insight Mock',
    repositories: [],
    series: [
        { id: 'series_001', query: '', name: 'A metric', stroke: 'var(--warning)' },
        { id: 'series_002', query: '', name: 'B metric', stroke: 'var(--warning)' },
    ],
    type: InsightType.SearchBased,
    executionType: InsightExecutionType.Backend,
    step: { weeks: 2 },
    filters: { excludeRepoRegexp: '', includeRepoRegexp: '', context: '' },
    dashboardReferenceCount: 0,
    isFrozen: false,
    seriesDisplayOptions: {},
    dashboards: [],
}

interface BackendInsightDatum {
    x: number
    value: number
    link?: string
}

const getXValue = (datum: BackendInsightDatum): Date => new Date(datum.x)
const getYValue = (datum: BackendInsightDatum): number => datum.value
const getLinkURL = (datum: BackendInsightDatum): string | undefined => datum.link

const LINE_CHART_CONTENT_MOCK: SeriesChartContent<BackendInsightDatum> = {
    series: [
        {
            id: 'series_001',
            data: [
                { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: 4000, link: '#A:1st_data_point' },
                { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: 4000, link: '#A:2st_data_point' },
                { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 5600, link: '#A:3rd_data_point' },
                { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, value: 9800, link: '#A:4th_data_point' },
                { x: 1588965700286, value: 12300, link: '#A:5th_data_point' },
            ],
            name: 'A metric',
            color: 'var(--warning)',
            getXValue,
            getYValue,
            getLinkURL,
        },
        {
            id: 'series_002',
            data: [
                { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: 15000 },
                { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: 26000 },
                { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 20000 },
                { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, value: 19000 },
                { x: 1588965700286, value: 17000 },
            ],
            name: 'B metric',
            color: 'var(--warning)',
            getXValue,
            getYValue,
            getLinkURL,
        },
    ],
}

const LINE_CHART_CONTENT_MOCK_EMPTY: SeriesChartContent<BackendInsightDatum> = {
    series: [
        {
            id: 'series_001',
            data: [],
            name: 'A metric',
            color: 'var(--warning)',
            getXValue,
            getYValue,
        },
        {
            id: 'series_002',
            data: [],
            name: 'B metric',
            color: 'var(--warning)',
            getXValue,
            getYValue,
        },
    ],
}

function generateSeries(chartContent: SeriesChartContent<BackendInsightDatum>, isFetchingHistoricalData: boolean) {
    return chartContent.series.map(series => ({
        seriesId: series.id,
        label: series.name,
        points: series.data.map(point => ({
            dateTime: new Date(point.x).toUTCString(),
            value: point.value,
            __typename: 'InsightDataPoint',
        })),
        status: {
            backfillQueuedAt: '2021-06-06T15:48:11Z',
            completedJobs: 0,
            pendingJobs: isFetchingHistoricalData ? 10 : 0,
            failedJobs: 0,
            __typename: 'InsightSeriesStatus',
        },
        __typename: 'InsightsSeries',
    }))
}

const mockInsightAPIResponse = ({
    isFetchingHistoricalData = false,
    delayAmount = 0,
    throwProcessingError = false,
    hasData = true,
} = {}): MockedResponse[] => {
    if (throwProcessingError) {
        return [
            {
                request: {
                    query: GET_INSIGHT_VIEW_GQL,
                    variables: {
                        id: 'searchInsights.insight.mock_backend_insight_id',
                        filters: { includeRepoRegex: '', excludeRepoRegex: '', searchContexts: [''] },
                        seriesDisplayOptions: {
                            limit: undefined,
                            sortOptions: undefined,
                        },
                    },
                },
                error: new InsightInProcessError(),
            },
        ]
    }

    return [
        {
            request: {
                query: GET_INSIGHT_VIEW_GQL,
                variables: {
                    id: 'searchInsights.insight.mock_backend_insight_id',
                    filters: { includeRepoRegex: '', excludeRepoRegex: '', searchContexts: [''] },
                    seriesDisplayOptions: {
                        limit: undefined,
                        sortOptions: undefined,
                    },
                },
            },
            result: {
                data: {
                    insightViews: {
                        nodes: [
                            {
                                id: 'searchInsights.insight.mock_backend_insight_id',
                                appliedSeriesDisplayOptions: {
                                    limit: 20,
                                    sortOptions: {
                                        mode: 'RESULT_COUNT',
                                        direction: 'DESC',
                                        __typename: 'SeriesSortOptions',
                                    },
                                    __typename: 'SeriesDisplayOptions',
                                },
                                defaultSeriesDisplayOptions: {
                                    limit: null,
                                    sortOptions: {
                                        mode: null,
                                        direction: null,
                                        __typename: 'SeriesSortOptions',
                                    },
                                    __typename: 'SeriesDisplayOptions',
                                },
                                dataSeries: generateSeries(
                                    hasData ? LINE_CHART_CONTENT_MOCK : LINE_CHART_CONTENT_MOCK_EMPTY,
                                    isFetchingHistoricalData
                                ),
                                __typename: 'InsightView',
                            },
                        ],
                        __typename: 'InsightViewConnection',
                    },
                },
            },
            delay: delayAmount,
        },
    ]
}

const TestBackendInsight: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <BackendInsightView
        style={{ width: 400, height: 400 }}
        insight={INSIGHT_CONFIGURATION_MOCK}
        telemetryService={NOOP_TELEMETRY_SERVICE}
        innerRef={() => {}}
    />
)

export const BackendInsight: Story = () => (
    <section>
        <article>
            <H2>Card</H2>
            <MockedTestProvider addTypename={true} mocks={mockInsightAPIResponse()}>
                <TestBackendInsight />
            </MockedTestProvider>
        </article>
        <article className="mt-3">
            <H2>Card with delay API</H2>
            <MockedTestProvider mocks={mockInsightAPIResponse({ delayAmount: 2000 })}>
                <TestBackendInsight />
            </MockedTestProvider>
        </article>
        <article className="mt-3">
            <H2>Card backfilling data</H2>
            <MockedTestProvider addTypename={true} mocks={mockInsightAPIResponse({ isFetchingHistoricalData: true })}>
                <TestBackendInsight />
            </MockedTestProvider>
        </article>
        <article className="mt-3">
            <H2>Card no data</H2>
            <MockedTestProvider addTypename={true} mocks={mockInsightAPIResponse({ hasData: false })}>
                <TestBackendInsight />
            </MockedTestProvider>
        </article>
        <article className="mt-3">
            <H2>Card insight syncing</H2>
            <MockedTestProvider addTypename={true} mocks={mockInsightAPIResponse({ throwProcessingError: true })}>
                <TestBackendInsight />
            </MockedTestProvider>
        </article>
        <article className="mt-3">
            <H2>Locked Card insight</H2>
            <MockedTestProvider addTypename={true} mocks={mockInsightAPIResponse()}>
                <BackendInsightView
                    style={{ width: 400, height: 400 }}
                    insight={{ ...INSIGHT_CONFIGURATION_MOCK, isFrozen: true }}
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                    innerRef={() => {}}
                />
            </MockedTestProvider>
        </article>
    </section>
)
