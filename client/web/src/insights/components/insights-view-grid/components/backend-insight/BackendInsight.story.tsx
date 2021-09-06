import { storiesOf } from '@storybook/react'
import React from 'react'
import { of, throwError } from 'rxjs'
import { delay } from 'rxjs/operators'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../../../components/WebStory'
import { InsightsApiContext } from '../../../../core/backend/api-provider'
import { InsightStillProcessingError } from '../../../../core/backend/api/get-backend-insight'
import { createMockInsightAPI } from '../../../../core/backend/create-insights-api'
import { InsightType } from '../../../../core/types'
import { SearchBackendBasedInsight } from '../../../../core/types/insight/search-insight'
import { LINE_CHART_CONTENT_MOCK, LINE_CHART_CONTENT_MOCK_EMPTY } from '../../../../mocks/charts-content'
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
} = {}) =>
    createMockInsightAPI({
        getBackendInsight: ({ id }) => {
            if (throwProcessingError) {
                return throwError(new InsightStillProcessingError())
            }

            return of({
                id,
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
    <BackendInsight
        style={{ width: 400, height: 400 }}
        insight={INSIGHT_CONFIGURATION_MOCK}
        settingsCascade={SETTINGS_CASCADE_MOCK}
        platformContext={{} as any}
        telemetryService={NOOP_TELEMETRY_SERVICE}
    />
)

add('Backend Insight Card', () => (
    <InsightsApiContext.Provider value={mockInsightAPI()}>
        <TestBackendInsight />
    </InsightsApiContext.Provider>
))

add('Backend Insight Card with delay API', () => (
    <InsightsApiContext.Provider value={mockInsightAPI({ delayAmount: 2000 })}>
        <TestBackendInsight />
    </InsightsApiContext.Provider>
))

add('Backend Insight Card backfilling data', () => (
    <InsightsApiContext.Provider value={mockInsightAPI({ isFetchingHistoricalData: true })}>
        <TestBackendInsight />
    </InsightsApiContext.Provider>
))

add('Backend Insight Card no data', () => (
    <InsightsApiContext.Provider value={mockInsightAPI({ hasData: false })}>
        <TestBackendInsight />
    </InsightsApiContext.Provider>
))

add('Backend Insight Card insight syncing', () => (
    <InsightsApiContext.Provider value={mockInsightAPI({ throwProcessingError: true })}>
        <TestBackendInsight />
    </InsightsApiContext.Provider>
))
