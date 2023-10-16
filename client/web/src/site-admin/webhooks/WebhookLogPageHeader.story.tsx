import React, { useState } from 'react'

import type { Decorator, Meta, StoryFn } from '@storybook/react'

import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import { Container } from '@sourcegraph/wildcard'

import { WebStory } from '../../components/WebStory'

import type { SelectedExternalService } from './backend'
import { buildHeaderMock } from './story/fixtures'
import { WebhookLogPageHeader } from './WebhookLogPageHeader'

const decorator: Decorator = story => (
    <Container>
        <div className="p-3 container">{story()}</div>
    </Container>
)

const config: Meta = {
    title: 'web/site-admin/webhooks/WebhookLogPageHeader',
    parameters: {
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    },
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

export const AllZeroes: StoryFn = args => (
    <WebStory>
        {() => (
            <MockedTestProvider mocks={buildHeaderMock(args.externalServiceCount, args.erroredWebhookCount)}>
                <WebhookLogPageHeaderContainer />
            </MockedTestProvider>
        )}
    </WebStory>
)
AllZeroes.argTypes = {
    externalServiceCount: {},
    erroredWebhookCount: {},
}
AllZeroes.args = {
    externalServiceCount: 0,
    erroredWebhookCount: 0,
}

AllZeroes.storyName = 'all zeroes'

export const ExternalServices: StoryFn = args => (
    <WebStory>
        {() => (
            <MockedTestProvider mocks={buildHeaderMock(args.externalServiceCount, args.erroredWebhookCount)}>
                <WebhookLogPageHeaderContainer />
            </MockedTestProvider>
        )}
    </WebStory>
)

ExternalServices.argTypes = {
    externalServiceCount: {},
    erroredWebhookCount: {},
}
ExternalServices.args = {
    externalServiceCount: 10,
    erroredWebhookCount: 0,
}

ExternalServices.storyName = 'external services'

export const ExternalServicesAndErrors: StoryFn = args => (
    <WebStory>
        {() => (
            <MockedTestProvider mocks={buildHeaderMock(args.externalServiceCount, args.erroredWebhookCount)}>
                <WebhookLogPageHeaderContainer />
            </MockedTestProvider>
        )}
    </WebStory>
)

ExternalServicesAndErrors.argTypes = {
    externalServiceCount: {},
    erroredWebhookCount: {},
}
ExternalServicesAndErrors.args = {
    externalServiceCount: 20,
    erroredWebhookCount: 500,
}

ExternalServicesAndErrors.storyName = 'external services and errors'

export const OnlyErrorsTurnedOn: StoryFn = args => (
    <WebStory>
        {() => (
            <MockedTestProvider mocks={buildHeaderMock(args.externalServiceCount, args.erroredWebhookCount)}>
                <WebhookLogPageHeaderContainer initialOnlyErrors={true} />
            </MockedTestProvider>
        )}
    </WebStory>
)
OnlyErrorsTurnedOn.argTypes = {
    externalServiceCount: {},
    erroredWebhookCount: {},
}
OnlyErrorsTurnedOn.args = {
    externalServiceCount: 20,
    erroredWebhookCount: 500,
}

OnlyErrorsTurnedOn.storyName = 'only errors turned on'

export const SpecificExternalServiceSelected: StoryFn = args => (
    <WebStory>
        {() => (
            <MockedTestProvider mocks={buildHeaderMock(args.externalServiceCount, args.erroredWebhookCount)}>
                <WebhookLogPageHeaderContainer initialExternalService={args.initialExternalService.toString()} />
            </MockedTestProvider>
        )}
    </WebStory>
)
SpecificExternalServiceSelected.argTypes = {
    initialExternalService: {
        control: { type: 'number', min: 0, max: 19 },
    },
    externalServiceCount: {},
    erroredWebhookCount: {},
}
SpecificExternalServiceSelected.args = {
    initialExternalService: 2,
    externalServiceCount: 20,
    erroredWebhookCount: 500,
}

SpecificExternalServiceSelected.storyName = 'specific external service selected'

export const UnmatchedExternalServiceSelected: StoryFn = args => (
    <WebStory>
        {() => (
            <MockedTestProvider mocks={buildHeaderMock(args.externalServiceCount, args.erroredWebhookCount)}>
                <WebhookLogPageHeaderContainer initialExternalService="unmatched" />
            </MockedTestProvider>
        )}
    </WebStory>
)

UnmatchedExternalServiceSelected.argTypes = {
    externalServiceCount: {},
    erroredWebhookCount: {},
}
UnmatchedExternalServiceSelected.args = {
    externalServiceCount: 20,
    erroredWebhookCount: 500,
}

UnmatchedExternalServiceSelected.storyName = 'unmatched external service selected'
