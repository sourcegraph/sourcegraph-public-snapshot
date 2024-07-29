import type { ComponentProps } from 'react'

import type { Meta, StoryFn } from '@storybook/react'

import { MockedStoryProvider } from '@sourcegraph/shared/src/stories'
import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'

import { WebStory } from '../components/WebStory'

import { MOCK_REQUESTS } from './graphql.mocks'
import { ListPage } from './ListPage'

const config: Meta = {
    title: 'web/prompts/ListPage',
    component: ListPage,
    decorators: [story => <div className="container mt-5">{story()}</div>],
    parameters: {},
}

export default config

const commonProps: ComponentProps<typeof ListPage> = {
    telemetryRecorder: noOpTelemetryRecorder,
}

export const Default: StoryFn = () => (
    <WebStory>
        {webProps => (
            <MockedStoryProvider mocks={MOCK_REQUESTS}>
                <ListPage {...webProps} {...commonProps} />
            </MockedStoryProvider>
        )}
    </WebStory>
)
