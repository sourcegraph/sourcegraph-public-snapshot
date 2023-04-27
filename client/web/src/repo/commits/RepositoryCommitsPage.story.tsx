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

export const RepositoryCommitsPageStory: Story<RepositoryCommitsPageProps> = () => (
    <MockedTestProvider>
        <WebStory initialEntries={['/perforce.sgdev.org/go/-/commits']} mocks={[mockRepositoryGitCommitsQuery]}>
            {props => <RepositoryCommitsPage revision="" repo={repo} {...props} />}
        </WebStory>
    </MockedTestProvider>
)

RepositoryCommitsPageStory.storyName = 'Repository commits page'
