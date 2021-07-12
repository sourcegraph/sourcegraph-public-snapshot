import { action } from '@storybook/addon-actions'
import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'
import React from 'react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { EnterpriseWebStory } from '../../../components/EnterpriseWebStory'

import { DetachChangesetsModal } from './DetachChangesetsModal'

const { add } = storiesOf('web/batches/details/DetachChangesetsModal', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

const changesetIDsFunction = () => Promise.resolve(['test-123', 'test-234'])
const detachAction = () => {
    action('DetachChangesets')
    return Promise.resolve()
}

add('Confirmation', () => (
    <EnterpriseWebStory>
        {props => (
            <DetachChangesetsModal
                {...props}
                afterCreate={noop}
                batchChangeID="test-123"
                changesetIDs={changesetIDsFunction}
                onCancel={noop}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                detachChangesets={detachAction}
            />
        )}
    </EnterpriseWebStory>
))
