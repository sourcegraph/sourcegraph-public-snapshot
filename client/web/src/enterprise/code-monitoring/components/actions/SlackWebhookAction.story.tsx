import { storiesOf } from '@storybook/react'
import React from 'react'
import sinon from 'sinon'

import { WebStory } from '../../../../components/WebStory'
import { ActionProps } from '../FormActionArea'

import { SlackWebhookAction } from './SlackWebhookAction'

const { add } = storiesOf('web/enterprise/code-monitoring/actions/SlackWebhookAction', module)

const defaultProps: ActionProps = {
    action: undefined,
    setAction: sinon.fake(),
    disabled: false,
    monitorName: 'Example code monitor',
}

const action: ActionProps['action'] = {
    __typename: 'MonitorSlackWebhook',
    id: 'id1',
    url: 'https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX',
    enabled: true,
}

add('disabled', () => <WebStory>{() => <SlackWebhookAction {...defaultProps} disabled={true} />}</WebStory>)

add('closed, not populated', () => <WebStory>{() => <SlackWebhookAction {...defaultProps} />}</WebStory>)

add('open, not populated', () => (
    <WebStory>{() => <SlackWebhookAction {...defaultProps} _testStartOpen={true} />}</WebStory>
))

add('closed, populated, enabled', () => (
    <WebStory>{() => <SlackWebhookAction {...defaultProps} action={action} />}</WebStory>
))

add('open, populated, enabled', () => (
    <WebStory>{() => <SlackWebhookAction {...defaultProps} _testStartOpen={true} action={action} />}</WebStory>
))

add('closed, populated, disabled', () => (
    <WebStory>{() => <SlackWebhookAction {...defaultProps} action={{ ...action, enabled: false }} />}</WebStory>
))

add('open, populated, disabled', () => (
    <WebStory>
        {() => <SlackWebhookAction {...defaultProps} _testStartOpen={true} action={{ ...action, enabled: false }} />}
    </WebStory>
))
