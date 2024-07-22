import type { Decorator, Meta, StoryFn } from '@storybook/react'
import classNames from 'classnames'

import { Container } from '@sourcegraph/wildcard'

import { WebStory } from '../../components/WebStory'
import type { WebhookLogFields } from '../../graphql-operations'

import {
    webhookLogNode,
    LARGE_HEADERS_JSON,
    LARGE_BODY_JSON,
    LARGE_HEADERS_PLAIN,
    LARGE_BODY_PLAIN,
} from './story/fixtures'
import { WebhookLogNode } from './WebhookLogNode'

import gridStyles from '../SiteAdminWebhookPage.module.scss'

const decorator: Decorator = story => (
    <Container>
        <div className={classNames('p-3', 'container', gridStyles.logs)}>{story()}</div>
    </Container>
)

const config: Meta = {
    title: 'web/site-admin/webhooks/WebhookLogNode',
    parameters: {},
    decorators: [decorator],
    argTypes: {
        receivedAt: {
            name: 'received at',
            control: { type: 'text' },
        },
        statusCode: {
            name: 'status code',
            control: { type: 'number', min: 100, max: 599 },
        },
    },
    args: {
        receivedAt: 'Sun Nov 07 2021 14:31:00 GMT-0500 (Eastern Standard Time)',
        statusCode: 204,
    },
}

export default config

// Most of the components of WebhookLogNode are more thoroughly tested elsewhere
// in the storybook, so this is just a limited number of cases to ensure the
// expando behaviour is correct, the date formatting does something useful, and
// the external service name is handled properly when there isn't an external
// service.
//
// Some bonus controls are provided for the tinkerers.

type StoryArguments = Pick<WebhookLogFields, 'receivedAt' | 'statusCode'>

export const Collapsed: StoryFn<StoryArguments> = args => (
    <WebStory>
        {() => (
            <>
                <WebhookLogNode
                    node={webhookLogNode(args, {
                        externalService: {
                            displayName: 'GitLab',
                        },
                    })}
                />
                <WebhookLogNode
                    node={webhookLogNode(args, {
                        externalService: {
                            displayName: 'BitBucket Server',
                        },
                        response: {
                            headers: LARGE_HEADERS_PLAIN,
                            body: LARGE_BODY_PLAIN,
                        },
                        request: {
                            headers: LARGE_HEADERS_JSON,
                            body: LARGE_BODY_JSON,
                            method: 'POST',
                            url: '/my/awesome/url',
                            version: 'HTTP/1.1',
                        },
                    })}
                />
            </>
        )}
    </WebStory>
)

export const ExpandedRequest: StoryFn<StoryArguments> = args => (
    <WebStory>
        {() => (
            <>
                <WebhookLogNode node={webhookLogNode(args)} initiallyExpanded={true} />
                <WebhookLogNode
                    node={webhookLogNode(args, {
                        request: {
                            headers: LARGE_HEADERS_JSON,
                            body: LARGE_BODY_JSON,
                            method: 'POST',
                            url: '/my/awesome/url',
                            version: 'HTTP/1.1',
                        },
                    })}
                    initiallyExpanded={true}
                />
            </>
        )}
    </WebStory>
)

ExpandedRequest.storyName = 'expanded request'

export const ExpandedResponse: StoryFn<StoryArguments> = args => (
    <WebStory>
        {() => (
            <>
                <WebhookLogNode node={webhookLogNode(args)} initiallyExpanded={true} initialTabIndex={1} />
                <WebhookLogNode
                    node={webhookLogNode(args, {
                        response: {
                            headers: LARGE_HEADERS_PLAIN,
                            body: LARGE_BODY_PLAIN,
                        },
                    })}
                    initiallyExpanded={true}
                    initialTabIndex={1}
                />
            </>
        )}
    </WebStory>
)

ExpandedResponse.storyName = 'expanded response'
