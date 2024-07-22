import type { ComponentProps } from 'react'

import type { Meta, StoryFn } from '@storybook/react'

import { MockedStoryProvider } from '@sourcegraph/shared/src/stories'
import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'

import { WebStory } from '../components/WebStory'

import { EditPage } from './EditPage'
import { MOCK_REQUESTS } from './graphql.mocks'

const config: Meta = {
    title: 'web/savedSearches/EditPage',
    component: EditPage,
    decorators: [story => <div className="container mt-5">{story()}</div>],
    parameters: {},
}

export default config

const commonProps: ComponentProps<typeof EditPage> = {
    isSourcegraphDotCom: false,
    telemetryRecorder: noOpTelemetryRecorder,
}

export const Default: StoryFn = () => (
    <WebStory>
        {webProps => (
            <MockedStoryProvider mocks={MOCK_REQUESTS}>
                <EditPage {...webProps} {...commonProps} />
            </MockedStoryProvider>
        )}
    </WebStory>
)
