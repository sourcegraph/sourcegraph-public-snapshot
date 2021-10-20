import { storiesOf } from '@storybook/react'
import delay from 'delay'
import { noop } from 'lodash'
import React from 'react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../../../../components/WebStory'
import { CodeInsightsBackendContext } from '../../../../core/backend/code-insights-backend-context'
import { CodeInsightsSettingsCascadeBackend } from '../../../../core/backend/code-insights-setting-cascade-backend'
import { SupportedInsightSubject } from '../../../../core/types/subjects'
import {
    createGlobalSubject,
    createOrgSubject,
    createUserSubject,
    SETTINGS_CASCADE_MOCK,
} from '../../../../mocks/settings-cascade'

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

const fakeAPIRequest = async () => {
    await delay(1000)

    throw new Error('Network error')
}

class CodeInsightsStoryBackend extends CodeInsightsSettingsCascadeBackend {
    public getSearchInsightContent = async () => {
        await sleep(2000)

        return {
            ...DEFAULT_MOCK_CHART_CONTENT,
            data: getRandomDataForMock(),
        }
    }

    // eslint-disable-next-line @typescript-eslint/require-await
    public getRepositorySuggestions = async () => [
        { id: '1', name: 'github.com/example/sub-repo-1' },
        { id: '2', name: 'github.com/example/sub-repo-2' },
        { id: '3', name: 'github.com/another-example/sub-repo-1' },
        { id: '4', name: 'github.com/another-example/sub-repo-2' },
    ]
}

const codeInsightsBackend = new CodeInsightsStoryBackend(SETTINGS_CASCADE_MOCK, {} as any)

const SUBJECTS = [
    createUserSubject('Emir Kusturica'),
    createOrgSubject('Warner Brothers'),
    createGlobalSubject('Global'),
] as SupportedInsightSubject[]

add('Page', () => (
    <CodeInsightsBackendContext.Provider value={codeInsightsBackend}>
        <SearchInsightCreationPage
            visibility="user_test_id"
            subjects={SUBJECTS}
            telemetryService={NOOP_TELEMETRY_SERVICE}
            onInsightCreateRequest={fakeAPIRequest}
            onSuccessfulCreation={noop}
            onCancel={noop}
        />
    </CodeInsightsBackendContext.Provider>
))
