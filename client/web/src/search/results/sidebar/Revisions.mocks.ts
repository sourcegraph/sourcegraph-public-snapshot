import type { MockedResponse } from '@apollo/client/testing'
import { GraphQLError } from 'graphql'

import type { RevisionsProps } from '@sourcegraph/branded'
import { getDocumentNode } from '@sourcegraph/http-client'

import { GitRefType } from '../../../../../shared/src/graphql-operations'
import type {
    SearchSidebarGitRefsResult,
    // SearchSidebarGitRefsVariables,
    SearchSidebarGitRefFields,
} from '../../../graphql-operations'

import { GIT_REVS_QUERY } from './Revisions'

export const MOCK_PROPS: RevisionsProps = {
    query: '',
    repoName: 'testorg/testrepo',
    onFilterClick: () => {},
}

export const FILTERED_MOCK_PROPS: RevisionsProps = {
    query: 'test',
    repoName: 'testorg/testrepo',
    onFilterClick: () => {},
}

function generateMockedRequest(
    type: GitRefType,
    first = 10,
    query = ''
): MockedResponse<SearchSidebarGitRefsResult>['request'] {
    return {
        query: getDocumentNode(GIT_REVS_QUERY),
        variables: {
            repo: MOCK_PROPS.repoName,
            first,
            query,
            type,
        },
    }
}

export function generateMockedResponses(
    type: GitRefType,
    totalCount: number,
    query = ''
): MockedResponse<SearchSidebarGitRefsResult>[] {
    const nodes: SearchSidebarGitRefFields[] = Array.from({ length: totalCount }, (_value, index) => {
        const id = `${type}-${index}`
        return {
            id,
            __typename: 'GitRef',
            name: `refs/heads/${id}-name`,
            displayName: `${id}-display-name`,
        }
    })
    return Array.from({ length: Math.max(1, Math.ceil(totalCount / 10)) }, (_value, index) => {
        const first = (index + 1) * 10
        return {
            request: generateMockedRequest(type, first, query),
            result: {
                data: {
                    repository: {
                        __typename: 'Repository',
                        id: 'repo',
                        gitRefs: {
                            __typename: 'GitRefConnection',
                            nodes: nodes.slice(0, Math.min(first, totalCount)),
                            totalCount,
                            pageInfo: {
                                hasNextPage: first < totalCount,
                            },
                        },
                    },
                },
            },
        }
    })
}

// For empty state tests
const emptyBranchesMocks = generateMockedResponses(GitRefType.GIT_BRANCH, 0)
const emptyTagsMocks = generateMockedResponses(GitRefType.GIT_TAG, 0)

// For tests with multiple fetches
const branchesMocks = generateMockedResponses(GitRefType.GIT_BRANCH, 16)
const tagsMocks = generateMockedResponses(GitRefType.GIT_TAG, 12)

// For tests with less than 10 results
const fewBranchesMocks = generateMockedResponses(GitRefType.GIT_BRANCH, 5)
const fewTagsMocks = generateMockedResponses(GitRefType.GIT_TAG, 2)

// For tests with filtered results
const filteredBranchesMocks = generateMockedResponses(GitRefType.GIT_BRANCH, 5, FILTERED_MOCK_PROPS.query)
const filteredTagsMocks = generateMockedResponses(GitRefType.GIT_TAG, 2, FILTERED_MOCK_PROPS.query)

// For tests with filtered but empty results
const emptyFilteredBranchesMocks = generateMockedResponses(GitRefType.GIT_BRANCH, 0, FILTERED_MOCK_PROPS.query)
const emptyFilteredTagsMocks = generateMockedResponses(GitRefType.GIT_TAG, 0, FILTERED_MOCK_PROPS.query)

export const EMPTY_MOCKS = [...emptyBranchesMocks, ...emptyTagsMocks]

export const DEFAULT_MOCKS = [...branchesMocks, ...tagsMocks]

export const FEW_RESULTS_MOCKS = [...fewBranchesMocks, ...fewTagsMocks]

export const FILTERED_MOCKS = [...filteredBranchesMocks, ...filteredTagsMocks]

export const EMPTY_FILTERED_MOCKS = [...emptyFilteredBranchesMocks, ...emptyFilteredTagsMocks]

export const GRAPHQL_ERROR_MOCKS = [
    { request: generateMockedRequest(GitRefType.GIT_BRANCH), result: { errors: [new GraphQLError('GraphQL error')] } },
    { request: generateMockedRequest(GitRefType.GIT_TAG), result: { errors: [new GraphQLError('GraphQL error')] } },
]

export const NETWORK_ERROR_MOCKS = [
    { request: generateMockedRequest(GitRefType.GIT_BRANCH), error: new Error('Network error') },
    { request: generateMockedRequest(GitRefType.GIT_TAG), error: new Error('Network error') },
]
