import type { Decorator, StoryFn, Meta } from '@storybook/react'

import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../WebStory'

import { AddExternalServicesPage } from './AddExternalServicesPage'
import { codeHostExternalServices, nonCodeHostExternalServices } from './externalServices'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/External services/AddExternalServicesPage',
    decorators: [decorator],
    parameters: {},
}

export default config

export const Overview: StoryFn = () => (
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
            />
        )}
    </WebStory>
)

export const AddConnectionBykind: StoryFn = () => (
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
            />
        )}
    </WebStory>
)

AddConnectionBykind.storyName = 'Add connection by kind'
