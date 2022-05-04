import { Meta, Story } from '@storybook/react'
import delay from 'delay'
import { noop, random } from 'lodash'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../../../../components/WebStory'
import { CodeInsightsBackendStoryMock } from '../../../../CodeInsightsBackendStoryMock'
import { SERIES_MOCK_CHART } from '../../../../components/creation-ui-kit'

import { SearchInsightCreationPage as SearchInsightCreationPageComponent } from './SearchInsightCreationPage'

const defaultStory: Meta = {
    title: 'web/insights/creation-ui/SearchInsightCreationPage',
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

const getRandomDataForMock = (): unknown[] =>
    new Array(6).fill(null).map((item, index) => ({
        x: 1588965700286 - index * 24 * 60 * 60 * 1000,
        a: random(20, 200),
        b: random(10, 200),
    }))

const fakeAPIRequest = async () => {
    await delay(1000)

    throw new Error('Network error')
}

const codeInsightsBackend = {
    getSearchInsightContent: async () => {
        await sleep(2000)

        return {
            ...SERIES_MOCK_CHART,
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
}

export const SearchInsightCreationPage: Story = () => (
    <CodeInsightsBackendStoryMock mocks={codeInsightsBackend}>
        <SearchInsightCreationPageComponent
            telemetryService={NOOP_TELEMETRY_SERVICE}
            onInsightCreateRequest={fakeAPIRequest}
            onSuccessfulCreation={noop}
            onCancel={noop}
        />
    </CodeInsightsBackendStoryMock>
)
