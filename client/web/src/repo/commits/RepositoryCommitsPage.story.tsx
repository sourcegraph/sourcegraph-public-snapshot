import { MockedResponse } from '@apollo/client/testing'
import { DecoratorFn, Meta, Story } from '@storybook/react'

import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../components/WebStory'
import { RepositoryFields, RepositoryGitCommitsResult, RepositoryGitCommitsVariables } from '../../graphql-operations'

import {
    REPOSITORY_GIT_COMMITS_QUERY,
    RepositoryCommitsPage,
    RepositoryCommitsPageProps,
} from './RepositoryCommitsPage'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/RepositoryCommitsPage',
    decorators: [decorator],
}

export default config

const mockRepositoryGitCommitsQuery: MockedResponse<RepositoryGitCommitsResult, RepositoryGitCommitsVariables> = {
    request: {
        query: getDocumentNode(REPOSITORY_GIT_COMMITS_QUERY),
        variables: {
            repo: 'UmVwb3NpdG9yeToyNjM4OQ==',
            revspec: '',
            filePath: '',
            first: 20,
            afterCursor: null,
        },
    },
    result: {
        data: {
            node: {
                isPerforceDepot: false,
                externalURLs: [
                    {
                        url: 'https://github.com/sourcegraph/sourcegraph',
                        serviceKind: 'GITHUB',
                        __typename: 'ExternalLink',
                    },
                ],
                commit: {
                    ancestors: {
                        nodes: [
                            {
                                id: 'R2l0Q29tbWl0OnsiciI6IlVtVndiM05wZEc5eWVUbzMiLCJjIjoiZmFiNmQ4MzQ0YTI4ZGVlNmFlMDlhOGNiMWFlNDk5NzNmMTE3MDU0YiJ9',
                                oid: 'fab6d8344a28dee6ae09a8cb1ae49973f117054b',
                                abbreviatedOID: 'fab6d83',
                                message:
                                    'batches: add changeset cell for displaying internal state (#50366)\n\nCo-authored-by: Kelli Rockwell \u003Ckelli@sourcegraph.com\u003E\r',
                                subject: 'batches: add changeset cell for displaying internal state (#50366)',
                                body: 'Co-authored-by: Kelli Rockwell \u003Ckelli@sourcegraph.com\u003E',
                                author: {
                                    person: {
                                        avatarURL: null,
                                        name: 'Bolaji Olajide',
                                        email: '25608335+BolajiOlajide@users.noreply.github.com',
                                        displayName: 'Bolaji Olajide',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-27T16:12:53Z',
                                    __typename: 'Signature',
                                },
                                committer: {
                                    person: {
                                        avatarURL: null,
                                        name: 'GitHub',
                                        email: 'noreply@github.com',
                                        displayName: 'GitHub',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-27T16:12:53Z',
                                    __typename: 'Signature',
                                },
                                parents: [
                                    {
                                        oid: 'fb61a539c3a10075935ca78bf1334aa0260af040',
                                        abbreviatedOID: 'fb61a53',
                                        url: '/github.com/sourcegraph/sourcegraph/-/commit/fb61a539c3a10075935ca78bf1334aa0260af040',
                                        __typename: 'GitCommit',
                                    },
                                ],
                                url: '/github.com/sourcegraph/sourcegraph/-/commit/fab6d8344a28dee6ae09a8cb1ae49973f117054b',
                                canonicalURL:
                                    '/github.com/sourcegraph/sourcegraph/-/commit/fab6d8344a28dee6ae09a8cb1ae49973f117054b',
                                externalURLs: [
                                    {
                                        url: 'https://github.com/sourcegraph/sourcegraph/commit/fab6d8344a28dee6ae09a8cb1ae49973f117054b',
                                        serviceKind: 'GITHUB',
                                        __typename: 'ExternalLink',
                                    },
                                ],
                                tree: {
                                    canonicalURL:
                                        '/github.com/sourcegraph/sourcegraph@fab6d8344a28dee6ae09a8cb1ae49973f117054b',
                                    __typename: 'GitTree',
                                },
                                __typename: 'GitCommit',
                            },
                            {
                                id: 'R2l0Q29tbWl0OnsiciI6IlVtVndiM05wZEc5eWVUbzMiLCJjIjoiZmI2MWE1MzljM2ExMDA3NTkzNWNhNzhiZjEzMzRhYTAyNjBhZjA0MCJ9',
                                oid: 'fb61a539c3a10075935ca78bf1334aa0260af040',
                                abbreviatedOID: 'fb61a53',
                                message:
                                    'Cody: ✨ Suggest follow-up topics (#51201)\n\nThis adds a new recipe that is used to suggest up to three follow-up\r\ntopics for Cody. The recipe is executed with every user chat message but\r\nwill not wait for the answer so that the experience is better (this is\r\nin line with comparable chat bots).\r\n\r\n## ToDo\r\n\r\n- [x] Add behind a feature flag for now \r\n\r\n## Test plan\r\n\r\n\r\n\r\nhttps://user-images.githubusercontent.com/458591/234884949-cd71893a-ee12-408f-8d7f-b6ca76497b66.mov\r\n\r\n\r\n\r\n\u003C!-- All pull requests REQUIRE a test plan:\r\nhttps://docs.sourcegraph.com/dev/background-information/testing_principles\r\n--\u003E',
                                subject: 'Cody: ✨ Suggest follow-up topics (#51201)',
                                body: 'This adds a new recipe that is used to suggest up to three follow-up\r\ntopics for Cody. The recipe is executed with every user chat message but\r\nwill not wait for the answer so that the experience is better (this is\r\nin line with comparable chat bots).\r\n\r\n## ToDo\r\n\r\n- [x] Add behind a feature flag for now \r\n\r\n## Test plan\r\n\r\n\r\n\r\nhttps://user-images.githubusercontent.com/458591/234884949-cd71893a-ee12-408f-8d7f-b6ca76497b66.mov\r\n\r\n\r\n\r\n\u003C!-- All pull requests REQUIRE a test plan:\r\nhttps://docs.sourcegraph.com/dev/background-information/testing_principles\r\n--\u003E',
                                author: {
                                    person: {
                                        avatarURL: null,
                                        name: 'Philipp Spiess',
                                        email: 'hello@philippspiess.com',
                                        displayName: 'Philipp Spiess',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-27T16:02:51Z',
                                    __typename: 'Signature',
                                },
                                committer: {
                                    person: {
                                        avatarURL: null,
                                        name: 'GitHub',
                                        email: 'noreply@github.com',
                                        displayName: 'GitHub',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-27T16:02:51Z',
                                    __typename: 'Signature',
                                },
                                parents: [
                                    {
                                        oid: '54605d8bd456670e0bd9d2b4295851c0240fcf9c',
                                        abbreviatedOID: '54605d8',
                                        url: '/github.com/sourcegraph/sourcegraph/-/commit/54605d8bd456670e0bd9d2b4295851c0240fcf9c',
                                        __typename: 'GitCommit',
                                    },
                                ],
                                url: '/github.com/sourcegraph/sourcegraph/-/commit/fb61a539c3a10075935ca78bf1334aa0260af040',
                                canonicalURL:
                                    '/github.com/sourcegraph/sourcegraph/-/commit/fb61a539c3a10075935ca78bf1334aa0260af040',
                                externalURLs: [
                                    {
                                        url: 'https://github.com/sourcegraph/sourcegraph/commit/fb61a539c3a10075935ca78bf1334aa0260af040',
                                        serviceKind: 'GITHUB',
                                        __typename: 'ExternalLink',
                                    },
                                ],
                                tree: {
                                    canonicalURL:
                                        '/github.com/sourcegraph/sourcegraph@fb61a539c3a10075935ca78bf1334aa0260af040',
                                    __typename: 'GitTree',
                                },
                                __typename: 'GitCommit',
                            },
                            {
                                id: 'R2l0Q29tbWl0OnsiciI6IlVtVndiM05wZEc5eWVUbzMiLCJjIjoiNTQ2MDVkOGJkNDU2NjcwZTBiZDlkMmI0Mjk1ODUxYzAyNDBmY2Y5YyJ9',
                                oid: '54605d8bd456670e0bd9d2b4295851c0240fcf9c',
                                abbreviatedOID: '54605d8',
                                message:
                                    'Cody: Release 0.0.9 (#51206)\n\nPreparing for a new release by pushing the version number.\r\n\r\n## Test plan\r\n\r\nOnly a version number change.\r\n\r\n\u003C!-- All pull requests REQUIRE a test plan:\r\nhttps://docs.sourcegraph.com/dev/background-information/testing_principles\r\n--\u003E',
                                subject: 'Cody: Release 0.0.9 (#51206)',
                                body: 'Preparing for a new release by pushing the version number.\r\n\r\n## Test plan\r\n\r\nOnly a version number change.\r\n\r\n\u003C!-- All pull requests REQUIRE a test plan:\r\nhttps://docs.sourcegraph.com/dev/background-information/testing_principles\r\n--\u003E',
                                author: {
                                    person: {
                                        avatarURL: null,
                                        name: 'Philipp Spiess',
                                        email: 'hello@philippspiess.com',
                                        displayName: 'Philipp Spiess',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-27T15:46:36Z',
                                    __typename: 'Signature',
                                },
                                committer: {
                                    person: {
                                        avatarURL: null,
                                        name: 'GitHub',
                                        email: 'noreply@github.com',
                                        displayName: 'GitHub',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-27T15:46:36Z',
                                    __typename: 'Signature',
                                },
                                parents: [
                                    {
                                        oid: '84dd723a59ee5e13b9bfed8357975357b36eb318',
                                        abbreviatedOID: '84dd723',
                                        url: '/github.com/sourcegraph/sourcegraph/-/commit/84dd723a59ee5e13b9bfed8357975357b36eb318',
                                        __typename: 'GitCommit',
                                    },
                                ],
                                url: '/github.com/sourcegraph/sourcegraph/-/commit/54605d8bd456670e0bd9d2b4295851c0240fcf9c',
                                canonicalURL:
                                    '/github.com/sourcegraph/sourcegraph/-/commit/54605d8bd456670e0bd9d2b4295851c0240fcf9c',
                                externalURLs: [
                                    {
                                        url: 'https://github.com/sourcegraph/sourcegraph/commit/54605d8bd456670e0bd9d2b4295851c0240fcf9c',
                                        serviceKind: 'GITHUB',
                                        __typename: 'ExternalLink',
                                    },
                                ],
                                tree: {
                                    canonicalURL:
                                        '/github.com/sourcegraph/sourcegraph@54605d8bd456670e0bd9d2b4295851c0240fcf9c',
                                    __typename: 'GitTree',
                                },
                                __typename: 'GitCommit',
                            },
                        ],
                        pageInfo: { hasNextPage: true, endCursor: '20', __typename: 'PageInfo' },
                        __typename: 'GitCommitConnection',
                    },
                    __typename: 'GitCommit',
                },
                __typename: 'Repository',
            },
        },
    },
}

const mockRepositoryPerforceChangelistsQuery: MockedResponse<
    RepositoryGitCommitsResult,
    RepositoryGitCommitsVariables
> = {
    request: {
        query: getDocumentNode(REPOSITORY_GIT_COMMITS_QUERY),
        variables: {
            repo: 'UmVwb3NpdG9yeToyNjM4OQ==',
            revspec: '',
            filePath: '',
            first: 20,
            afterCursor: null,
        },
    },
    result: {
        data: {
            node: {
                isPerforceDepot: true,
                externalURLs: [],
                commit: {
                    ancestors: {
                        nodes: [
                            {
                                id: 'R2l0Q29tbWl0OnsiciI6IlVtVndiM05wZEc5eWVUb3lOak00T1E9PSIsImMiOiJhNzdkMWIzMDliY2U1ZGJiNjM1ODFiY2E4YzJjYWFjMDEzZWM5Mzg3In0=',
                                oid: '48485',
                                abbreviatedOID: '48485',
                                message: '48485 - test-5386\n[p4-fusion: depot-paths = "//go/": change = 48485]',
                                subject: 'test-5386',
                                body: null,
                                author: {
                                    person: {
                                        avatarURL: null,
                                        name: 'admin',
                                        email: 'admin@perforce-server-7df6ff678c-lkzfb',
                                        displayName: 'admin',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2021-08-13T19:24:59Z',
                                    __typename: 'Signature',
                                },
                                committer: {
                                    person: {
                                        avatarURL: null,
                                        name: 'admin',
                                        email: 'admin@perforce-server-7df6ff678c-lkzfb',
                                        displayName: 'admin',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2021-08-13T19:24:59Z',
                                    __typename: 'Signature',
                                },
                                parents: [
                                    {
                                        oid: 'd9994dc548fd79b473ce05198c88282890983fa9',
                                        abbreviatedOID: 'd9994dc',
                                        url: '/perforce.sgdev.org/go/-/commit/d9994dc548fd79b473ce05198c88282890983fa9',
                                        __typename: 'GitCommit',
                                    },
                                ],
                                url: '/perforce.sgdev.org/go/-/commit/a77d1b309bce5dbb63581bca8c2caac013ec9387',
                                canonicalURL:
                                    '/perforce.sgdev.org/go/-/commit/a77d1b309bce5dbb63581bca8c2caac013ec9387',
                                externalURLs: [],
                                tree: {
                                    canonicalURL: '/perforce.sgdev.org/go@a77d1b309bce5dbb63581bca8c2caac013ec9387',
                                    __typename: 'GitTree',
                                },
                                __typename: 'GitCommit',
                            },
                            {
                                id: 'R2l0Q29tbWl0OnsiciI6IlVtVndiM05wZEc5eWVUb3lOak00T1E9PSIsImMiOiJkOTk5NGRjNTQ4ZmQ3OWI0NzNjZTA1MTk4Yzg4MjgyODkwOTgzZmE5In0=',
                                oid: '1012',
                                abbreviatedOID: '1012',
                                message: '1012 - :boar:\n\n[p4-fusion: depot-paths = "//go/": change = 1012]',
                                subject: ':boar:',
                                body: null,
                                author: {
                                    person: {
                                        avatarURL: null,
                                        name: 'admin',
                                        email: 'admin@perforce-server-7df6ff678c-lkzfb',
                                        displayName: 'admin',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2021-08-09T22:01:07Z',
                                    __typename: 'Signature',
                                },
                                committer: {
                                    person: {
                                        avatarURL: null,
                                        name: 'admin',
                                        email: 'admin@perforce-server-7df6ff678c-lkzfb',
                                        displayName: 'admin',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2021-08-09T22:01:07Z',
                                    __typename: 'Signature',
                                },
                                parents: [
                                    {
                                        oid: 'd7205ebf51c2c7486ee74ded9f000ddf1b0cca24',
                                        abbreviatedOID: 'd7205eb',
                                        url: '/perforce.sgdev.org/go/-/commit/d7205ebf51c2c7486ee74ded9f000ddf1b0cca24',
                                        __typename: 'GitCommit',
                                    },
                                ],
                                url: '/perforce.sgdev.org/go/-/commit/d9994dc548fd79b473ce05198c88282890983fa9',
                                canonicalURL:
                                    '/perforce.sgdev.org/go/-/commit/d9994dc548fd79b473ce05198c88282890983fa9',
                                externalURLs: [],
                                tree: {
                                    canonicalURL: '/perforce.sgdev.org/go@d9994dc548fd79b473ce05198c88282890983fa9',
                                    __typename: 'GitTree',
                                },
                                __typename: 'GitCommit',
                            },
                            {
                                id: 'R2l0Q29tbWl0OnsiciI6IlVtVndiM05wZEc5eWVUb3lOak00T1E9PSIsImMiOiJkNzIwNWViZjUxYzJjNzQ4NmVlNzRkZWQ5ZjAwMGRkZjFiMGNjYTI0In0=',
                                oid: '1011',
                                abbreviatedOID: '1011',
                                message:
                                    '1011 - Add Go source code\n\n[p4-fusion: depot-paths = "//go/": change = 1011]',
                                subject: 'Add Go source code',
                                body: null,
                                author: {
                                    person: {
                                        avatarURL: null,
                                        name: 'admin',
                                        email: 'admin@perforce-server-7df6ff678c-lkzfb',
                                        displayName: 'admin',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2021-08-09T21:44:50Z',
                                    __typename: 'Signature',
                                },
                                committer: {
                                    person: {
                                        avatarURL: null,
                                        name: 'admin',
                                        email: 'admin@perforce-server-7df6ff678c-lkzfb',
                                        displayName: 'admin',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2021-08-09T21:44:50Z',
                                    __typename: 'Signature',
                                },
                                parents: [],
                                url: '/perforce.sgdev.org/go/-/commit/d7205ebf51c2c7486ee74ded9f000ddf1b0cca24',
                                canonicalURL:
                                    '/perforce.sgdev.org/go/-/commit/d7205ebf51c2c7486ee74ded9f000ddf1b0cca24',
                                externalURLs: [],
                                tree: {
                                    canonicalURL: '/perforce.sgdev.org/go@d7205ebf51c2c7486ee74ded9f000ddf1b0cca24',
                                    __typename: 'GitTree',
                                },
                                __typename: 'GitCommit',
                            },
                        ],
                        pageInfo: { hasNextPage: false, endCursor: null, __typename: 'PageInfo' },
                        __typename: 'GitCommitConnection',
                    },
                    __typename: 'GitCommit',
                },
                __typename: 'Repository',
            },
        },
    },
}

const repo: RepositoryFields = {
    id: 'repo-id',
    name: 'github.com/sourcegraph/sourcegraph',
    url: 'https://github.com/sourcegraph/sourcegraph/perforce',
    isPerforceDepot: false,
    description: '',
    viewerCanAdminister: false,
    isFork: false,
    externalURLs: [],
    externalRepository: {
        __typename: 'ExternalRepository',
        serviceType: '',
        serviceID: '',
    },
    defaultBranch: null,
    metadata: [],
}

export const GitCommitsStory: Story<RepositoryCommitsPageProps> = () => (
    <MockedTestProvider>
        <WebStory
            initialEntries={['/github.com/sourcegraph/sourcegraph/-/commits']}
            mocks={[mockRepositoryGitCommitsQuery]}
        >
            {props => <RepositoryCommitsPage revision="" repo={repo} {...props} />}
        </WebStory>
    </MockedTestProvider>
)

GitCommitsStory.storyName = 'Git commits'

export const PerforceChangelistsStory: Story<RepositoryCommitsPageProps> = () => (
    <MockedTestProvider>
        <WebStory
            initialEntries={['/perforce.sgdev.org/go/-/commits']}
            mocks={[mockRepositoryPerforceChangelistsQuery]}
        >
            {props => <RepositoryCommitsPage revision="" repo={repo} {...props} />}
        </WebStory>
    </MockedTestProvider>
)

PerforceChangelistsStory.storyName = 'Perforce changelists'
