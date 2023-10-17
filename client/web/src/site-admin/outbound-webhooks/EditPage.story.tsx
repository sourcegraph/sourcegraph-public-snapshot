import type { Decorator, Meta, StoryFn } from '@storybook/react'
import { WildcardMockLink } from 'wildcard-mock-link'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../components/WebStory'

import { EditPage } from './EditPage'
import { logConnectionLink, buildOutboundWebhookMock, eventTypesMock } from './mocks'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/site-admin/webhooks/outgoing/EditPage',
    decorators: [decorator],
}

export default config

export const Page: StoryFn = () => (
    <WebStory>
        {() => (
            <MockedTestProvider
                link={new WildcardMockLink([logConnectionLink, buildOutboundWebhookMock(''), eventTypesMock])}
            >
                <EditPage telemetryService={NOOP_TELEMETRY_SERVICE} />
            </MockedTestProvider>
        )}
    </WebStory>
)

Page.storyName = 'Page'
