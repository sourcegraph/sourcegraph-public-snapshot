import { useState, useEffect, useCallback } from 'react'

import type { ApolloError } from '@apollo/client'

import { gql, useLazyQuery } from '@sourcegraph/http-client'

import type {
    RepositoryRecordedCommandsResult,
    RepositoryRecordedCommandsVariables,
    RepositoryRecordedCommandFields,
} from '../../../graphql-operations'

export const RepoPermissionsInfoQuery = gql`
    query RepoPermissionsInfo($repoID: ID!, $first: Int, $last: Int, $after: String, $before: String, $query: String) {
        node(id: $repoID) {
            __typename
            ... on Repository {
                ...RepoPermissionsInfoRepoNode
            }
        }
    }

    fragment RepoPermissionsInfoRepoNode on Repository {
        permissionsInfo {
            syncedAt
            updatedAt
            unrestricted
            users(first: $first, last: $last, after: $after, before: $before, query: $query) {
                nodes {
                    ...PermissionsInfoUserFields
                }
                totalCount
                pageInfo {
                    hasNextPage
                    hasPreviousPage
                    startCursor
                    endCursor
                }
            }
        }
    }

    fragment PermissionsInfoUserFields on PermissionsInfoUserNode {
        id
        reason
        updatedAt
        user {
            id
            username
            displayName
            email
            avatarURL
        }
    }
`

export const REPOSITORY_RECORDED_COMMANDS_QUERY = gql`
    query RepositoryRecordedCommands($id: ID!, $offset: Int!, $limit: Int) {
        node(id: $id) {
            __typename
            ... on Repository {
                isRecordingEnabled
                recordedCommands(offset: $offset, limit: $limit) {
                    nodes {
                        ...RepositoryRecordedCommandFields
                    }
                    totalCount
                    pageInfo {
                        hasNextPage
                    }
                }
            }
        }
    }

    fragment RepositoryRecordedCommandFields on RecordedCommand {
        path
        start
        duration
        command
        dir
        output
        isSuccess
    }
`

export const REPOSITORY_RECORDED_COMMANDS_LIMIT = 40

interface UseFetchRecordedCommandsResult {
    loading: boolean
    error: ApolloError | undefined
    recordedCommands: RepositoryRecordedCommandFields[]
    hasNextPage: boolean
    fetchMore: (offset: number) => void
    isRecordingEnabled?: boolean
}

export const useFetchRecordedCommands = (repoId: string): UseFetchRecordedCommandsResult => {
    const [recordedCommands, setRecordedCommands] = useState<RepositoryRecordedCommandFields[]>([])
    const [hasNextPage, setHasNextPage] = useState(false)
    const [isRecordingEnabled, setIsRecordingEnabled] = useState<boolean>()

    const [fetchRecordedCommands, { loading, error }] = useLazyQuery<
        RepositoryRecordedCommandsResult,
        RepositoryRecordedCommandsVariables
    >(REPOSITORY_RECORDED_COMMANDS_QUERY, {
        onCompleted: data => {
            const { node } = data

            if (!node) {
                throw new Error(`Repository with ID ${repoId} does not exist`)
            }

            if (node.__typename !== 'Repository') {
                throw new Error(`Node is a ${node.__typename}, not a Repository`)
            }

            setRecordedCommands(prev => [...prev, ...node.recordedCommands.nodes])
            setHasNextPage(node.recordedCommands.pageInfo.hasNextPage)
            setIsRecordingEnabled(node.isRecordingEnabled)
            return node.recordedCommands
        },
    })

    const getRecordedCommands = useCallback(
        (offset: number) => {
            fetchRecordedCommands({ variables: { id: repoId, offset, limit: REPOSITORY_RECORDED_COMMANDS_LIMIT } })
                // we'll catch the error in the tuple returned from `useLazyQuery`
                .catch(() => {})
        },
        [fetchRecordedCommands, repoId]
    )

    useEffect(() => {
        // We fetch the first set of recordedCommannds on mount
        getRecordedCommands(0)
    }, [fetchRecordedCommands, getRecordedCommands])

    return {
        recordedCommands,
        loading,
        error,
        hasNextPage,
        fetchMore: getRecordedCommands,
        isRecordingEnabled,
    }
}
