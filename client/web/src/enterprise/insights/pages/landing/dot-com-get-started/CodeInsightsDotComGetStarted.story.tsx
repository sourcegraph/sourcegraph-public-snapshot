import { Meta } from '@storybook/react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../../../components/WebStory'

import { CodeInsightsDotComGetStarted } from './CodeInsightsDotComGetStarted'

export default {
    title: 'insights/dot-com-landing/CodeInsightsDotComGetStarted',
    decorators: [story => <WebStory>{() => story()}</WebStory>],
} as Meta

export const CodeInsightsDotComGetStartedStory = () => (
    <CodeInsightsDotComGetStarted telemetryService={NOOP_TELEMETRY_SERVICE} />
)
