import { boolean } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'
import React from 'react'
import { of } from 'rxjs'
import { delay } from 'rxjs/operators'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../../../components/WebStory'
import { InsightsApiContext } from '../../../../core/backend/api-provider'
import { createMockInsightAPI } from '../../../../core/backend/insights-api'
import { InsightType } from '../../../../core/types'
import { SearchBackendBasedInsight } from '../../../../core/types/insight/search-insight'
import { LINE_CHART_CONTENT_MOCK } from '../../../../mocks/charts-content'
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

const mockInsightAPI = ({ isFetchingHistoricalData = false, delayAmount = 0 }) =>
    createMockInsightAPI({
        getBackendInsightById: ({ id }) =>
            of({
                id,
                view: {
                    title: 'Backend Insight Mock',
                    subtitle: 'Backend insight description text',
                    content: [LINE_CHART_CONTENT_MOCK],
                    isFetchingHistoricalData,
                },
            }).pipe(delay(delayAmount)),
    })

const loadingKnob = () => boolean('Backend loading', false)

add('Backend Insight Card', () => (
    <InsightsApiContext.Provider value={mockInsightAPI({ isFetchingHistoricalData: loadingKnob() })}>
        <BackendInsight
            style={{ width: 400, height: 400 }}
            insight={INSIGHT_CONFIGURATION_MOCK}
            settingsCascade={SETTINGS_CASCADE_MOCK}
            platformContext={{} as any}
            telemetryService={NOOP_TELEMETRY_SERVICE}
        />
    </InsightsApiContext.Provider>
))

add('Backend Insight Card with delay API', () => (
    <InsightsApiContext.Provider value={mockInsightAPI({ isFetchingHistoricalData: loadingKnob(), delayAmount: 2000 })}>
        <BackendInsight
            style={{ width: 400, height: 400 }}
            insight={INSIGHT_CONFIGURATION_MOCK}
            settingsCascade={SETTINGS_CASCADE_MOCK}
            platformContext={{} as any}
            telemetryService={NOOP_TELEMETRY_SERVICE}
        />
    </InsightsApiContext.Provider>
))
