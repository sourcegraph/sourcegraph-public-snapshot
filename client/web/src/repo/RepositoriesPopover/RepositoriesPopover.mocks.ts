import type { MockedResponse } from '@apollo/client/testing'

import { getDocumentNode } from '@sourcegraph/http-client'

import type { RepositoriesForPopoverResult, RepositoryPopoverFields } from '../../graphql-operations'

import { REPOSITORIES_FOR_POPOVER, BATCH_COUNT } from './RepositoriesPopover'

interface GenerateRepositoryNodesParameters {
    count: number
    offset?: number
}

const generateRepositoryNodes = ({ count, offset = 0 }: GenerateRepositoryNodesParameters): RepositoryPopoverFields[] =>
    new Array(count).fill(null).map((_value, index) => {
        const increment = index + offset
        return {
            __typename: 'Repository',
            id: `repository-${increment}`,
            name: `github.com/some-org/repository-name-${increment}`,
        }
    })

const MOCK_CURSOR = '12345'

const repositoriesMock: MockedResponse<RepositoriesForPopoverResult> = {
    request: {
        query: getDocumentNode(REPOSITORIES_FOR_POPOVER),
        variables: {
            query: '',
            first: BATCH_COUNT,
            after: null,
        },
    },
    result: {
        data: {
            repositories: {
                nodes: generateRepositoryNodes({ count: 10 }),
                pageInfo: {
                    hasNextPage: true,
                    endCursor: MOCK_CURSOR,
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
            first: BATCH_COUNT,
            after: MOCK_CURSOR,
        },
    },
    result: {
        data: {
            repositories: {
                nodes: generateRepositoryNodes({ count: 10, offset: 10 }),
                pageInfo: {
                    hasNextPage: false,
                    endCursor: null,
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
            first: BATCH_COUNT,
            after: null,
        },
    },
    result: {
        data: {
            repositories: {
                nodes: generateRepositoryNodes({ count: 2, offset: 10 }),
                pageInfo: {
                    hasNextPage: false,
                    endCursor: null,
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
            first: BATCH_COUNT,
            after: null,
        },
    },
    result: {
        data: {
            repositories: {
                nodes: [],
                pageInfo: {
                    hasNextPage: false,
                    endCursor: null,
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
