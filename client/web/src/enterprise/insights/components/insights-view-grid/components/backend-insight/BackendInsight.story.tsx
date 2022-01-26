import { storiesOf } from '@storybook/react'
import React from 'react'
import { of, throwError } from 'rxjs'
import { delay } from 'rxjs/operators'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../../../../components/WebStory'
import { LINE_CHART_CONTENT_MOCK, LINE_CHART_CONTENT_MOCK_EMPTY } from '../../../../../../views/mocks/charts-content'
import { CodeInsightsBackendContext } from '../../../../core/backend/code-insights-backend-context'
import { CodeInsightsSettingsCascadeBackend } from '../../../../core/backend/setting-based-api/code-insights-setting-cascade-backend'
import { InsightInProcessError } from '../../../../core/backend/utils/errors'
import { BackendInsight, InsightExecutionType, InsightType, isCaptureGroupInsight } from '../../../../core/types'
import { SearchBackendBasedInsight } from '../../../../core/types/insight/search-insight'
import { SETTINGS_CASCADE_MOCK } from '../../../../mocks/settings-cascade'

import { BackendInsightView } from './BackendInsight'

const { add } = storiesOf('web/insights/BackendInsight', module).addDecorator(story => (
    <WebStory>{() => story()}</WebStory>
))

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
        public getBackendInsightData = (insight: BackendInsight) => {
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

add('Backend Insight Card', () => (
    <CodeInsightsBackendContext.Provider value={mockInsightAPI()}>
        <TestBackendInsight />
    </CodeInsightsBackendContext.Provider>
))

add('Backend Insight Card with delay API', () => (
    <CodeInsightsBackendContext.Provider value={mockInsightAPI({ delayAmount: 2000 })}>
        <TestBackendInsight />
    </CodeInsightsBackendContext.Provider>
))

add('Backend Insight Card backfilling data', () => (
    <CodeInsightsBackendContext.Provider value={mockInsightAPI({ isFetchingHistoricalData: true })}>
        <TestBackendInsight />
    </CodeInsightsBackendContext.Provider>
))

add('Backend Insight Card no data', () => (
    <CodeInsightsBackendContext.Provider value={mockInsightAPI({ hasData: false })}>
        <TestBackendInsight />
    </CodeInsightsBackendContext.Provider>
))

add('Backend Insight Card insight syncing', () => (
    <CodeInsightsBackendContext.Provider value={mockInsightAPI({ throwProcessingError: true })}>
        <TestBackendInsight />
    </CodeInsightsBackendContext.Provider>
))
