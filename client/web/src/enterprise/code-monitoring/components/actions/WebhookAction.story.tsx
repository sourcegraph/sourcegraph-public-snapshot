import { storiesOf } from '@storybook/react'
import React from 'react'
import sinon from 'sinon'

import { WebStory } from '../../../../components/WebStory'
import { mockAuthenticatedUser } from '../../testing/util'
import { ActionProps } from '../FormActionArea'

import { WebhookAction } from './WebhookAction'

const { add } = storiesOf('web/enterprise/code-monitoring/actions/WebhookAction', module).addParameters({
    chromatic: { disableSnapshot: false },
})

const defaultProps: ActionProps = {
    action: undefined,
    setAction: sinon.fake(),
    disabled: false,
    monitorName: 'Example code monitor',
    authenticatedUser: mockAuthenticatedUser,
}

const action: ActionProps['action'] = {
    __typename: 'MonitorWebhook',
    id: 'id1',
    url: 'https://example.com',
    enabled: true,
    includeResults: false,
}

add('WebhookAction', () => (
    <WebStory>
        {() => (
            <>
                <h2>Action card disabled</h2>
                <WebhookAction {...defaultProps} disabled={true} />

                <h2>Closed, not populated</h2>
                <WebhookAction {...defaultProps} />

                <h2>Open, not populated</h2>
                <WebhookAction {...defaultProps} _testStartOpen={true} />

                <h2>Closed, populated, enabled</h2>
                <WebhookAction {...defaultProps} action={action} />

                <h2>Open, populated, enabled</h2>
                <WebhookAction {...defaultProps} _testStartOpen={true} action={action} />

                <h2>Open, populated with error, enabled</h2>
                <WebhookAction {...defaultProps} _testStartOpen={true} action={{ ...action, url: 'mailto:test' }} />

                <h2>Closed, populated, disabled</h2>
                <WebhookAction {...defaultProps} action={{ ...action, enabled: false }} />

                <h2>Open, populated, disabled</h2>
                <WebhookAction {...defaultProps} _testStartOpen={true} action={{ ...action, enabled: false }} />
            </>
        )}
    </WebStory>
))
