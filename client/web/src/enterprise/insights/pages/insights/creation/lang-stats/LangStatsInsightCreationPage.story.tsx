import { Meta, Story } from '@storybook/react'
import delay from 'delay'
import { noop } from 'lodash'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../../../../components/WebStory'
import { CodeInsightsBackendStoryMock } from '../../../../CodeInsightsBackendStoryMock'

import { getRandomLangStatsMock } from './components/live-preview-chart/constants'
import { LangStatsInsightCreationPage as LangStatsInsightCreationPageComponent } from './LangStatsInsightCreationPage'

const defaultStory: Meta = {
    title: 'web/insights/creation-ui/LangStatsInsightCreationPage',
    decorators: [story => <WebStory>{() => story()}</WebStory>],
    parameters: {
        chromatic: {
            viewports: [576, 1440],
            disableSnapshot: false,
        },
    },
}

export default defaultStory

function sleep(delay: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, delay))
}

const fakeAPIRequest = async () => {
    await delay(1000)

    throw new Error('Network error')
}

const codeInsightsBackend = {
    getLangStatsInsightContent: async () => {
        await sleep(2000)

        return getRandomLangStatsMock()
    },
    getRepositorySuggestions: async () => {
        await sleep(2000)

        return [
            { id: '1', name: 'github.com/example/sub-repo-1' },
            { id: '2', name: 'github.com/example/sub-repo-2' },
            { id: '3', name: 'github.com/another-example/sub-repo-1' },
            { id: '4', name: 'github.com/another-example/sub-repo-2' },
        ]
    },
}

export const LangStatsInsightCreationPage: Story = () => (
    <CodeInsightsBackendStoryMock mocks={codeInsightsBackend}>
        <LangStatsInsightCreationPageComponent
            telemetryService={NOOP_TELEMETRY_SERVICE}
            onInsightCreateRequest={fakeAPIRequest}
            onSuccessfulCreation={noop}
            onCancel={noop}
        />
    </CodeInsightsBackendStoryMock>
)
