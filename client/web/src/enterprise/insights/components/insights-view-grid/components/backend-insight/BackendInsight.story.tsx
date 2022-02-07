import { Meta, Story } from '@storybook/react'
import React from 'react'
import { of, throwError } from 'rxjs'
import { delay } from 'rxjs/operators'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../../../../components/WebStory'
import { LINE_CHART_CONTENT_MOCK, LINE_CHART_CONTENT_MOCK_EMPTY } from '../../../../../../views/mocks/charts-content'
import { CodeInsightsBackendContext } from '../../../../core/backend/code-insights-backend-context'
import { CodeInsightsSettingsCascadeBackend } from '../../../../core/backend/setting-based-api/code-insights-setting-cascade-backend'
import { InsightInProcessError } from '../../../../core/backend/utils/errors'
import {
    BackendInsight as BackendInsightType,
    InsightExecutionType,
    InsightType,
    isCaptureGroupInsight,
} from '../../../../core/types'
import { SearchBackendBasedInsight } from '../../../../core/types/insight/search-insight'
import { SETTINGS_CASCADE_MOCK } from '../../../../mocks/settings-cascade'

import { BackendInsightView } from './BackendInsight'

export default {
    title: 'web/insights/BackendInsight',
    decorators: [story => <WebStory>{() => story()}</WebStory>],
} as Meta

const INSIGHT_CONFIGURATION_MOCK: SearchBackendBasedInsight = {
    title: 'Mock Backend Insight',
    series: [],
    visibility: '',
    type: InsightExecutionType.Backend,
    viewType: InsightType.SearchBased,
    id: 'searchInsights.insight.mock_backend_insight_id',
    step: { weeks: 2 },
    filters: { excludeRepoRegexp: '', includeRepoRegexp: '' },
}

const mockInsightAPI = ({
    isFetchingHistoricalData = false,
    delayAmount = 0,
    throwProcessingError = false,
    hasData = true,
} = {}) => {
    class CodeInsightsStoryBackend extends CodeInsightsSettingsCascadeBackend {
        public getBackendInsightData = (insight: BackendInsightType) => {
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
        }
    }

    return new CodeInsightsStoryBackend(SETTINGS_CASCADE_MOCK, {} as any)
}

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
            <CodeInsightsBackendContext.Provider value={mockInsightAPI()}>
                <TestBackendInsight />
            </CodeInsightsBackendContext.Provider>
        </article>
        <article className="mt-3">
            <h2>Card with delay API</h2>
            <CodeInsightsBackendContext.Provider value={mockInsightAPI({ delayAmount: 2000 })}>
                <TestBackendInsight />
            </CodeInsightsBackendContext.Provider>
        </article>
        <article className="mt-3">
            <h2>Card backfilling data</h2>
            <CodeInsightsBackendContext.Provider value={mockInsightAPI({ isFetchingHistoricalData: true })}>
                <TestBackendInsight />
            </CodeInsightsBackendContext.Provider>
        </article>
        <article className="mt-3">
            <h2>Card no data</h2>
            <CodeInsightsBackendContext.Provider value={mockInsightAPI({ hasData: false })}>
                <TestBackendInsight />
            </CodeInsightsBackendContext.Provider>
        </article>
        <article className="mt-3">
            <h2>Card insight syncing</h2>
            <CodeInsightsBackendContext.Provider value={mockInsightAPI({ throwProcessingError: true })}>
                <TestBackendInsight />
            </CodeInsightsBackendContext.Provider>
        </article>
    </section>
)
