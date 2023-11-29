import type { Decorator, Meta, StoryFn } from '@storybook/react'
import { WildcardMockLink } from 'wildcard-mock-link'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../components/WebStory'

import { buildOutboundWebhooksConnectionLink } from './mocks'
import { OutboundWebhooksPage } from './OutboundWebhooksPage'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/site-admin/webhooks/outgoing/OutboundWebhooksPage',
    decorators: [decorator],
}

export default config

export const Empty: StoryFn = () => (
    <WebStory>
        {() => (
            <MockedTestProvider link={new WildcardMockLink([buildOutboundWebhooksConnectionLink(0)])}>
                <OutboundWebhooksPage telemetryService={NOOP_TELEMETRY_SERVICE} />
            </MockedTestProvider>
        )}
    </WebStory>
)

Empty.storyName = 'Empty'

export const NotEmpty: StoryFn = () => (
    <WebStory>
        {() => (
            <MockedTestProvider link={new WildcardMockLink([buildOutboundWebhooksConnectionLink(20)])}>
                <OutboundWebhooksPage telemetryService={NOOP_TELEMETRY_SERVICE} />
            </MockedTestProvider>
        )}
    </WebStory>
)

NotEmpty.storyName = 'Not empty'
