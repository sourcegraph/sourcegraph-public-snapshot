import type { MockedResponse } from '@apollo/client/testing'
import type { Meta, StoryFn } from '@storybook/react'

import { getDocumentNode } from '@sourcegraph/http-client'

import { WebStory } from '../../../components/WebStory'
import { ExternalServiceKind, type FetchOwnersAndHistoryResult, RepositoryType } from '../../../graphql-operations'

import { FETCH_OWNERS_AND_HISTORY } from './grapqlQueries'
import { HistoryAndOwnBar } from './HistoryAndOwnBar'

window.context.experimentalFeatures = { perforceChangelistMapping: 'enabled' }

const barData: FetchOwnersAndHistoryResult = {
    node: {
        sourceType: RepositoryType.GIT_REPOSITORY,
        commit: {
            blob: {
                contributors: {
                    totalCount: 0,
                },
                ownership: {
                    nodes: [
                        {
                            owner: {
                                id: 'user1',
                                avatarURL: null,
                                teamDisplayName: 'Xclaesse',
                                url: '/teams/xclaesse',
                                external: false,
                                name: 'xclaesse',
                                __typename: 'Team',
                            },
                            __typename: 'Ownership',
                        },
                        {
                            owner: {
                                email: '',
                                avatarURL: 'https://avatars.githubusercontent.com/u/5090588?v=4',
                                displayName: 'pwithnall',
                                user: {
                                    id: 'user2',
                                    displayName: 'Philip Withnall',
                                    url: '/users/pwithnall',
                                    username: 'pwithnall',
                                    primaryEmail: null,
                                },
                                __typename: 'Person',
                            },
                            __typename: 'Ownership',
                        },
                        {
                            owner: {
                                email: '',
                                avatarURL: null,
                                displayName: 'nirbheek',
                                user: null,
                                __typename: 'Person',
                            },
                            __typename: 'Ownership',
                        },
                    ],
                    totalCount: 3,
                    __typename: 'OwnershipConnection',
                },
                __typename: 'GitBlob',
            },
            ancestors: {
                nodes: [
                    {
                        id: 'R2l0Q29tbWl0OnsiciI6IlVtVndiM05wZEc5eWVUb3hNRGN3Tnc9PSIsImMiOiIxNzQxZmMyYzZlYjhlMTFmNGU3ODZjY2M1YzE5YzBkYTMyNzYzMGE1In0=',
                        oid: '1741fc2c6eb8e11f4e786ccc5c19c0da327630a5',
                        abbreviatedOID: '1741fc2',
                        perforceChangelist: null,
                        message:
                            'build: Drop use of G_DISABLE_DEPRECATED from the build system\n\nIt’s no longer used in any of the headers. See preceding commits.\n\nSigned-off-by: Philip Withnall \u003Cwithnall@endlessm.com\u003E',
                        subject: 'build: Drop use of G_DISABLE_DEPRECATED from the build system',
                        body: 'It’s no longer used in any of the headers. See preceding commits.\n\nSigned-off-by: Philip Withnall \u003Cwithnall@endlessm.com\u003E',
                        author: {
                            person: {
                                avatarURL: null,
                                name: 'Philip Withnall',
                                email: 'withnall@endlessm.com',
                                displayName: 'Philip Withnall',
                                user: null,
                                __typename: 'Person',
                            },
                            date: '2019-05-27T17:19:07Z',
                            __typename: 'Signature',
                        },
                        committer: {
                            person: {
                                avatarURL: null,
                                name: 'Philip Withnall',
                                email: 'withnall@endlessm.com',
                                displayName: 'Philip Withnall',
                                user: null,
                                __typename: 'Person',
                            },
                            date: '2019-05-30T09:38:45Z',
                            __typename: 'Signature',
                        },
                        parents: [
                            {
                                oid: '99b412bb192c0062753cbf960169b1f99335080f',
                                abbreviatedOID: '99b412b',
                                perforceChangelist: null,
                                url: '/ghe.sgdev.org/sourcegraph/GNOME-glib/-/commit/99b412bb192c0062753cbf960169b1f99335080f',
                                __typename: 'GitCommit',
                            },
                        ],
                        url: '/ghe.sgdev.org/sourcegraph/GNOME-glib/-/commit/1741fc2c6eb8e11f4e786ccc5c19c0da327630a5',
                        canonicalURL:
                            '/ghe.sgdev.org/sourcegraph/GNOME-glib/-/commit/1741fc2c6eb8e11f4e786ccc5c19c0da327630a5',
                        externalURLs: [
                            {
                                url: 'https://ghe.sgdev.org/sourcegraph/GNOME-glib/commit/1741fc2c6eb8e11f4e786ccc5c19c0da327630a5',
                                serviceKind: ExternalServiceKind.GITHUB,
                                __typename: 'ExternalLink',
                            },
                        ],
                        tree: {
                            canonicalURL:
                                '/ghe.sgdev.org/sourcegraph/GNOME-glib@1741fc2c6eb8e11f4e786ccc5c19c0da327630a5',
                            __typename: 'GitTree',
                        },
                        __typename: 'GitCommit',
                    },
                ],
                __typename: 'GitCommitConnection',
            },
            __typename: 'GitCommit',
        },
        changelist: null,
        __typename: 'Repository',
    },
}

const variables = { repoID: 'VXNlcjox', filePath: 'test.tsx' }

const mockLoaded: MockedResponse = {
    request: {
        query: getDocumentNode(FETCH_OWNERS_AND_HISTORY),
        variables,
    },
    result: { data: barData },
}

const config: Meta = {
    title: 'web/repo/blob/own/HistoryAndOwnBar',
    parameters: {
        chromatic: { disableSnapshot: false },
    },
}

export default config

export const Default: StoryFn = () => (
    <WebStory mocks={[mockLoaded]}>{() => <HistoryAndOwnBar enableOwnershipPanel={true} {...variables} />}</WebStory>
)
