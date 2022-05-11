import { Meta } from '@storybook/react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../../../../../../components/WebStory'

import { CodeInsightsExamplesSlider } from './CodeInsightsExamplesSlider'

export default {
    title: 'insights/dot-com-landing/CodeInsightsExampleSlider',
    decorators: [story => <WebStory>{() => story()}</WebStory>],
} as Meta

export const CodeInsightsExampleSliderExample = () => (
    <div style={{ width: 500 }}>
        <CodeInsightsExamplesSlider telemetryService={NOOP_TELEMETRY_SERVICE} />
    </div>
)
