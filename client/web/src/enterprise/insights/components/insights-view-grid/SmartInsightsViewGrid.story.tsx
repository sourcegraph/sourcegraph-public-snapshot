import { storiesOf } from '@storybook/react'
import React from 'react'
import { of } from 'rxjs'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../../components/WebStory'
import { LINE_CHART_CONTENT_MOCK } from '../../../../views/mocks/charts-content'
import { CodeInsightsBackendContext } from '../../core/backend/code-insights-backend-context'
import { CodeInsightsSettingsCascadeBackend } from '../../core/backend/code-insights-setting-cascade-backend'
import { Insight, InsightType } from '../../core/types'
import { SearchBackendBasedInsight } from '../../core/types/insight/search-insight'
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

class CodeInsightsStoryBackend extends CodeInsightsSettingsCascadeBackend {
    constructor() {
        super(SETTINGS_CASCADE_MOCK, {} as any)
    }

    public getInsights = () =>
        of([
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
        ] as Insight[])

    public getBackendInsightData = (input: SearchBackendBasedInsight) =>
        of({
            id: input.id,
            view: {
                title: 'Backend Insight Mock',
                subtitle: 'Backend insight description text',
                content: [LINE_CHART_CONTENT_MOCK],
                isFetchingHistoricalData: false,
            },
        })
}

const codeInsightsApi = new CodeInsightsStoryBackend()

add('SmartInsightsViewGrid', () => (
    <CodeInsightsBackendContext.Provider value={codeInsightsApi}>
        <SmartInsightsViewGrid insights={insights} telemetryService={NOOP_TELEMETRY_SERVICE} />
    </CodeInsightsBackendContext.Provider>
))
