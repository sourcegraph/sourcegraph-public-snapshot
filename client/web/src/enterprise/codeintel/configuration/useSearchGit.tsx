import { ApolloError } from '@apollo/client'

import { gql, useQuery } from '@sourcegraph/shared/src/graphql/graphql'

import {
    SearchGitTagsResult,
    SearchGitTagsVariables,
    SearchGitBranchesResult,
    SearchGitBranchesVariables,
    RepositoryNameResult,
    RepositoryNameVariables,
} from '../../../graphql-operations'

interface SearchGitObjectResult {
    previewResult: GitObjectPreviewResult
    isLoadingPreview: boolean
    previewError: ApolloError | undefined
}

export interface GitObjectPreviewResult {
    preview: { name: string; revlike: string }[]
    totalCount: number
}

// Tags
export const SEARCH_GIT_TAGS = gql`
    query SearchGitTags($id: ID!, $query: String!) {
        node(id: $id) {
            ...RepositoryTagsFields
        }
    }

    fragment RepositoryTagsFields on Repository {
        __typename
        name
        tags(query: $query, first: 10) {
            nodes {
                displayName
            }

            totalCount
        }
    }
`
export const useSearchGitTags = (id: string, pattern: string): SearchGitObjectResult => {
    const { data, loading, error } = useQuery<SearchGitTagsResult, SearchGitTagsVariables>(SEARCH_GIT_TAGS, {
        variables: { id, query: pattern },
    })

    const { tags: { nodes = [], totalCount = 0 } = {}, name = '' } = data?.node || {}
    const previewResult = { preview: nodes.map(({ displayName: revlike }) => ({ name, revlike })), totalCount }

    return {
        previewResult,
        isLoadingPreview: loading,
        previewError: error,
    }
}

// Branches
export const SEARCH_GIT_BRANCHES = gql`
    query SearchGitBranches($id: ID!, $query: String!) {
        node(id: $id) {
            ...RepositoryBranchesFields
        }
    }

    fragment RepositoryBranchesFields on Repository {
        __typename
        name
        branches(query: $query, first: 10) {
            nodes {
                displayName
            }

            totalCount
        }
    }
`

export const useSearchGitBranches = (id: string, pattern: string): SearchGitObjectResult => {
    const { data, loading, error } = useQuery<SearchGitBranchesResult, SearchGitBranchesVariables>(
        SEARCH_GIT_BRANCHES,
        { variables: { id, query: pattern } }
    )
    const { branches: { nodes = [], totalCount = 0 } = {}, name = '' } = data?.node || {}
    const previewResult = { preview: nodes.map(({ displayName: revlike }) => ({ name, revlike })), totalCount }

    return {
        previewResult,
        isLoadingPreview: loading,
        previewError: error,
    }
}

// Commit aka repoName
export const SEARCH_REPO_NAME = gql`
    query RepositoryName($id: ID!) {
        node(id: $id) {
            ...RepositoryNameFields
        }
    }

    fragment RepositoryNameFields on Repository {
        __typename
        name
    }
`

export const useSearchRepoName = (id: string, pattern: string): SearchGitObjectResult => {
    const { data, loading, error } = useQuery<RepositoryNameResult, RepositoryNameVariables>(SEARCH_REPO_NAME, {
        variables: { id },
    })

    const previewResult = data?.node
        ? { preview: [{ name: data?.node.name, revlike: pattern }], totalCount: 1 }
        : { preview: [], totalCount: 0 }

    return {
        previewResult,
        isLoadingPreview: loading,
        previewError: error,
    }
}
