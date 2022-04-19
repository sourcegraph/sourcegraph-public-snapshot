import React from 'react'

import { Meta, Story } from '@storybook/react'
import { Observable, of, throwError } from 'rxjs'
import { delay } from 'rxjs/operators'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../../../../components/WebStory'
import { CodeInsightsBackendStoryMock } from '../../../../CodeInsightsBackendStoryMock'
import { BackendInsightData, SearchBackendBasedInsight, SeriesChartContent } from '../../../../core'
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

const INSIGHT_CONFIGURATION_MOCK: SearchBackendBasedInsight = {
    id: 'searchInsights.insight.mock_backend_insight_id',
    title: 'Backend Insight Mock',
    series: [],
    type: InsightType.SearchBased,
    executionType: InsightExecutionType.Backend,
    step: { weeks: 2 },
    filters: { excludeRepoRegexp: '', includeRepoRegexp: '', contexts: [] },
    dashboardReferenceCount: 0,
    isFrozen: false,
}

interface BackendInsightDatum {
    x: number
    a: number
    b: number
    linkA: string
}

const LINE_CHART_CONTENT_MOCK: SeriesChartContent<BackendInsightDatum> = {
    data: [
        { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, a: 4000, b: 15000, linkA: '#A:1st_data_point' },
        { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, a: 4000, b: 26000, linkA: '#A:2st_data_point' },
        { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, a: 5600, b: 20000, linkA: '#A:3rd_data_point' },
        { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, a: 9800, b: 19000, linkA: '#A:4th_data_point' },
        { x: 1588965700286, a: 12300, b: 17000, linkA: '#A:5th_data_point' },
    ],
    getXValue: datum => new Date(datum.x),
    series: [
        {
            dataKey: 'a',
            name: 'A metric',
            color: 'var(--warning)',
            getLinkURL: datum => datum.linkA,
        },
        {
            dataKey: 'b',
            name: 'B metric',
            color: 'var(--warning)',
        },
    ],
}

const LINE_CHART_CONTENT_MOCK_EMPTY: SeriesChartContent<BackendInsightDatum> = {
    data: [],
    getXValue: datum => new Date(datum.x),
    series: [
        {
            dataKey: 'a',
            name: 'A metric',
            color: 'var(--warning)',
        },
        {
            dataKey: 'b',
            name: 'B metric',
            color: 'var(--warning)',
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

const TestBackendInsight: React.FunctionComponent = () => (
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
            <h2>Card</h2>
            <CodeInsightsBackendStoryMock mocks={mockInsightAPI()}>
                <TestBackendInsight />
            </CodeInsightsBackendStoryMock>
        </article>
        <article className="mt-3">
            <h2>Card with delay API</h2>
            <CodeInsightsBackendStoryMock mocks={mockInsightAPI({ delayAmount: 2000 })}>
                <TestBackendInsight />
            </CodeInsightsBackendStoryMock>
        </article>
        <article className="mt-3">
            <h2>Card backfilling data</h2>
            <CodeInsightsBackendStoryMock mocks={mockInsightAPI({ isFetchingHistoricalData: true })}>
                <TestBackendInsight />
            </CodeInsightsBackendStoryMock>
        </article>
        <article className="mt-3">
            <h2>Card no data</h2>
            <CodeInsightsBackendStoryMock mocks={mockInsightAPI({ hasData: false })}>
                <TestBackendInsight />
            </CodeInsightsBackendStoryMock>
        </article>
        <article className="mt-3">
            <h2>Card insight syncing</h2>
            <CodeInsightsBackendStoryMock mocks={mockInsightAPI({ throwProcessingError: true })}>
                <TestBackendInsight />
            </CodeInsightsBackendStoryMock>
        </article>
        <article className="mt-3">
            <h2>Locked Card insight</h2>
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
