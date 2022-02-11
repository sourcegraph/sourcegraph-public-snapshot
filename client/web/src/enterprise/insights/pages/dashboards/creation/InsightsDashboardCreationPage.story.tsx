import { Meta, Story } from '@storybook/react'
import React from 'react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../../../components/WebStory'
import { CodeInsightsBackendContext } from '../../../core/backend/code-insights-backend-context'
import { CodeInsightsSettingsCascadeBackend } from '../../../core/backend/setting-based-api/code-insights-setting-cascade-backend'
import { SETTINGS_CASCADE_MOCK } from '../../../mocks/settings-cascade'

import { InsightsDashboardCreationPage as InsightsDashboardCreationPageComponent } from './InsightsDashboardCreationPage'

export default {
    title: 'web/insights/InsightsDashboardCreationPage',
    decorators: [story => <WebStory>{() => story()}</WebStory>],
    parameters: {
        chromatic: {
            viewports: [576, 1440],
            disableSnapshot: false,
        },
    },
} as Meta

const PLATFORM_CONTEXT = {
    // eslint-disable-next-line @typescript-eslint/require-await
    updateSettings: async (...args: any[]) => {
        console.log('PLATFORM CONTEXT update settings with', { ...args })
    },
}

const codeInsightsBackend = new CodeInsightsSettingsCascadeBackend(SETTINGS_CASCADE_MOCK, PLATFORM_CONTEXT)

export const InsightsDashboardCreationPage: Story = () => (
    <CodeInsightsBackendContext.Provider value={codeInsightsBackend}>
        <InsightsDashboardCreationPageComponent telemetryService={NOOP_TELEMETRY_SERVICE} />
    </CodeInsightsBackendContext.Provider>
)
