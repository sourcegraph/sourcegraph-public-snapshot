import { number } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'
import React, { useState } from 'react'

import { getDocumentNode } from '@sourcegraph/shared/src/graphql/apollo'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import { WebStory } from '@sourcegraph/web/src/components/WebStory'
import { Container } from '@sourcegraph/wildcard'

import { SelectedExternalService, WEBHOOK_LOG_PAGE_HEADER } from './backend'
import { WebhookLogPageHeader } from './WebhookLogPageHeader'

const { add } = storiesOf('web/site-admin/webhooks/WebhookLogPageHeader', module)
    .addDecorator(story => (
        <Container>
            <div className="p-3 container">{story()}</div>
        </Container>
    ))
    .addParameters({
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    })

const buildExternalServices = (count: number) => {
    const services = []
    count = number('external service count', count)

    for (let index = 0; index < count; index++) {
        services.push({
            __typename: 'ExternalService',
            id: index.toString(),
            displayName: `External service ${index}`,
        })
    }

    return services
}

const buildMock = (externalServiceCount: number, webhookLogCount: number) => [
    {
        request: { query: getDocumentNode(WEBHOOK_LOG_PAGE_HEADER) },
        result: {
            data: {
                externalServices: {
                    totalCount: externalServiceCount,
                    nodes: buildExternalServices(externalServiceCount),
                },
                webhookLogs: {
                    totalCount: number('errored webhook count', webhookLogCount),
                },
            },
        },
    },
]

const WebhookLogPageHeaderContainer: React.FunctionComponent<{
    initialExternalService?: SelectedExternalService
    initialOnlyErrors?: boolean
}> = ({ initialExternalService, initialOnlyErrors }) => {
    const [onlyErrors, setOnlyErrors] = useState(initialOnlyErrors === true)
    const [externalService, setExternalService] = useState(initialExternalService ?? 'all')

    return (
        <WebhookLogPageHeader
            externalService={externalService}
            onlyErrors={onlyErrors}
            onExternalServiceSelected={setExternalService}
            onSetErrors={setOnlyErrors}
        />
    )
}

add('all zeroes', () => (
    <WebStory>
        {() => (
            <MockedTestProvider mocks={buildMock(0, 0)}>
                <WebhookLogPageHeaderContainer />
            </MockedTestProvider>
        )}
    </WebStory>
))

add('external services', () => (
    <WebStory>
        {() => (
            <MockedTestProvider mocks={buildMock(10, 0)}>
                <WebhookLogPageHeaderContainer />
            </MockedTestProvider>
        )}
    </WebStory>
))

add('external services and errors', () => (
    <WebStory>
        {() => (
            <MockedTestProvider mocks={buildMock(20, 500)}>
                <WebhookLogPageHeaderContainer />
            </MockedTestProvider>
        )}
    </WebStory>
))

add('only errors turned on', () => (
    <WebStory>
        {() => (
            <MockedTestProvider mocks={buildMock(20, 500)}>
                <WebhookLogPageHeaderContainer initialOnlyErrors={true} />
            </MockedTestProvider>
        )}
    </WebStory>
))

add('specific external service selected', () => (
    <WebStory>
        {() => (
            <MockedTestProvider mocks={buildMock(20, 500)}>
                <WebhookLogPageHeaderContainer
                    initialExternalService={number('selected external service', 2, { min: 0, max: 19 }).toString()}
                />
            </MockedTestProvider>
        )}
    </WebStory>
))

add('unmatched external service selected', () => (
    <WebStory>
        {() => (
            <MockedTestProvider mocks={buildMock(20, 500)}>
                <WebhookLogPageHeaderContainer initialExternalService="unmatched" />
            </MockedTestProvider>
        )}
    </WebStory>
))
