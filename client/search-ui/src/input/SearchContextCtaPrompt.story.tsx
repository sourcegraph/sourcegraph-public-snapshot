import { storiesOf } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { SearchContextCtaPrompt } from './SearchContextCtaPrompt'

const { add } = storiesOf('search-ui/input/SearchContextCtaPrompt', module)
    .addParameters({
        chromatic: { viewports: [500], disableSnapshot: false },
    })
    .addDecorator(story => (
        <div className="dropdown-menu show" style={{ position: 'static' }}>
            {story()}
        </div>
    ))

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
}

add(
    'not authenticated',
    () => (
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
    ),
    {}
)

add(
    'authenticated without added external services',
    () => (
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
    ),
    {}
)

add(
    'authenticated with added external services',
    () => (
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
    ),
    {}
)

add(
    'authenticated with private code',
    () => (
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
    ),
    {}
)
