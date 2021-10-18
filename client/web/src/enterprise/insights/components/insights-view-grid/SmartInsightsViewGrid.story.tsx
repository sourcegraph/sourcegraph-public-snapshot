import { storiesOf } from '@storybook/react'
import React from 'react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../../components/WebStory'
import { CodeInsightsBackendContext } from '../../core/backend/code-insights-backend-context'
import { CodeInsightsSettingsCascadeBackend } from '../../core/backend/code-insights-setting-cascade-backend'
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

const codeInsightsApi = new CodeInsightsSettingsCascadeBackend(SETTINGS_CASCADE_MOCK, {} as any)

add('SmartInsightsViewGrid', () => (
    <CodeInsightsBackendContext.Provider value={codeInsightsApi}>
        <SmartInsightsViewGrid insights={insights} telemetryService={NOOP_TELEMETRY_SERVICE} />
    </CodeInsightsBackendContext.Provider>
))
