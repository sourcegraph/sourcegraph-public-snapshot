import React, { useState } from 'react'

import { MockedResponse } from '@apollo/client/testing'
import { DecoratorFn, Meta, Story } from '@storybook/react'

import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import { Container } from '@sourcegraph/wildcard'

import { WebStory } from '../components/WebStory'
import { WebhookByIDLogPageHeaderResult } from '../graphql-operations'

import { WebhookInfoLogPageHeader } from './WebhookInfoLogPageHeader'
import { SelectedExternalService, WEBHOOK_BY_ID_LOG_PAGE_HEADER } from './webhooks/backend'

const decorator: DecoratorFn = story => (
    <Container>
        <div className="p-3 container">{story()}</div>
    </Container>
)

const config: Meta = {
    title: 'web/src/site-admin/WebhookInfoLogPageHeader',
    decorators: [decorator],
    argTypes: {
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

    return <WebhookInfoLogPageHeader webhookID="1" onlyErrors={onlyErrors} onSetOnlyErrors={setOnlyErrors} />
}

export const ExternalServicesAndErrors: Story = args => (
    <WebStory>
        {() => (
            <MockedTestProvider mocks={buildHeaderMock(args.erroredWebhookCount)}>
                <WebhookInfoLogPageHeaderContainer />
            </MockedTestProvider>
        )}
    </WebStory>
)

ExternalServicesAndErrors.storyName = 'has errors'

ExternalServicesAndErrors.argTypes = {
    erroredWebhookCount: {
        defaultValue: 500,
    },
}

const buildHeaderMock = (webhookLogCount: number): MockedResponse<WebhookByIDLogPageHeaderResult>[] => [
    {
        request: {
            query: getDocumentNode(WEBHOOK_BY_ID_LOG_PAGE_HEADER),
            variables: {
                webhookID: '1',
            },
        },
        result: {
            data: {
                webhookLogs: {
                    totalCount: webhookLogCount,
                },
            },
        },
    },
]
