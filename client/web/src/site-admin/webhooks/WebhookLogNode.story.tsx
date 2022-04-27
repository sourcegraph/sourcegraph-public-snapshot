import { storiesOf } from '@storybook/react'
import classNames from 'classnames'

import { Container } from '@sourcegraph/wildcard'

import { WebStory } from '../../components/WebStory'

import { webhookLogNode } from './story/fixtures'
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

add('collapsed', () => (
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
))
add('expanded request', () => (
    <WebStory>{() => <WebhookLogNode node={webhookLogNode()} initiallyExpanded={true} />}</WebStory>
))
add('expanded response', () => (
    <WebStory>{() => <WebhookLogNode node={webhookLogNode()} initiallyExpanded={true} initialTabIndex={1} />}</WebStory>
))
