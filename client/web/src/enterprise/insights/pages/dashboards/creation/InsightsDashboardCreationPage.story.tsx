import { storiesOf } from '@storybook/react'
import React from 'react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../../../components/WebStory'
import { authUser } from '../../../../../search/panels/utils'
import { CodeInsightsBackendContext } from '../../../core/backend/code-insights-backend-context'
import { CodeInsightsSettingsCascadeBackend } from '../../../core/backend/code-insights-setting-cascade-backend';
import { SETTINGS_CASCADE_MOCK } from '../../../mocks/settings-cascade'

import { InsightsDashboardCreationPage } from './InsightsDashboardCreationPage'

const { add } = storiesOf('web/insights/InsightsDashboardCreationPage', module)
    .addDecorator(story => <WebStory>{() => story()}</WebStory>)
    .addParameters({
        chromatic: {
            viewports: [576, 1440],
        },
    })

const PLATFORM_CONTEXT = {
    // eslint-disable-next-line @typescript-eslint/require-await
    updateSettings: async (...args: any[]) => {
        console.log('PLATFORM CONTEXT update settings with', { ...args })
    },
}

const codeInsightsBackend = new CodeInsightsSettingsCascadeBackend(SETTINGS_CASCADE_MOCK, PLATFORM_CONTEXT)

add('Page', () => (
    <CodeInsightsBackendContext.Provider value={codeInsightsBackend}>
        <InsightsDashboardCreationPage telemetryService={NOOP_TELEMETRY_SERVICE} authenticatedUser={authUser} />
    </CodeInsightsBackendContext.Provider>
))
