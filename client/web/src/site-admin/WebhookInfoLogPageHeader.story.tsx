import React, { useState } from 'react'

import { DecoratorFn, Meta, Story } from '@storybook/react'

import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import { Container } from '@sourcegraph/wildcard'

import { WebStory } from '../components/WebStory'

import { WebhookInfoLogPageHeader } from './WebhookInfoLogPageHeader'
import { SelectedExternalService } from './webhooks/backend'
import { buildHeaderMock } from './webhooks/story/fixtures'

const decorator: DecoratorFn = story => (
    <Container>
        <div className="p-3 container">{story()}</div>
    </Container>
)

const config: Meta = {
    title: 'web/src/site-admin/WebhookInfoLogPageHeader',
    decorators: [decorator],
    argTypes: {
        externalServiceCount: {
            name: 'external service count',
            control: { type: 'number' },
        },
        erroredWebhookCount: {
            name: 'errored webhook count',
            control: { type: 'number' },
        },
    },
}

export default config

// Create a component to handle the minimum state management required for a
// WebhookInfoLogPageHeader.
const WebhookInfoLogPageHeaderContainer: React.FunctionComponent<
    React.PropsWithChildren<{
        initialExternalService?: SelectedExternalService
        initialOnlyErrors?: boolean
    }>
> = ({ initialOnlyErrors }) => {
    const [onlyErrors, setOnlyErrors] = useState(initialOnlyErrors === true)

    return <WebhookInfoLogPageHeader onlyErrors={onlyErrors} onSetOnlyErrors={setOnlyErrors} />
}

export const ExternalServicesAndErrors: Story = args => (
    <WebStory>
        {() => (
            <MockedTestProvider mocks={buildHeaderMock(args.externalServiceCount, args.erroredWebhookCount)}>
                <WebhookInfoLogPageHeaderContainer />
            </MockedTestProvider>
        )}
    </WebStory>
)

ExternalServicesAndErrors.storyName = 'external services and errors'

ExternalServicesAndErrors.argTypes = {
    externalServiceCount: {
        defaultValue: 20,
    },
    erroredWebhookCount: {
        defaultValue: 500,
    },
}
