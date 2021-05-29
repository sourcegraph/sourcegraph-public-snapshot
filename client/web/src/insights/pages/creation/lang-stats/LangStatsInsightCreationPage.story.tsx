import { storiesOf } from '@storybook/react'
import React from 'react'

import { EMPTY_SETTINGS_CASCADE } from '@sourcegraph/shared/src/settings/settings'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../../components/WebStory'
import { authUser } from '../../../../search/panels/utils'
import { InsightsApiContext } from '../../../core/backend/api-provider'
import { createMockInsightAPI } from '../../../core/backend/insights-api'

import { getRandomLangStatsMock } from './components/live-preview-chart/live-preview-mock-data'
import { LangStatsInsightCreationPage, LangStatsInsightCreationPageProps } from './LangStatsInsightCreationPage'

const { add } = storiesOf('web/insights/CreateLangStatsInsightPageProps', module)
    .addDecorator(story => <WebStory>{() => story()}</WebStory>)
    .addParameters({
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    })

const PLATFORM_CONTEXT: LangStatsInsightCreationPageProps['platformContext'] = {
    // eslint-disable-next-line @typescript-eslint/require-await
    updateSettings: async (...args) => {
        console.log('PLATFORM CONTEXT update settings with', { ...args })
    },
}

function sleep(delay: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, delay))
}

const mockAPI = createMockInsightAPI({
    getLangStatsInsightContent: async () => {
        await sleep(2000)

        return getRandomLangStatsMock()
    },
})

add('Page', () => (
    <InsightsApiContext.Provider value={mockAPI}>
        <LangStatsInsightCreationPage
            telemetryService={NOOP_TELEMETRY_SERVICE}
            platformContext={PLATFORM_CONTEXT}
            settingsCascade={EMPTY_SETTINGS_CASCADE}
            authenticatedUser={authUser}
        />
    </InsightsApiContext.Provider>
))
