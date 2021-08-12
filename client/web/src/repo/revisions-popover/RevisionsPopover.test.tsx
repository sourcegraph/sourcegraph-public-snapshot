import { MockedProvider, MockedResponse } from '@apollo/client/testing'
import { RenderResult, cleanup } from '@testing-library/react'
import { subDays } from 'date-fns'
import React from 'react'

import { GitRefType } from '@sourcegraph/shared/src/graphql-operations'
import { getDocumentNode } from '@sourcegraph/shared/src/graphql/graphql'
import { waitForNextApolloResponse } from '@sourcegraph/shared/src/testing/apollo'
import { renderWithRouter } from '@sourcegraph/shared/src/testing/render-with-router'

import { RepositoryGitCommitResult, RepositoryGitRefsResult } from '../../graphql-operations'
import { REPOSITORY_GIT_REFS } from '../GitReference'

import { RevisionsPopover, RevisionsPopoverProps } from './RevisionsPopover'
import { REPOSITORY_GIT_COMMIT } from './RevisionsPopoverCommits'

describe('RevisionsPopover', () => {
    let queries: RenderResult

    // eslint-disable-next-line ban/ban
    const togglePopoverMock = jest.fn()

    beforeEach(() => {
        togglePopoverMock.mockReset()
    })

    const props: RevisionsPopoverProps = {
        repo: 'some-repo-id',
        repoName: 'testorg/testrepo',
        defaultBranch: 'main',
        currentRev: undefined,
        currentCommitID: undefined,
        togglePopover: togglePopoverMock,
        getURLFromRevision: undefined,
        allowSpeculativeSearch: false,
    }

    const yesterday = subDays(new Date(), 1).toISOString()

    const commitPerson = {
        displayName: 'display-name',
        user: {
            username: 'username',
        },
    }
    const branchMock: MockedResponse<RepositoryGitRefsResult> = {
        request: {
            query: getDocumentNode(REPOSITORY_GIT_REFS),
            variables: {
                query: '',
                first: 50,
                repo: props.repo,
                type: GitRefType.GIT_BRANCH,
                withBehindAhead: false,
            },
        },
        result: {
            data: {
                node: {
                    __typename: 'Repository',
                    gitRefs: {
                        __typename: 'GitRefConnection',
                        totalCount: 100,
                        nodes: [
                            {
                                __typename: 'GitRef',
                                id: 'id-1',
                                displayName: 'branch-display-name',
                                abbrevName: 'branch-display-name',
                                name: 'refs/heads/branch-display-name',
                                url: '/github.com/testorg/testrepo@branch-display-name',
                                target: {
                                    commit: {
                                        author: {
                                            __typename: 'Signature',
                                            date: yesterday,
                                            person: commitPerson,
                                        },
                                        committer: {
                                            __typename: 'Signature',
                                            date: yesterday,
                                            person: commitPerson,
                                        },
                                        behindAhead: null,
                                    },
                                },
                            },
                        ],
                        pageInfo: {
                            hasNextPage: true,
                        },
                    },
                },
            },
        },
    }

    const tagsMock: MockedResponse<RepositoryGitRefsResult> = {
        request: {
            query: getDocumentNode(REPOSITORY_GIT_REFS),
            variables: {
                query: '',
                first: 50,
                repo: props.repo,
                type: GitRefType.GIT_TAG,
                withBehindAhead: false,
            },
        },
        result: {
            data: {
                node: {
                    __typename: 'Repository',
                    gitRefs: {
                        __typename: 'GitRefConnection',
                        totalCount: 100,
                        nodes: [
                            {
                                __typename: 'GitRef',
                                id: 'id-1',
                                displayName: 'branch-display-name',
                                abbrevName: 'branch-display-name',
                                name: 'refs/heads/branch-display-name',
                                url: '/github.com/testorg/testrepo@branch-display-name',
                                target: {
                                    commit: {
                                        author: {
                                            __typename: 'Signature',
                                            date: yesterday,
                                            person: commitPerson,
                                        },
                                        committer: {
                                            __typename: 'Signature',
                                            date: yesterday,
                                            person: commitPerson,
                                        },
                                        behindAhead: null,
                                    },
                                },
                            },
                        ],
                        pageInfo: {
                            hasNextPage: true,
                        },
                    },
                },
            },
        },
    }

    const commitsMock: MockedResponse<RepositoryGitCommitResult> = {
        request: {
            query: getDocumentNode(REPOSITORY_GIT_COMMIT),
            variables: {
                query: '',
                first: 15,
                repo: props.repo,
                revision: props.defaultBranch,
            },
        },
        result: {
            data: {
                node: {
                    __typename: 'Repository',
                    commit: {
                        ancestors: {
                            nodes: [
                                {
                                    id: 'some-id',
                                    oid: 'some-oid-12345',
                                    abbreviatedOID: 'some-oid',
                                    author: {
                                        person: {
                                            name: commitPerson.displayName,
                                            avatarURL: null,
                                        },
                                        date: yesterday,
                                    },
                                    subject: 'Commit: do something',
                                },
                            ],
                            pageInfo: {
                                hasNextPage: true,
                            },
                        },
                    },
                },
            },
        },
    }

    describe('Branches', () => {
        beforeEach(async () => {
            queries = renderWithRouter(
                <MockedProvider mocks={[branchMock, tagsMock, commitsMock]} addTypename={false}>
                    <RevisionsPopover {...props} />
                </MockedProvider>,
                { route: `/${props.repoName}` }
            )

            await waitForNextApolloResponse()
        })

        it('renders results correctly', () => {
            queries.debug()
        })

        it.skip('fetches additional results correctly', () => {})

        it.skip('filters correctly', () => {})
    })

    describe.skip('Tags', () => {})

    describe.skip('Commits', () => {})

    afterEach(cleanup)
})

/**
 *
 * 1. Renders results correctly
 * 2. Filters correctly
 * 3. Fetches additional results correctly
 *
 * Additional:
 * 1. URL logic
 * 2. Speculative search
 *
 */
