import { storiesOf } from '@storybook/react'
import sinon from 'sinon'

import { H2 } from '@sourcegraph/wildcard'

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
                <H2>Action card disabled</H2>
                <SlackWebhookAction {...defaultProps} disabled={true} />

                <H2>Closed, not populated</H2>
                <SlackWebhookAction {...defaultProps} />

                <H2>Open, not populated</H2>
                <SlackWebhookAction {...defaultProps} _testStartOpen={true} />

                <H2>Closed, populated, enabled</H2>
                <SlackWebhookAction {...defaultProps} action={action} />

                <H2>Open, populated, enabled</H2>
                <SlackWebhookAction {...defaultProps} _testStartOpen={true} action={action} />

                <H2>Open, populated with error, enabled</H2>
                <SlackWebhookAction
                    {...defaultProps}
                    _testStartOpen={true}
                    action={{ ...action, url: 'https://example.com' }}
                />

                <H2>Closed, populated, disabled</H2>
                <SlackWebhookAction {...defaultProps} action={{ ...action, enabled: false }} />

                <H2>Open, populated, disabled</H2>
                <SlackWebhookAction {...defaultProps} _testStartOpen={true} action={{ ...action, enabled: false }} />
            </>
        )}
    </WebStory>
))
