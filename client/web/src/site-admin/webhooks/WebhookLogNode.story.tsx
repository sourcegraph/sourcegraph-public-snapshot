import { number, text } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'
import classNames from 'classnames'
import React from 'react'

import { WebStory } from '@sourcegraph/web/src/components/WebStory'
import { Container } from '@sourcegraph/wildcard'

import { WebhookLogFields } from '../../graphql-operations'

import { BODY_JSON, BODY_PLAIN, HEADERS_JSON, HEADERS_PLAIN } from './story/fixtures'
import { WebhookLogNode } from './WebhookLogNode'
import gridStyles from './WebhookLogPage.module.scss'

const { add } = storiesOf('web/site-admin/webhooks/WebhookLogNode', module)
    .addDecorator(story => (
        <Container>
            <div className={classNames('p-3', 'container', gridStyles.logs)}>{story()}</div>
        </Container>
    ))
    .addParameters({
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    })

// Most of the components of WebhookLogNode are more thoroughly tested elsewhere
// in the storybook, so this is just a limited number of cases to ensure the
// expando behaviour is correct, the date formatting does something useful, and
// the external service name is handled properly when there isn't an external
// service.
//
// Some bonus knobs are provided for the tinkerers.

const createNode = (overrides?: Partial<WebhookLogFields>): WebhookLogFields => ({
    id: overrides?.id ?? 'ID',
    receivedAt: overrides?.receivedAt ?? text('received at', '2021-11-07T19:31:00Z'),
    statusCode: overrides?.statusCode ?? number('status code', 204, { min: 100, max: 599 }),
    externalService: overrides?.externalService ?? null,
    request: overrides?.request ?? {
        headers: HEADERS_JSON,
        body: BODY_JSON,
        method: 'POST',
        url: '/my/url',
        version: 'HTTP/1.1',
    },
    response: overrides?.response ?? {
        headers: HEADERS_PLAIN,
        body: BODY_PLAIN,
    },
})

add('collapsed', () => (
    <WebStory>
        {() => (
            <WebhookLogNode
                node={createNode({
                    externalService: {
                        displayName: 'GitLab',
                    },
                })}
            />
        )}
    </WebStory>
))
add('expanded request', () => (
    <WebStory>{() => <WebhookLogNode node={createNode()} initiallyExpanded={true} />}</WebStory>
))
add('expanded response', () => (
    <WebStory>{() => <WebhookLogNode node={createNode()} initiallyExpanded={true} initialTabIndex={1} />}</WebStory>
))
