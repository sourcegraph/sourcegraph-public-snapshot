import type { Decorator, Meta, StoryFn } from '@storybook/react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../components/WebStory'

import { CreatePage } from './CreatePage'
import { eventTypesMock } from './mocks'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/site-admin/webhooks/outgoing/CreatePage',
    decorators: [decorator],
}

export default config

export const Page: StoryFn = () => (
    <WebStory>
        {() => (
            <MockedTestProvider mocks={[eventTypesMock]}>
                <CreatePage telemetryService={NOOP_TELEMETRY_SERVICE} />
            </MockedTestProvider>
        )}
    </WebStory>
)

Page.storyName = 'Page'
