import type { MockedResponse } from '@apollo/client/testing'
import { subDays } from 'date-fns'

import { getDocumentNode } from '@sourcegraph/http-client'
import { GitRefType } from '@sourcegraph/shared/src/graphql-operations'

import type {
    GitCommitAncestorsConnectionFields,
    GitRefConnectionFields,
    RepositoryGitCommitResult,
    RepositoryGitRefsResult,
} from '../../graphql-operations'
import { REPOSITORY_GIT_REFS } from '../GitReference'

import type { RevisionsPopoverProps } from './RevisionsPopover'
import { REPOSITORY_GIT_COMMIT } from './RevisionsPopoverCommits'

export const MOCK_PROPS: RevisionsPopoverProps = {
    repoId: 'some-repo-id',
    repoName: 'testorg/testrepo',
    repoServiceType: 'github',
    defaultBranch: 'main',
    currentRev: 'main',
    togglePopover: () => null,
    showSpeculativeResults: false,
}

const yesterday = subDays(new Date(), 1).toISOString()

const commitPerson = {
    displayName: 'display-name',
    user: {
        username: 'username',
    },
}

const generateGitReferenceNodes = (nodeCount: number, variant: GitRefType): GitRefConnectionFields['nodes'] =>
    new Array(nodeCount).fill(null).map((_value, index) => {
        const id = `${variant}-${index}`
        return {
            __typename: 'GitRef',
            id,
            displayName: `${id}-display-name`,
            abbrevName: `${id}-abbrev-name`,
            name: `refs/heads/${id}`,
            url: `/github.com/testorg/testrepo@${id}-display-name`,
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
                    behindAhead: {},
                },
            },
        }
    }) as GitRefConnectionFields['nodes']

const generateGitCommitNodes = (nodeCount: number): GitCommitAncestorsConnectionFields['nodes'] =>
    new Array(nodeCount).fill(null).map((_value, index) => ({
        __typename: 'GitCommit',
        id: `git-commit-${index}`,
        oid: `git-commit-oid-${index}`,
        abbreviatedOID: `git-commit-oid-${index}`,
        author: {
            person: {
                name: commitPerson.displayName,
                avatarURL: null,
            },
            date: yesterday,
        },
        subject: `Commit ${index}: Hello world`,
    }))

const branchesMock: MockedResponse<RepositoryGitRefsResult> = {
    request: {
        query: getDocumentNode(REPOSITORY_GIT_REFS),
        variables: {
            query: '',
            first: 50,
            repo: MOCK_PROPS.repoId,
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
                    nodes: generateGitReferenceNodes(50, GitRefType.GIT_BRANCH),
                    pageInfo: {
                        hasNextPage: true,
                    },
                },
            },
        },
    },
}

const additionalBranchesMock: MockedResponse<RepositoryGitRefsResult> = {
    request: {
        ...branchesMock.request,
        variables: {
            ...branchesMock.request.variables,
            first: 100,
        },
    },
    result: {
        data: {
            node: {
                __typename: 'Repository',
                gitRefs: {
                    __typename: 'GitRefConnection',
                    totalCount: 100,
                    nodes: generateGitReferenceNodes(100, GitRefType.GIT_BRANCH),
                    pageInfo: {
                        hasNextPage: false,
                    },
                },
            },
        },
    },
}

const filteredBranchesMock: MockedResponse<RepositoryGitRefsResult> = {
    request: {
        ...branchesMock.request,
        variables: {
            ...branchesMock.request.variables,
            query: 'some query',
        },
    },
    result: {
        data: {
            node: {
                __typename: 'Repository',
                gitRefs: {
                    __typename: 'GitRefConnection',
                    totalCount: 2,
                    nodes: generateGitReferenceNodes(2, GitRefType.GIT_BRANCH),
                    pageInfo: {
                        hasNextPage: false,
                    },
                },
            },
        },
    },
}

