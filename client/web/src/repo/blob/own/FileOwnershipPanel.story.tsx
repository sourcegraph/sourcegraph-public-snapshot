import type { Meta, StoryFn } from '@storybook/react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { setupHandlers } from '@sourcegraph/shared/src/testing/msw/storybook'

import { WebStoryWithoutMocks } from '../../../components/WebStory'

import { FileOwnershipPanel } from './FileOwnershipPanel'
import { FETCH_OWNERS } from './grapqlQueries'

const handlers = setupHandlers()

const ownershipMock = handlers.mockGraphql({
    query: FETCH_OWNERS,
    mocks: {
        OwnershipConnection: () => ({
            totalOwners: 4,
            nodes: [
                {
                    owner: { __typename: 'Person', email: 'alice@example.com', displayName: '', user: null },
                    reasons: [{ __typename: 'CodeownersFileEntry' }],
                },
                {
                    owner: { __typename: 'Person', user: { username: 'bob', displayName: 'Bob the Builder' } },
                    reasons: [
                        { __typename: 'CodeownersFileEntry' },
                        { __typename: 'RecentContributorOwnershipSignal' },
                    ],
                },
                { owner: { __typename: 'Team' }, reasons: [{ __typename: 'CodeownersFileEntry' }] },
                {
                    owner: { __typename: 'Person', user: null, email: '' },
                    reasons: [
                        { __typename: 'RecentContributorOwnershipSignal' },
                        { __typename: 'RecentViewOwnershipSignal' },
                    ],
                },
                {
                    owner: { __typename: 'Person', user: null, email: '' },
                    reasons: [{ __typename: 'RecentViewOwnershipSignal' }],
                },
            ],
        }),
        RecentViewOwnershipSignal: () => ({
            title: 'Recent View',
            description: 'Associated because they have viewed this file in the last 90 days.',
        }),
        CodeownersFileEntry: () => ({
            title: 'CodeOwner',
            description: 'This person is listed in the CODEOWNERS file',
            codeownersFile: {
                __typename: 'VirtualFile',
                url: '/own',
            },
        }),
        RecentContributorOwnershipSignal: () => ({
            title: 'Recent Contributor',
            description: 'Associated because they have contributed to this file in the last 90 days',
        }),
    },
})

const config: Meta = {
    title: 'web/repo/blob/own/FileOwnership',
    parameters: {
        chromatic: { disableSnapshot: false },
    },
}

export default config

export const Default: StoryFn = () => (
    <WebStoryWithoutMocks>
        {() => (
            <FileOwnershipPanel
                repoID="github.com/sourcegraph/sourcegraph"
                filePath="README.md"
                telemetryService={NOOP_TELEMETRY_SERVICE}
            />
        )}
    </WebStoryWithoutMocks>
)
Default.parameters = {
    msw: {
        handlers: [ownershipMock],
    },
}
