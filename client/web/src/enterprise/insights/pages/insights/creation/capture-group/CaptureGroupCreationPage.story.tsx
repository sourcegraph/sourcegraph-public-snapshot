import type { Meta, Story } from '@storybook/react'
import { noop } from 'lodash'

import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../../../../components/WebStory'
import { useCodeInsightsLicenseState } from '../../../../stores'

import { CaptureGroupCreationPage as CaptureGroupCreationPageComponent } from './CaptureGroupCreationPage'

export default {
    title: 'web/insights/creation-ui/capture-group/CaptureGroupCreationPage',
    decorators: [story => <WebStory>{() => <div className="p-3 container web-content">{story()}</div>}</WebStory>],
    parameters: {
        chromatic: {
            viewports: [576, 1440],
            disableSnapshot: false,
        },
    },
} as Meta

export const CaptureGroupCreationPage: Story = () => {
    useCodeInsightsLicenseState.setState({ licensed: true, insightsLimit: null })

    return (
        <CaptureGroupCreationPageComponent
            backUrl="/insights/create"
            telemetryService={NOOP_TELEMETRY_SERVICE}
            telemetryRecorder={noOpTelemetryRecorder}
            onSuccessfulCreation={noop}
            onInsightCreateRequest={() => Promise.resolve()}
            onCancel={noop}
        />
    )
}
