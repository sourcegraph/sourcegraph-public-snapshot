import type { ComponentProps } from 'react'

import type { Meta, StoryFn } from '@storybook/react'

import { MockedStoryProvider } from '@sourcegraph/shared/src/stories'
import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'

import { WebStory } from '../components/WebStory'

import { MOCK_REQUESTS } from './graphql.mocks'
import { NewForm } from './NewForm'

const config: Meta = {
    title: 'web/savedSearches/NewForm',
    component: NewForm,
    decorators: [story => <div className="container mt-5">{story()}</div>],
    parameters: {},
}

export default config

const commonProps: ComponentProps<typeof NewForm> = {
    isSourcegraphDotCom: false,
    telemetryRecorder: noOpTelemetryRecorder,
}

export const Default: StoryFn = () => (
    <WebStory>
        {webProps => (
            <MockedStoryProvider mocks={MOCK_REQUESTS}>
                <NewForm {...webProps} {...commonProps} />
            </MockedStoryProvider>
        )}
    </WebStory>
)
