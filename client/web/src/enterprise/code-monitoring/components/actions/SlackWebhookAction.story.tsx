import { storiesOf } from '@storybook/react'
import React from 'react'
import sinon from 'sinon'

import { WebStory } from '../../../../components/WebStory'
import { mockAuthenticatedUser } from '../../testing/util'
import { ActionProps } from '../FormActionArea'

import { SlackWebhookAction } from './SlackWebhookAction'

const { add } = storiesOf('web/enterprise/code-monitoring/actions/SlackWebhookAction', module).addParameters({
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
    __typename: 'MonitorSlackWebhook',
    id: 'id1',
    url: 'https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX',
    enabled: true,
    includeResults: false,
}

add('SlackWebhookAction', () => (
    <WebStory>
        {() => (
            <>
                <h2>Action card disabled</h2>
                <SlackWebhookAction {...defaultProps} disabled={true} />

                <h2>Closed, not populated</h2>
                <SlackWebhookAction {...defaultProps} />

                <h2>Open, not populated</h2>
                <SlackWebhookAction {...defaultProps} _testStartOpen={true} />

                <h2>Closed, populated, enabled</h2>
                <SlackWebhookAction {...defaultProps} action={action} />

                <h2>Open, populated, enabled</h2>
                <SlackWebhookAction {...defaultProps} _testStartOpen={true} action={action} />

                <h2>Open, populated with error, enabled</h2>
                <SlackWebhookAction
                    {...defaultProps}
                    _testStartOpen={true}
                    action={{ ...action, url: 'https://example.com' }}
                />

                <h2>Closed, populated, disabled</h2>
                <SlackWebhookAction {...defaultProps} action={{ ...action, enabled: false }} />

                <h2>Open, populated, disabled</h2>
                <SlackWebhookAction {...defaultProps} _testStartOpen={true} action={{ ...action, enabled: false }} />
            </>
        )}
    </WebStory>
))
