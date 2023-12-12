import type { DecoratorFn, Story, Meta } from '@storybook/react'

import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../WebStory'

import { AddExternalServicesPage } from './AddExternalServicesPage'
import { codeHostExternalServices, nonCodeHostExternalServices } from './externalServices'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/External services/AddExternalServicesPage',
    decorators: [decorator],
    parameters: {
        chromatic: {
            // Delay screenshot taking, so Monaco has some time to get syntax highlighting prepared.
            delay: 2000,
        },
    },
}

export default config

export const Overview: Story = () => (
    <WebStory>
        {webProps => (
            <AddExternalServicesPage
                {...webProps}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                telemetryRecorder={noOpTelemetryRecorder}
                codeHostExternalServices={codeHostExternalServices}
                nonCodeHostExternalServices={nonCodeHostExternalServices}
                autoFocusForm={false}
                externalServicesFromFile={false}
                allowEditExternalServicesWithFile={false}
                isCodyApp={false}
            />
        )}
    </WebStory>
)

export const OverviewWithBusinessLicense: Story = () => {
    window.context.licenseInfo = { currentPlan: 'business-0' }
    return (
        <WebStory>
            {webProps => (
                <AddExternalServicesPage
                    {...webProps}
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                    telemetryRecorder={noOpTelemetryRecorder}
                    codeHostExternalServices={codeHostExternalServices}
                    nonCodeHostExternalServices={nonCodeHostExternalServices}
                    autoFocusForm={false}
                    externalServicesFromFile={false}
                    allowEditExternalServicesWithFile={false}
                    isCodyApp={false}
                />
            )}
        </WebStory>
    )
}

export const AddConnectionBykind: Story = () => (
    <WebStory initialEntries={['/page?id=github']}>
        {webProps => (
            <AddExternalServicesPage
                {...webProps}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                telemetryRecorder={noOpTelemetryRecorder}
                codeHostExternalServices={codeHostExternalServices}
                nonCodeHostExternalServices={nonCodeHostExternalServices}
                autoFocusForm={false}
                externalServicesFromFile={false}
                allowEditExternalServicesWithFile={false}
                isCodyApp={false}
            />
        )}
    </WebStory>
)

AddConnectionBykind.storyName = 'Add connection by kind'
