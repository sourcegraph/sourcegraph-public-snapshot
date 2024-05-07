import type { Decorator, Meta, StoryFn } from '@storybook/react'

import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'

import { WebStory } from '../../components/WebStory'

import { SurveyToast } from '.'

const decorator: Decorator = story => <WebStory>{() => <div className="container mt-3">{story()}</div>}</WebStory>

const config: Meta = {
    title: 'web/SurveyToast',
    decorators: [decorator],
}

export default config

export const Toast: StoryFn = () => (
    <SurveyToast forceVisible={true} authenticatedUser={null} telemetryRecorder={noOpTelemetryRecorder} />
)
