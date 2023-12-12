import type { Meta } from '@storybook/react'

import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../../../../../../components/WebStory'

import { CodeInsightsExamplesSlider } from './CodeInsightsExamplesSlider'

export default {
    title: 'web/insights/dot-com-landing/CodeInsightsExampleSlider',
    decorators: [story => <WebStory>{() => story()}</WebStory>],
} as Meta

export const CodeInsightsExampleSliderExample = () => (
    <div style={{ width: 500 }}>
        <CodeInsightsExamplesSlider
            telemetryService={NOOP_TELEMETRY_SERVICE}
            telemetryRecorder={noOpTelemetryRecorder}
        />
    </div>
)
