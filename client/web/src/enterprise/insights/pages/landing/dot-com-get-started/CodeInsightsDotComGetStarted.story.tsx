import type { Meta } from '@storybook/react'

import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../../../components/WebStory'

import { CodeInsightsDotComGetStarted } from './CodeInsightsDotComGetStarted'

export default {
    title: 'web/insights/dot-com-landing/CodeInsightsDotComGetStarted',
    decorators: [story => <WebStory>{() => story()}</WebStory>],
} as Meta

export const CodeInsightsDotComGetStartedStory = () => (
    <CodeInsightsDotComGetStarted
        telemetryService={NOOP_TELEMETRY_SERVICE}
        telemetryRecorder={noOpTelemetryRecorder}
        authenticatedUser={null}
    />
)
