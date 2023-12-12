import { action } from '@storybook/addon-actions'
import type { Story, Meta, DecoratorFn } from '@storybook/react'
import { noop } from 'lodash'

import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../../components/WebStory'

import { DetachChangesetsModal } from './DetachChangesetsModal'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/details/DetachChangesetsModal',
    decorators: [decorator],
}

export default config

const detachAction = () => {
    action('DetachChangesets')
    return Promise.resolve()
}

export const Confirmation: Story = () => (
    <WebStory>
        {props => (
            <DetachChangesetsModal
                {...props}
                afterCreate={noop}
                batchChangeID="test-123"
                changesetIDs={['test-123', 'test-234']}
                onCancel={noop}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                telemetryRecorder={noOpTelemetryRecorder}
                detachChangesets={detachAction}
            />
        )}
    </WebStory>
)
