import React, { useState } from 'react'

import { number } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'

import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import { Container } from '@sourcegraph/wildcard'

import { WebStory } from '../../components/WebStory'

import { SelectedExternalService } from './backend'
import { buildHeaderMock } from './story/fixtures'
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

// Create a component to handle the minimum state management required for a
// WebhookLogPageHeader.
const WebhookLogPageHeaderContainer: React.FunctionComponent<
    React.PropsWithChildren<{
        initialExternalService?: SelectedExternalService
        initialOnlyErrors?: boolean
    }>
> = ({ initialExternalService, initialOnlyErrors }) => {
    const [onlyErrors, setOnlyErrors] = useState(initialOnlyErrors === true)
    const [externalService, setExternalService] = useState(initialExternalService ?? 'all')

    return (
        <WebhookLogPageHeader
            externalService={externalService}
            onlyErrors={onlyErrors}
            onSelectExternalService={setExternalService}
            onSetOnlyErrors={setOnlyErrors}
        />
    )
}

add('all zeroes', () => (
    <WebStory>
        {() => (
            <MockedTestProvider mocks={buildHeaderMock(0, 0)}>
                <WebhookLogPageHeaderContainer />
            </MockedTestProvider>
        )}
    </WebStory>
))

add('external services', () => (
    <WebStory>
        {() => (
            <MockedTestProvider mocks={buildHeaderMock(10, 0)}>
                <WebhookLogPageHeaderContainer />
            </MockedTestProvider>
        )}
    </WebStory>
))

add('external services and errors', () => (
    <WebStory>
        {() => (
            <MockedTestProvider mocks={buildHeaderMock(20, 500)}>
                <WebhookLogPageHeaderContainer />
            </MockedTestProvider>
        )}
    </WebStory>
))

add('only errors turned on', () => (
    <WebStory>
        {() => (
            <MockedTestProvider mocks={buildHeaderMock(20, 500)}>
                <WebhookLogPageHeaderContainer initialOnlyErrors={true} />
            </MockedTestProvider>
        )}
    </WebStory>
))

add('specific external service selected', () => (
    <WebStory>
        {() => (
            <MockedTestProvider mocks={buildHeaderMock(20, 500)}>
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
            <MockedTestProvider mocks={buildHeaderMock(20, 500)}>
                <WebhookLogPageHeaderContainer initialExternalService="unmatched" />
            </MockedTestProvider>
        )}
    </WebStory>
))
