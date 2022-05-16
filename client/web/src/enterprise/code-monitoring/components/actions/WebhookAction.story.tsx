import { storiesOf } from '@storybook/react'
import sinon from 'sinon'

import { Typography } from '@sourcegraph/wildcard'

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
                <Typography.H2>Action card disabled</Typography.H2>
                <WebhookAction {...defaultProps} disabled={true} />

                <Typography.H2>Closed, not populated</Typography.H2>
                <WebhookAction {...defaultProps} />

                <Typography.H2>Open, not populated</Typography.H2>
                <WebhookAction {...defaultProps} _testStartOpen={true} />

                <Typography.H2>Closed, populated, enabled</Typography.H2>
                <WebhookAction {...defaultProps} action={action} />

                <Typography.H2>Open, populated, enabled</Typography.H2>
                <WebhookAction {...defaultProps} _testStartOpen={true} action={action} />

                <Typography.H2>Open, populated with error, enabled</Typography.H2>
                <WebhookAction {...defaultProps} _testStartOpen={true} action={{ ...action, url: 'mailto:test' }} />

                <Typography.H2>Closed, populated, disabled</Typography.H2>
                <WebhookAction {...defaultProps} action={{ ...action, enabled: false }} />

                <Typography.H2>Open, populated, disabled</Typography.H2>
                <WebhookAction {...defaultProps} _testStartOpen={true} action={{ ...action, enabled: false }} />
            </>
        )}
    </WebStory>
))
