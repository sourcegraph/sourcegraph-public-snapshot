import React, { useState } from 'react'

import { number } from '@storybook/addon-knobs'
import { DecoratorFn, Meta, Story } from '@storybook/react'

import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import { Container } from '@sourcegraph/wildcard'

import { WebStory } from '../../components/WebStory'

import { SelectedExternalService } from './backend'
import { buildHeaderMock } from './story/fixtures'
import { WebhookLogPageHeader } from './WebhookLogPageHeader'

const decorator: DecoratorFn = story => (
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

export const AllZeroes: Story = () => (
    <WebStory>
        {() => (
            <MockedTestProvider mocks={buildHeaderMock(0, 0)}>
                <WebhookLogPageHeaderContainer />
            </MockedTestProvider>
        )}
    </WebStory>
)

AllZeroes.storyName = 'all zeroes'

export const ExternalServices: Story = () => (
    <WebStory>
        {() => (
            <MockedTestProvider mocks={buildHeaderMock(10, 0)}>
                <WebhookLogPageHeaderContainer />
            </MockedTestProvider>
        )}
    </WebStory>
)

ExternalServices.storyName = 'external services'

export const ExternalServicesAndErrors: Story = () => (
    <WebStory>
        {() => (
            <MockedTestProvider mocks={buildHeaderMock(20, 500)}>
                <WebhookLogPageHeaderContainer />
            </MockedTestProvider>
        )}
    </WebStory>
)

ExternalServicesAndErrors.storyName = 'external services and errors'

export const OnlyErrorsTurnedOn: Story = () => (
    <WebStory>
        {() => (
            <MockedTestProvider mocks={buildHeaderMock(20, 500)}>
                <WebhookLogPageHeaderContainer initialOnlyErrors={true} />
            </MockedTestProvider>
        )}
    </WebStory>
)

OnlyErrorsTurnedOn.storyName = 'only errors turned on'

export const SpecificExternalServiceSelected: Story = () => (
    <WebStory>
        {() => (
            <MockedTestProvider mocks={buildHeaderMock(20, 500)}>
                <WebhookLogPageHeaderContainer
                    initialExternalService={number('selected external service', 2, { min: 0, max: 19 }).toString()}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)

SpecificExternalServiceSelected.storyName = 'specific external service selected'

export const UnmatchedExternalServiceSelected: Story = () => (
    <WebStory>
        {() => (
            <MockedTestProvider mocks={buildHeaderMock(20, 500)}>
                <WebhookLogPageHeaderContainer initialExternalService="unmatched" />
            </MockedTestProvider>
        )}
    </WebStory>
)

UnmatchedExternalServiceSelected.storyName = 'unmatched external service selected'
