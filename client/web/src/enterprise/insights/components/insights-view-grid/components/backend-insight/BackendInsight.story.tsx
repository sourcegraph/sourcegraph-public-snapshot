import React from 'react'

import { Meta, Story } from '@storybook/react'
import { Observable, of, throwError } from 'rxjs'
import { delay } from 'rxjs/operators'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { H2 } from '@sourcegraph/wildcard'

import { WebStory } from '../../../../../../components/WebStory'
import { CodeInsightsBackendStoryMock } from '../../../../CodeInsightsBackendStoryMock'
import { BackendInsightData, SearchBasedInsight, SeriesChartContent } from '../../../../core'
import { InsightInProcessError } from '../../../../core/backend/utils/errors'
import {
    BackendInsight as BackendInsightType,
    InsightExecutionType,
    InsightType,
    isCaptureGroupInsight,
} from '../../../../core/types'

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
    series: [],
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

const mockInsightAPI = ({
    isFetchingHistoricalData = false,
    delayAmount = 0,
    throwProcessingError = false,
    hasData = true,
} = {}) => ({
    getBackendInsightData: (insight: BackendInsightType): Observable<BackendInsightData> => {
        if (isCaptureGroupInsight(insight)) {
            throw new Error('This demo does not support capture group insight')
        }

        if (throwProcessingError) {
            return throwError(new InsightInProcessError())
        }

        return of({
            content: hasData ? LINE_CHART_CONTENT_MOCK : LINE_CHART_CONTENT_MOCK_EMPTY,
            isFetchingHistoricalData,
        }).pipe(delay(delayAmount))
    },
})

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
            <CodeInsightsBackendStoryMock mocks={mockInsightAPI()}>
                <TestBackendInsight />
            </CodeInsightsBackendStoryMock>
        </article>
        <article className="mt-3">
            <H2>Card with delay API</H2>
            <CodeInsightsBackendStoryMock mocks={mockInsightAPI({ delayAmount: 2000 })}>
                <TestBackendInsight />
            </CodeInsightsBackendStoryMock>
        </article>
        <article className="mt-3">
            <H2>Card backfilling data</H2>
            <CodeInsightsBackendStoryMock mocks={mockInsightAPI({ isFetchingHistoricalData: true })}>
                <TestBackendInsight />
            </CodeInsightsBackendStoryMock>
        </article>
        <article className="mt-3">
            <H2>Card no data</H2>
            <CodeInsightsBackendStoryMock mocks={mockInsightAPI({ hasData: false })}>
                <TestBackendInsight />
            </CodeInsightsBackendStoryMock>
        </article>
        <article className="mt-3">
            <H2>Card insight syncing</H2>
            <CodeInsightsBackendStoryMock mocks={mockInsightAPI({ throwProcessingError: true })}>
                <TestBackendInsight />
            </CodeInsightsBackendStoryMock>
        </article>
        <article className="mt-3">
            <H2>Locked Card insight</H2>
            <CodeInsightsBackendStoryMock mocks={mockInsightAPI()}>
                <BackendInsightView
                    style={{ width: 400, height: 400 }}
                    insight={{ ...INSIGHT_CONFIGURATION_MOCK, isFrozen: true }}
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                    innerRef={() => {}}
                />
            </CodeInsightsBackendStoryMock>
        </article>
    </section>
)
