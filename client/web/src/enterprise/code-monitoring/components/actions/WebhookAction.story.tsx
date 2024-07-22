import type { Meta, StoryFn } from '@storybook/react'
import sinon from 'sinon'

import { H2 } from '@sourcegraph/wildcard'

import { WebStory } from '../../../../components/WebStory'
import { mockAuthenticatedUser } from '../../testing/util'
import type { ActionProps } from '../FormActionArea'

import { WebhookAction } from './WebhookAction'

const config: Meta = {
    title: 'web/enterprise/code-monitoring/actions/WebhookAction',
    parameters: {},
}

export default config

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

export const WebhookActionStory: StoryFn = () => (
    <WebStory>
        {() => (
            <>
                <H2>Action card disabled</H2>
                <WebhookAction {...defaultProps} disabled={true} />

                <H2>Closed, not populated</H2>
                <WebhookAction {...defaultProps} />

                <H2>Open, not populated</H2>
                <WebhookAction {...defaultProps} _testStartOpen={true} />

                <H2>Closed, populated, enabled</H2>
                <WebhookAction {...defaultProps} action={action} />

                <H2>Open, populated, enabled</H2>
                <WebhookAction {...defaultProps} _testStartOpen={true} action={action} />

                <H2>Open, populated with error, enabled</H2>
                <WebhookAction {...defaultProps} _testStartOpen={true} action={{ ...action, url: 'mailto:test' }} />

                <H2>Closed, populated, disabled</H2>
                <WebhookAction {...defaultProps} action={{ ...action, enabled: false }} />

                <H2>Open, populated, disabled</H2>
                <WebhookAction {...defaultProps} _testStartOpen={true} action={{ ...action, enabled: false }} />
            </>
        )}
    </WebStory>
)

WebhookActionStory.storyName = 'WebhookAction'
