import type { DecoratorFn, Meta, Story } from '@storybook/react'
import { WildcardMockLink } from 'wildcard-mock-link'

import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../components/WebStory'

import { EditPage } from './EditPage'
import { logConnectionLink, buildOutboundWebhookMock, eventTypesMock } from './mocks'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/site-admin/webhooks/outgoing/EditPage',
    decorators: [decorator],
}

export default config

export const Page: Story = () => (
    <WebStory>
        {() => (
            <MockedTestProvider
                link={new WildcardMockLink([logConnectionLink, buildOutboundWebhookMock(''), eventTypesMock])}
            >
                <EditPage telemetryService={NOOP_TELEMETRY_SERVICE} telemetryRecorder={noOpTelemetryRecorder} />
            </MockedTestProvider>
        )}
    </WebStory>
)

Page.storyName = 'Page'
