import { MockedResponse } from '@apollo/client/testing'

import { getDocumentNode } from '@sourcegraph/shared/src/graphql/graphql'

import { RepositoriesForPopoverResult, RepositoryPopoverFields } from '../../graphql-operations'

import { REPOSITORIES_FOR_POPOVER } from './RepositoriesPopover'

const generateRepositoryNodes = (nodeCount: number): RepositoryPopoverFields[] =>
    new Array(nodeCount).fill(null).map((_value, index) => ({
        __typename: 'Repository',
        id: `repository-${index}`,
        name: `github.com/some-org/repository-name-${index}`,
    }))

const repositoriesMock: MockedResponse<RepositoriesForPopoverResult> = {
    request: {
        query: getDocumentNode(REPOSITORIES_FOR_POPOVER),
        variables: {
            query: '',
            first: 10,
        },
    },
    result: {
        data: {
            repositories: {
                nodes: generateRepositoryNodes(10),
                totalCount: null,
                pageInfo: {
                    hasNextPage: true,
                },
            },
        },
    },
}

const additionalRepositoriesMock: MockedResponse<RepositoriesForPopoverResult> = {
    request: {
        query: getDocumentNode(REPOSITORIES_FOR_POPOVER),
        variables: {
            query: '',
            first: 20,
        },
    },
    result: {
        data: {
            repositories: {
                nodes: generateRepositoryNodes(20),
                totalCount: null,
                pageInfo: {
                    hasNextPage: false,
                },
            },
        },
    },
}

const filteredRepositoriesMock: MockedResponse<RepositoriesForPopoverResult> = {
    request: {
        query: getDocumentNode(REPOSITORIES_FOR_POPOVER),
        variables: {
            query: 'some query',
            first: 10,
        },
    },
    result: {
        data: {
            repositories: {
                nodes: generateRepositoryNodes(2),
                totalCount: null,
                pageInfo: {
                    hasNextPage: false,
                },
            },
        },
    },
}

const additionalFilteredRepositoriesMock: MockedResponse<RepositoriesForPopoverResult> = {
    request: {
        query: getDocumentNode(REPOSITORIES_FOR_POPOVER),
        variables: {
            query: 'some other query',
            first: 10,
        },
    },
    result: {
        data: {
            repositories: {
                nodes: [],
                totalCount: null,
                pageInfo: {
                    hasNextPage: false,
                },
            },
        },
    },
}

export const MOCK_REQUESTS = [
    repositoriesMock,
    additionalRepositoriesMock,
    filteredRepositoriesMock,
    additionalFilteredRepositoriesMock,
]
