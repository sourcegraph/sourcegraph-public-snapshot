import React from 'react'

import { Meta, Story } from '@storybook/react'
import { of, throwError } from 'rxjs'
import { delay } from 'rxjs/operators'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../../../../components/WebStory'
import { LINE_CHART_CONTENT_MOCK, LINE_CHART_CONTENT_MOCK_EMPTY } from '../../../../../../views/mocks/charts-content'
import { CodeInsightsBackendStoryMock } from '../../../../CodeInsightsBackendStoryMock'
import { InsightInProcessError } from '../../../../core/backend/utils/errors'
import {
    BackendInsight as BackendInsightType,
    InsightExecutionType,
    InsightType,
    isCaptureGroupInsight,
} from '../../../../core/types'
import { SearchBackendBasedInsight } from '../../../../core/types/insight/types/search-insight'

import { BackendInsightView } from './BackendInsight'

const defaultStory: Meta = {
    title: 'web/insights/BackendInsight',
    decorators: [story => <WebStory>{() => story()}</WebStory>],
}

export default defaultStory

const INSIGHT_CONFIGURATION_MOCK: SearchBackendBasedInsight = {
    title: 'Mock Backend Insight',
    series: [],
    executionType: InsightExecutionType.Backend,
    type: InsightType.SearchBased,
    id: 'searchInsights.insight.mock_backend_insight_id',
    step: { weeks: 2 },
    filters: { excludeRepoRegexp: '', includeRepoRegexp: '' },
    dashboardReferenceCount: 0,
    isFrozen: false,
}

const mockInsightAPI = ({
    isFetchingHistoricalData = false,
    delayAmount = 0,
    throwProcessingError = false,
    hasData = true,
} = {}) => ({
    getBackendInsightData: (insight: BackendInsightType) => {
        if (isCaptureGroupInsight(insight)) {
            throw new Error('This demo does not support capture group insight')
        }

        if (throwProcessingError) {
            return throwError(new InsightInProcessError())
        }

        return of({
            id: insight.id,
            view: {
                title: 'Backend Insight Mock',
                subtitle: 'Backend insight description text',
                content: [hasData ? LINE_CHART_CONTENT_MOCK : LINE_CHART_CONTENT_MOCK_EMPTY],
                isFetchingHistoricalData,
            },
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
    </section>
)
