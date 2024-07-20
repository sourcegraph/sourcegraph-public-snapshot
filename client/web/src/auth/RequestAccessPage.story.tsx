import type { Meta, StoryFn } from '@storybook/react'

import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'

import { WebStory } from '../components/WebStory'

import { RequestAccessPage } from './RequestAccessPage'

const config: Meta = {
    title: 'web/auth/RequestAccessPage',
    parameters: {},
}

export default config

export const Default: StoryFn = () => (
    <WebStory>{() => <RequestAccessPage telemetryRecorder={noOpTelemetryRecorder} />}</WebStory>
)

export const Done: StoryFn = () => (
    <WebStory initialEntries={[{ pathname: '/done' }]}>
        {() => <RequestAccessPage telemetryRecorder={noOpTelemetryRecorder} />}
    </WebStory>
)
