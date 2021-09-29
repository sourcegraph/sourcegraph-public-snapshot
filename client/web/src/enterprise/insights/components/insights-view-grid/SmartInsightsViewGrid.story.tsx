import { storiesOf } from '@storybook/react'
import React from 'react'
import { of } from 'rxjs'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../../components/WebStory'
import { LINE_CHART_CONTENT_MOCK } from '../../../../views/mocks/charts-content'
import { InsightsApiContext } from '../../core/backend/api-provider'
import { createMockInsightAPI } from '../../core/backend/create-insights-api'
import { Insight, InsightType } from '../../core/types'
import { SETTINGS_CASCADE_MOCK } from '../../mocks/settings-cascade'

import { SmartInsightsViewGrid } from './SmartInsightsViewGrid'

const { add } = storiesOf('web/insights/SmartInsightsViewGrid', module).addDecorator(story => (
    <WebStory>{() => story()}</WebStory>
))

const insights: Insight[] = [
    {
        id: 'searchInsights.insight.Backend_1',
        type: InsightType.Backend,
        title: 'Backend insight #1',
        series: [],
        visibility: 'personal',
    },
    {
        id: 'searchInsights.insight.Backend_2',
        type: InsightType.Backend,
        title: 'Backend insight #2',
        series: [],
        visibility: 'personal',
    },
]

const mockInsightAPI = createMockInsightAPI({
    getBackendInsight: ({ id }) =>
        of({
            id,
            view: {
                title: 'Backend Insight Mock',
                subtitle: 'Backend insight description text',
                content: [LINE_CHART_CONTENT_MOCK],
                isFetchingHistoricalData: false,
            },
        }),
})

add('SmartInsightsViewGrid', () => (
    <InsightsApiContext.Provider value={mockInsightAPI}>
        <SmartInsightsViewGrid
            insights={insights}
            settingsCascade={SETTINGS_CASCADE_MOCK}
            telemetryService={NOOP_TELEMETRY_SERVICE}
            platformContext={{} as any}
        />
    </InsightsApiContext.Provider>
))
