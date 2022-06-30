import { Meta, Story, DecoratorFn } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { SearchContextCtaPrompt } from './SearchContextCtaPrompt'

const decorator: DecoratorFn = story => (
    <div className="dropdown-menu show" style={{ position: 'static' }}>
        {story()}
    </div>
)

const config: Meta = {
    title: 'search-ui/input/SearchContextCtaPrompt',
    parameters: {
        chromatic: { viewports: [500], disableSnapshot: false },
    },
    decorators: [decorator],
}

export default config

const authUser: AuthenticatedUser = {
    __typename: 'User',
    id: '0',
    email: 'alice@sourcegraph.com',
    username: 'alice',
    avatarURL: null,
    session: { canSignOut: true },
    displayName: null,
    url: '',
    settingsURL: '#',
    siteAdmin: true,
    organizations: {
        nodes: [],
    },
    tags: ['AllowUserExternalServicePublic'],
    viewerCanAdminister: true,
    databaseID: 0,
    tosAccepted: true,
    searchable: true,
    emails: [],
}

export const NotAuthenticated: Story = () => (
    <BrandedStory>
        {() => (
            <SearchContextCtaPrompt
                telemetryService={NOOP_TELEMETRY_SERVICE}
                authenticatedUser={null}
                hasUserAddedExternalServices={false}
                onDismiss={() => {}}
            />
        )}
    </BrandedStory>
)

NotAuthenticated.storyName = 'not authenticated'

export const AuthenticatedWithoutAddedExternalServices: Story = () => (
    <BrandedStory>
        {() => (
            <SearchContextCtaPrompt
                telemetryService={NOOP_TELEMETRY_SERVICE}
                authenticatedUser={authUser}
                hasUserAddedExternalServices={false}
                onDismiss={() => {}}
            />
        )}
    </BrandedStory>
)

AuthenticatedWithoutAddedExternalServices.storyName = 'authenticated without added external services'

export const AuthenticatedWithAddedExternalServices: Story = () => (
    <BrandedStory>
        {() => (
            <SearchContextCtaPrompt
                telemetryService={NOOP_TELEMETRY_SERVICE}
                authenticatedUser={authUser}
                hasUserAddedExternalServices={true}
                onDismiss={() => {}}
            />
        )}
    </BrandedStory>
)

AuthenticatedWithAddedExternalServices.storyName = 'authenticated with added external services'

export const AuthenticatedWithPrivateCode: Story = () => (
    <BrandedStory>
        {() => (
            <SearchContextCtaPrompt
                telemetryService={NOOP_TELEMETRY_SERVICE}
                authenticatedUser={{ ...authUser, tags: ['AllowUserExternalServicePrivate'] }}
                hasUserAddedExternalServices={false}
                onDismiss={() => {}}
            />
        )}
    </BrandedStory>
)

AuthenticatedWithPrivateCode.storyName = 'authenticated with private code'
