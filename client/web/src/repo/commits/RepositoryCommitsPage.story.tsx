import type { MockedResponse } from '@apollo/client/testing'
import type { Decorator, Meta, StoryFn } from '@storybook/react'

import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../components/WebStory'
import {
    ExternalServiceKind,
    type RepositoryFields,
    type RepositoryGitCommitsResult,
    type RepositoryGitCommitsVariables,
    RepositoryType,
} from '../../graphql-operations'

import {
    REPOSITORY_GIT_COMMITS_QUERY,
    RepositoryCommitsPage,
    type RepositoryCommitsPageProps,
} from './RepositoryCommitsPage'

window.context.experimentalFeatures = { perforceChangelistMapping: 'enabled' }

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

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
                sourceType: RepositoryType.GIT_REPOSITORY,
                externalURLs: [
                    {
                        url: 'https://github.com/sourcegraph/sourcegraph',
                        serviceKind: ExternalServiceKind.GITHUB,
                        __typename: 'ExternalLink',
                    },
                ],
                commit: {
                    ancestors: {
                        nodes: [
                            {
                                id: 'R2l0Q29tbWl0OnsiciI6IlVtVndiM05wZEc5eWVUbzMiLCJjIjoiYjFlMWVjMTZlMmMyZDkzMmNmMjMwYzcxYzc2MDUxNmZiOTY1OGNkYiJ9',
                                oid: 'b1e1ec16e2c2d932cf230c71c760516fb9658cdb',
                                abbreviatedOID: 'b1e1ec1',
                                perforceChangelist: null,
                                message:
                                    'Document OpenAI API compatibility for completions (#51208)\n\nCo-authored-by: Malo Marrec <malo.marrec@gmail.com>',
                                subject: 'Document OpenAI API compatibility for completions (#51208)',
                                body: 'Co-authored-by: Malo Marrec <malo.marrec@gmail.com>',
                                author: {
                                    person: {
                                        avatarURL: null,
                                        name: 'David Veszelovszki',
                                        email: 'veszelovszki@gmail.com',
                                        displayName: 'David Veszelovszki',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-05-02T11:00:43Z',
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
                                    date: '2023-05-02T11:00:43Z',
                                    __typename: 'Signature',
                                },
                                parents: [
                                    {
                                        oid: '59ce236cca392975a383cdab14d73244f05d21b6',
                                        abbreviatedOID: '59ce236',
                                        perforceChangelist: null,
                                        url: '/github.com/sourcegraph/sourcegraph/-/commit/59ce236cca392975a383cdab14d73244f05d21b6',
                                        __typename: 'GitCommit',
                                    },
                                ],
                                url: '/github.com/sourcegraph/sourcegraph/-/commit/b1e1ec16e2c2d932cf230c71c760516fb9658cdb',
                                canonicalURL:
                                    '/github.com/sourcegraph/sourcegraph/-/commit/b1e1ec16e2c2d932cf230c71c760516fb9658cdb',
                                externalURLs: [
                                    {
                                        url: 'https://github.com/sourcegraph/sourcegraph/commit/b1e1ec16e2c2d932cf230c71c760516fb9658cdb',
                                        serviceKind: ExternalServiceKind.GITHUB,
                                        __typename: 'ExternalLink',
                                    },
                                ],
                                tree: {
                                    canonicalURL:
                                        '/github.com/sourcegraph/sourcegraph@b1e1ec16e2c2d932cf230c71c760516fb9658cdb',
                                    __typename: 'GitTree',
                                },
                                __typename: 'GitCommit',
                            },
                        ],
                        pageInfo: {
                            hasNextPage: true,
                            endCursor: '20',
                            __typename: 'PageInfo',
                        },
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
                sourceType: RepositoryType.PERFORCE_DEPOT,
                externalURLs: [],
                commit: {
                    ancestors: {
                        nodes: [
                            {
                                id: 'R2l0Q29tbWl0OnsiciI6IlVtVndiM05wZEc5eWVUb3lOak00T1E9PSIsImMiOiJhNzdkMWIzMDliY2U1ZGJiNjM1ODFiY2E4YzJjYWFjMDEzZWM5Mzg3In0=',
                                oid: 'f27c0f663b36e5294947964d7c25672f6e0e34fe',
                                abbreviatedOID: 'f27c0f663b',
                                perforceChangelist: {
                                    __typename: 'PerforceChangelist',
                                    cid: '48485',
                                    canonicalURL: '/go/-/changelist/48485',
                                },
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
                                        perforceChangelist: {
                                            __typename: 'PerforceChangelist',
                                            cid: '48484',
                                            canonicalURL: '/go/-/changelist/48484',
                                        },
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
                                oid: '93abc40a6193539d27022a4ce558fcf7d82d3998',
                                abbreviatedOID: '93abc40a61',
                                perforceChangelist: {
                                    __typename: 'PerforceChangelist',
                                    cid: '1012',
                                    canonicalURL: '/go/-/changelist/1012',
                                },
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
                                        perforceChangelist: {
                                            __typename: 'PerforceChangelist',
                                            cid: '48485',
                                            canonicalURL: '/go/-/changelist/48485',
                                        },
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
                                oid: '314a3c2830b6aa500dfbb23ac05aa2304c58f1e1',
                                abbreviatedOID: '314a3c2830',
                                perforceChangelist: {
                                    __typename: 'PerforceChangelist',
                                    cid: '1011',
                                    canonicalURL: '/go/-/changelist/1011',
                                },
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

const gitRepo: RepositoryFields = {
    id: 'repo-id',
    name: 'github.com/sourcegraph/sourcegraph',
    url: 'https://github.com/sourcegraph/sourcegraph',
    sourceType: RepositoryType.GIT_REPOSITORY,
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
    topics: [],
}

export const GitCommitsStory: StoryFn<RepositoryCommitsPageProps> = () => (
    <MockedTestProvider>
        <WebStory
            initialEntries={['/github.com/sourcegraph/sourcegraph/-/commits']}
            mocks={[mockRepositoryGitCommitsQuery]}
        >
            {props => <RepositoryCommitsPage revision="" repo={gitRepo} {...props} />}
        </WebStory>
    </MockedTestProvider>
)

GitCommitsStory.storyName = 'Git commits'

export const GitCommitsInPathStory: StoryFn<RepositoryCommitsPageProps> = () => (
    <MockedTestProvider>
        <WebStory
            initialEntries={['/github.com/sourcegraph/sourcegraph/-/commits/somePath']}
            mocks={[mockRepositoryGitCommitsQuery]}
        >
            {props => <RepositoryCommitsPage revision="" repo={gitRepo} {...props} />}
        </WebStory>
    </MockedTestProvider>
)

GitCommitsInPathStory.storyName = 'Git commits in a path'

const perforceRepo: RepositoryFields = {
    id: 'repo-id',
    name: 'github.com/sourcegraph/sourcegraph',
    url: 'https://github.com/sourcegraph/sourcegraph',
    sourceType: RepositoryType.PERFORCE_DEPOT,
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
    topics: [],
}

export const PerforceChangelistsStory: StoryFn<RepositoryCommitsPageProps> = () => (
    <MockedTestProvider>
        <WebStory
            initialEntries={['/perforce.sgdev.org/go/-/changelists']}
            mocks={[mockRepositoryPerforceChangelistsQuery]}
        >
            {props => <RepositoryCommitsPage revision="" repo={perforceRepo} {...props} />}
        </WebStory>
    </MockedTestProvider>
)

PerforceChangelistsStory.storyName = 'Perforce changelists'

export const PerforceChangelistsInPathStory: StoryFn<RepositoryCommitsPageProps> = () => (
    <MockedTestProvider>
        <WebStory
            initialEntries={['/perforce.sgdev.org/go/-/changelists/somePath']}
            mocks={[mockRepositoryPerforceChangelistsQuery]}
        >
            {props => <RepositoryCommitsPage revision="" repo={perforceRepo} {...props} />}
        </WebStory>
    </MockedTestProvider>
)

PerforceChangelistsInPathStory.storyName = 'Perforce changelists in a path'
