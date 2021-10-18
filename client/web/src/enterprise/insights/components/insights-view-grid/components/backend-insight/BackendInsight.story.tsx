import { storiesOf } from '@storybook/react'
import React from 'react'
import { of, throwError } from 'rxjs'
import { delay } from 'rxjs/operators'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../../../../components/WebStory'
import { LINE_CHART_CONTENT_MOCK, LINE_CHART_CONTENT_MOCK_EMPTY } from '../../../../../../views/mocks/charts-content'
import { InsightStillProcessingError } from '../../../../core/backend/api/get-backend-insight'
import { CodeInsightsBackendContext } from '../../../../core/backend/code-insights-backend-context'
import { CodeInsightsSettingsCascadeBackend } from '../../../../core/backend/code-insights-setting-cascade-backend'
import { InsightType } from '../../../../core/types'
import { SearchBackendBasedInsight } from '../../../../core/types/insight/search-insight'
import { SETTINGS_CASCADE_MOCK } from '../../../../mocks/settings-cascade'

import { BackendInsight } from './BackendInsight'

const { add } = storiesOf('web/insights/BackendInsight', module).addDecorator(story => (
    <WebStory>{() => story()}</WebStory>
))

const INSIGHT_CONFIGURATION_MOCK: SearchBackendBasedInsight = {
    title: 'Mock Backend Insight',
    series: [],
    visibility: '',
    type: InsightType.Backend,
    id: 'searchInsights.insight.mock_backend_insight_id',
}

const mockInsightAPI = ({
    isFetchingHistoricalData = false,
    delayAmount = 0,
    throwProcessingError = false,
    hasData = true,
} = {}) => {
    class CodeInsightsStoryBackend extends CodeInsightsSettingsCascadeBackend {
        public getBackendInsightData = (insight: SearchBackendBasedInsight) => {
            if (throwProcessingError) {
                return throwError(new InsightStillProcessingError())
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
    <BackendInsight
        style={{ width: 400, height: 400 }}
        insight={INSIGHT_CONFIGURATION_MOCK}
        telemetryService={NOOP_TELEMETRY_SERVICE}
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
