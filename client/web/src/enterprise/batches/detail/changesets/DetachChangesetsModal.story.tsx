import { action } from '@storybook/addon-actions'
import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'
import React from 'react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../../components/WebStory'

import { DetachChangesetsModal } from './DetachChangesetsModal'

const { add } = storiesOf('web/batches/details/DetachChangesetsModal', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

const detachAction = () => {
    action('DetachChangesets')
    return Promise.resolve()
}

add('Confirmation', () => (
    <WebStory>
        {props => (
            <DetachChangesetsModal
                {...props}
                afterCreate={noop}
                batchChangeID="test-123"
                changesetIDs={['test-123', 'test-234']}
                onCancel={noop}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                detachChangesets={detachAction}
            />
        )}
    </WebStory>
))
