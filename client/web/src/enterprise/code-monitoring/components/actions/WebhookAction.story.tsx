import { storiesOf } from '@storybook/react'
import React from 'react'
import sinon from 'sinon'

import { WebStory } from '../../../../components/WebStory'
import { ActionProps } from '../FormActionArea'

import { WebhookAction } from './WebhookAction'

const { add } = storiesOf('web/enterprise/code-monitoring/actions/WebhookAction', module)

const defaultProps: ActionProps = {
    action: undefined,
    setAction: sinon.fake(),
    disabled: false,
    monitorName: 'Example code monitor',
}

const action: ActionProps['action'] = {
    __typename: 'MonitorWebhook',
    id: 'id1',
    url: 'https://example.com',
    enabled: true,
}

add('disabled', () => <WebStory>{() => <WebhookAction {...defaultProps} disabled={true} />}</WebStory>)

add('closed, not populated', () => <WebStory>{() => <WebhookAction {...defaultProps} />}</WebStory>)

add('open, not populated', () => <WebStory>{() => <WebhookAction {...defaultProps} _testStartOpen={true} />}</WebStory>)

add('closed, populated, enabled', () => (
    <WebStory>{() => <WebhookAction {...defaultProps} action={action} />}</WebStory>
))

add('open, populated, enabled', () => (
    <WebStory>{() => <WebhookAction {...defaultProps} _testStartOpen={true} action={action} />}</WebStory>
))

add('closed, populated, disabled', () => (
    <WebStory>{() => <WebhookAction {...defaultProps} action={{ ...action, enabled: false }} />}</WebStory>
))

add('open, populated, disabled', () => (
    <WebStory>
        {() => <WebhookAction {...defaultProps} _testStartOpen={true} action={{ ...action, enabled: false }} />}
    </WebStory>
))
