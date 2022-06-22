import { DecoratorFn, Meta, Story } from '@storybook/react'
import classNames from 'classnames'

import { Container } from '@sourcegraph/wildcard'

import { WebStory } from '../../components/WebStory'

import { webhookLogNode } from './story/fixtures'
import { WebhookLogNode } from './WebhookLogNode'

import gridStyles from './WebhookLogPage.module.scss'

const decorator: DecoratorFn = story => (
    <Container>
        <div className={classNames('p-3', 'container', gridStyles.logs)}>{story()}</div>
    </Container>
)

const config: Meta = {
    title: 'web/site-admin/webhooks/WebhookLogNode',
    parameters: {
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    },
    decorators: [decorator],
}

export default config

// Most of the components of WebhookLogNode are more thoroughly tested elsewhere
// in the storybook, so this is just a limited number of cases to ensure the
// expando behaviour is correct, the date formatting does something useful, and
// the external service name is handled properly when there isn't an external
// service.
//
// Some bonus knobs are provided for the tinkerers.

export const Collapsed: Story = () => (
    <WebStory>
        {() => (
            <WebhookLogNode
                node={webhookLogNode({
                    externalService: {
                        displayName: 'GitLab',
                    },
                })}
            />
        )}
    </WebStory>
)

export const ExpandedRequest: Story = () => (
    <WebStory>{() => <WebhookLogNode node={webhookLogNode()} initiallyExpanded={true} />}</WebStory>
)

ExpandedRequest.storyName = 'expanded request'

export const ExpandedResponse: Story = () => (
    <WebStory>{() => <WebhookLogNode node={webhookLogNode()} initiallyExpanded={true} initialTabIndex={1} />}</WebStory>
)

ExpandedResponse.storyName = 'expanded response'
