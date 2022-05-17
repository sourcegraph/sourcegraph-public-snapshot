import { storiesOf } from '@storybook/react'
import sinon from 'sinon'

import { Typography } from '@sourcegraph/wildcard'

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
                <Typography.H2>Action card disabled</Typography.H2>
                <SlackWebhookAction {...defaultProps} disabled={true} />

                <Typography.H2>Closed, not populated</Typography.H2>
                <SlackWebhookAction {...defaultProps} />

                <Typography.H2>Open, not populated</Typography.H2>
                <SlackWebhookAction {...defaultProps} _testStartOpen={true} />

                <Typography.H2>Closed, populated, enabled</Typography.H2>
                <SlackWebhookAction {...defaultProps} action={action} />

                <Typography.H2>Open, populated, enabled</Typography.H2>
                <SlackWebhookAction {...defaultProps} _testStartOpen={true} action={action} />

                <Typography.H2>Open, populated with error, enabled</Typography.H2>
                <SlackWebhookAction
                    {...defaultProps}
                    _testStartOpen={true}
                    action={{ ...action, url: 'https://example.com' }}
                />

                <Typography.H2>Closed, populated, disabled</Typography.H2>
                <SlackWebhookAction {...defaultProps} action={{ ...action, enabled: false }} />

                <Typography.H2>Open, populated, disabled</Typography.H2>
                <SlackWebhookAction {...defaultProps} _testStartOpen={true} action={{ ...action, enabled: false }} />
            </>
        )}
    </WebStory>
))
