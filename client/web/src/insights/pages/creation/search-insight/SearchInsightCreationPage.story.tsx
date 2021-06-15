import { storiesOf } from '@storybook/react'
import React from 'react'

import { EMPTY_SETTINGS_CASCADE } from '@sourcegraph/shared/src/settings/settings'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../../components/WebStory'
import { authUser } from '../../../../search/panels/utils'
import { InsightsApiContext } from '../../../core/backend/api-provider'
import { createMockInsightAPI } from '../../../core/backend/insights-api'

import {
    DEFAULT_MOCK_CHART_CONTENT,
    getRandomDataForMock,
} from './components/live-preview-chart/live-preview-mock-data'
import { SearchInsightCreationPage, SearchInsightCreationPageProps } from './SearchInsightCreationPage'

const { add } = storiesOf('web/insights/SearchInsightCreationPage', module)
    .addDecorator(story => <WebStory>{() => story()}</WebStory>)
    .addParameters({
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    })

const PLATFORM_CONTEXT: SearchInsightCreationPageProps['platformContext'] = {
    // eslint-disable-next-line @typescript-eslint/require-await
    updateSettings: async (...args) => {
        console.log('PLATFORM CONTEXT update settings with', { ...args })
    },
}

function sleep(delay: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, delay))
}

const mockAPI = createMockInsightAPI({
    getSearchInsightContent: async () => {
        await sleep(2000)

        return {
            ...DEFAULT_MOCK_CHART_CONTENT,
            data: getRandomDataForMock(),
        }
    },
    // eslint-disable-next-line @typescript-eslint/require-await
    getRepositorySuggestions: async () => [
        { id: '1', name: 'github.com/example/sub-repo-1' },
        { id: '2', name: 'github.com/example/sub-repo-2' },
        { id: '3', name: 'github.com/another-example/sub-repo-1' },
        { id: '4', name: 'github.com/another-example/sub-repo-2' },
    ],
})

add('Page', () => (
    <InsightsApiContext.Provider value={mockAPI}>
        <SearchInsightCreationPage
            telemetryService={NOOP_TELEMETRY_SERVICE}
            platformContext={PLATFORM_CONTEXT}
            settingsCascade={EMPTY_SETTINGS_CASCADE}
            authenticatedUser={authUser}
        />
    </InsightsApiContext.Provider>
))
