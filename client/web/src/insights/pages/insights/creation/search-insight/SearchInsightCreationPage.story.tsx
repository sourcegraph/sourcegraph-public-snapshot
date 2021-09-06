import { storiesOf } from '@storybook/react'
import delay from 'delay'
import { noop } from 'lodash'
import React from 'react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../../../components/WebStory'
import { InsightsApiContext } from '../../../../core/backend/api-provider'
import { createMockInsightAPI } from '../../../../core/backend/create-insights-api'
import { SETTINGS_CASCADE_MOCK } from '../../../../mocks/settings-cascade'

import {
    DEFAULT_MOCK_CHART_CONTENT,
    getRandomDataForMock,
} from './components/live-preview-chart/live-preview-mock-data'
import { SearchInsightCreationPage } from './SearchInsightCreationPage'

const { add } = storiesOf('web/insights/SearchInsightCreationPage', module)
    .addDecorator(story => <WebStory>{() => story()}</WebStory>)
    .addParameters({
        chromatic: {
            viewports: [576, 1440],
        },
    })

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

const fakeAPIRequest = async () => {
    await delay(1000)

    throw new Error('Network error')
}

add('Page', () => (
    <InsightsApiContext.Provider value={mockAPI}>
        <SearchInsightCreationPage
            visibility="user_test_id"
            telemetryService={NOOP_TELEMETRY_SERVICE}
            settingsCascade={SETTINGS_CASCADE_MOCK}
            onInsightCreateRequest={fakeAPIRequest}
            onSuccessfulCreation={noop}
            onCancel={noop}
        />
    </InsightsApiContext.Provider>
))
