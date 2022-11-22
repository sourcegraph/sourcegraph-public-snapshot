import { Meta, Story } from '@storybook/react'
import delay from 'delay'
import { noop } from 'lodash'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../../../../components/WebStory'
import { useCodeInsightsState } from '../../../../../../stores'
import { CodeInsightsBackendStoryMock } from '../../../../CodeInsightsBackendStoryMock'

import { getRandomLangStatsMock } from './components/live-preview-chart/constants'
import { LangStatsInsightCreationPage as LangStatsInsightCreationPageComponent } from './LangStatsInsightCreationPage'

const defaultStory: Meta = {
    title: 'web/insights/creation-ui/lang-stats/LangStatsInsightCreationPage',
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
}

export const LangStatsInsightCreationPage: Story = () => {
    useCodeInsightsState.setState({ licensed: true, insightsLimit: null })

    return (
        <CodeInsightsBackendStoryMock mocks={codeInsightsBackend}>
            <LangStatsInsightCreationPageComponent
                telemetryService={NOOP_TELEMETRY_SERVICE}
                onInsightCreateRequest={fakeAPIRequest}
                onSuccessfulCreation={noop}
                onCancel={noop}
            />
        </CodeInsightsBackendStoryMock>
    )
}
