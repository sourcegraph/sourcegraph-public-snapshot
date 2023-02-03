import { DecoratorFn, Meta, Story } from '@storybook/react'
import { WildcardMockLink } from 'wildcard-mock-link'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../components/WebStory'

import { buildOutboundWebhooksConnectionLink } from './mocks'
import { OutboundWebhooksPage } from './OutboundWebhooksPage'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/site-admin/outbound-webhooks/OutboundWebhooksPage',
    decorators: [decorator],
}

export default config

export const Empty: Story = () => (
    <WebStory>
        {() => (
            <MockedTestProvider link={new WildcardMockLink([buildOutboundWebhooksConnectionLink(0)])}>
                <OutboundWebhooksPage telemetryService={NOOP_TELEMETRY_SERVICE} />
            </MockedTestProvider>
        )}
    </WebStory>
)

Empty.storyName = 'Empty'

export const NotEmpty: Story = () => (
    <WebStory>
        {() => (
            <MockedTestProvider link={new WildcardMockLink([buildOutboundWebhooksConnectionLink(20)])}>
                <OutboundWebhooksPage telemetryService={NOOP_TELEMETRY_SERVICE} />
            </MockedTestProvider>
        )}
    </WebStory>
)

NotEmpty.storyName = 'Not empty'
