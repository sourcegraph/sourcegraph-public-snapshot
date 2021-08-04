import { storiesOf } from '@storybook/react'
import React from 'react'
import { of } from 'rxjs'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../../../components/WebStory'
import { InsightsApiContext } from '../../../../core/backend/api-provider'
import { createMockInsightAPI } from '../../../../core/backend/insights-api'
import { InsightType, SearchBackendBasedInsight } from '../../../../core/types'
import { LINE_CHART_CONTENT_MOCK } from '../../../../mocks/charts-content'
import { SETTINGS_CASCADE } from '../../../../mocks/settings-cascade'

import { BackendInsight } from './BackendInsight'

const { add } = storiesOf('web/insights/BackendInsight', module).addDecorator(story => (
    <WebStory>{() => story()}</WebStory>
))

const INSIGHT_CONFIGURATION_MOCK: SearchBackendBasedInsight = {
    title: 'Mock Backend Insight',
    repositories: [],
    series: [],
    step: { months: 2 },
    visibility: '',
    type: InsightType.Backend,
    id: 'searchInsights.insight.mock_backend_insight_id',
}

const mockInsightAPI = createMockInsightAPI({
    getBackendInsightById: (id: string) =>
        of({
            id,
            view: {
                title: 'Backend Insight Mock',
                subtitle: 'Backend insight description text',
                content: [LINE_CHART_CONTENT_MOCK],
            },
        }),
})

add('Backend Insight Card', () => (
    <InsightsApiContext.Provider value={mockInsightAPI}>
        <BackendInsight
            style={{ width: 400, height: 400 }}
            insight={INSIGHT_CONFIGURATION_MOCK}
            settingsCascade={SETTINGS_CASCADE}
            platformContext={{} as any}
            telemetryService={NOOP_TELEMETRY_SERVICE}
        />
    </InsightsApiContext.Provider>
))
