import { storiesOf } from '@storybook/react'
import React from 'react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { AuthenticatedUser } from '../../auth'
import { WebStory } from '../../components/WebStory'

import { SearchContextCtaPrompt } from './SearchContextCtaPrompt'

const { add } = storiesOf('web/searchContexts/SearchContextCtaPrompt', module)
    .addParameters({
        chromatic: { viewports: [500] },
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
}

add(
    'not authenticated',
    () => (
        <WebStory>
            {() => (
                <SearchContextCtaPrompt
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                    authenticatedUser={null}
                    hasUserAddedExternalServices={false}
                    onDismiss={() => {}}
                />
            )}
        </WebStory>
    ),
    {}
)

add(
    'authenticated without added external services',
    () => (
        <WebStory>
            {() => (
                <SearchContextCtaPrompt
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                    authenticatedUser={authUser}
                    hasUserAddedExternalServices={false}
                    onDismiss={() => {}}
                />
            )}
        </WebStory>
    ),
    {}
)

add(
    'authenticated with added external services',
    () => (
        <WebStory>
            {() => (
                <SearchContextCtaPrompt
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                    authenticatedUser={authUser}
                    hasUserAddedExternalServices={true}
                    onDismiss={() => {}}
                />
            )}
        </WebStory>
    ),
    {}
)

add(
    'authenticated with private code',
    () => (
        <WebStory>
            {() => (
                <SearchContextCtaPrompt
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                    authenticatedUser={{ ...authUser, tags: ['AllowUserExternalServicePrivate'] }}
                    hasUserAddedExternalServices={false}
                    onDismiss={() => {}}
                />
            )}
        </WebStory>
    ),
    {}
)