const filteredBranchesNoResultsMock: MockedResponse<RepositoryGitRefsResult> = {
    request: {
        ...branchesMock.request,
        variables: {
            ...branchesMock.request.variables,
            query: 'some other query',
        },
    },
    result: {
        data: {
            node: {
                __typename: 'Repository',
                gitRefs: {
                    __typename: 'GitRefConnection',
                    totalCount: 0,
                    nodes: [],
                    pageInfo: {
                        hasNextPage: false,
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
            repo: MOCK_PROPS.repoId,
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
                    nodes: generateGitReferenceNodes(50, GitRefType.GIT_TAG),
                    pageInfo: {
                        hasNextPage: true,
                    },
                },
            },
        },
    },
}

const additionalTagsMock: MockedResponse<RepositoryGitRefsResult> = {
    request: {
        ...tagsMock.request,
        variables: {
            ...tagsMock.request.variables,
            first: 100,
        },
    },
    result: {
        data: {
            node: {
                __typename: 'Repository',
                gitRefs: {
                    __typename: 'GitRefConnection',
                    totalCount: 100,
                    nodes: generateGitReferenceNodes(100, GitRefType.GIT_TAG),
                    pageInfo: {
                        hasNextPage: false,
                    },
                },
            },
        },
    },
}

const filteredTagsMock: MockedResponse<RepositoryGitRefsResult> = {
    request: {
        ...tagsMock.request,
        variables: {
            ...tagsMock.request.variables,
            query: 'some query',
        },
    },
    result: {
        data: {
            node: {
                __typename: 'Repository',
                gitRefs: {
                    __typename: 'GitRefConnection',
                    totalCount: 2,
                    nodes: generateGitReferenceNodes(2, GitRefType.GIT_TAG),
                    pageInfo: {
                        hasNextPage: false,
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
            repo: MOCK_PROPS.repoId,
            revision: MOCK_PROPS.currentRev,
        },
    },
    result: {
        data: {
            node: {
                __typename: 'Repository',
                commit: {
                    __typename: 'GitCommit',
                    ancestors: {
                        __typename: 'GitCommitConnection',
                        nodes: generateGitCommitNodes(15),
                        pageInfo: {
                            hasNextPage: true,
                        },
                    },
                },
            },
        },
    },
}

const additionalCommitsMock: MockedResponse<RepositoryGitCommitResult> = {
    request: {
        ...commitsMock.request,
        variables: {
            ...commitsMock.request.variables,
            first: 30,
        },
    },
    result: {
        data: {
            node: {
                __typename: 'Repository',
                commit: {
                    __typename: 'GitCommit',
                    ancestors: {
                        __typename: 'GitCommitConnection',
                        nodes: generateGitCommitNodes(30),
                        pageInfo: {
                            hasNextPage: false,
                        },
                    },
                },
            },
        },
    },
}

const filteredCommitsMock: MockedResponse<RepositoryGitCommitResult> = {
    request: {
        ...commitsMock.request,
        variables: {
            ...commitsMock.request.variables,
            query: 'some query',
        },
    },
    result: {
        data: {
            node: {
                __typename: 'Repository',
                commit: {
                    __typename: 'GitCommit',
                    ancestors: {
                        __typename: 'GitCommitConnection',
                        nodes: generateGitCommitNodes(2),
                        pageInfo: {
                            hasNextPage: false,
                        },
                    },
                },
            },
        },
    },
}

/**
 * This mock is to test the case where a speculative revision is provided as context.
 * In this case, the code should not error as it is still valid to display 0 results.
 */
const noCommitsMock: MockedResponse<RepositoryGitCommitResult> = {
    request: {
        ...commitsMock.request,
        variables: {
            ...commitsMock.request.variables,
            revision: 'non-existent-revision',
        },
    },
    result: {
        data: {
            node: {
                __typename: 'Repository',
                commit: null,
            },
        },
    },
}

export const MOCK_REQUESTS = [
    branchesMock,
    additionalBranchesMock,
    filteredBranchesMock,
    filteredBranchesNoResultsMock,
    tagsMock,
    additionalTagsMock,
    filteredTagsMock,
    commitsMock,
    additionalCommitsMock,
    filteredCommitsMock,
    noCommitsMock,
]
