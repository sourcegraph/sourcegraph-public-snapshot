import { ApolloError } from '@apollo/client'
import React, { useMemo, useCallback } from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { useQuery } from '@sourcegraph/shared/src/graphql/graphql'

import {
    BatchChangesCodeHostsFields,
    GlobalBatchChangesCodeHostsResult,
    GlobalBatchChangesCodeHostsVariables,
    UserBatchChangesCodeHostsResult,
    UserBatchChangesCodeHostsVariables,
} from '../../../graphql-operations'

import { GLOBAL_BATCH_CHANGES_CODE_HOSTS, USER_BATCH_CHANGES_CODE_HOSTS } from './backend'
import { CodeHostConnectionNode } from './CodeHostConnectionNode'

const getBatchChangesCodeHosts = (
    data: UserBatchChangesCodeHostsResult | GlobalBatchChangesCodeHostsResult
): BatchChangesCodeHostsFields => {
    if ('batchChangesCodeHosts' in data) {
        return data.batchChangesCodeHosts
    }

    if (data.node === null) {
        throw new Error('User not found')
    }

    if (data.node.__typename !== 'User') {
        throw new Error(`Node is a ${data.node.__typename}, not a User`)
    }

    return data.node.batchChangesCodeHosts
}

interface CodeHostConnectionNodesProps {
    userID: Scalars['ID']
}

export const CodeHostConnectionNodes: React.FunctionComponent<CodeHostConnectionNodesProps> = ({ userID }) => {
    const variables = useMemo(
        () => ({
            user: userID,
            first: 15,
            after: null,
        }),
        [userID]
    )

    const response = useQuery<UserBatchChangesCodeHostsResult, UserBatchChangesCodeHostsVariables>(
        USER_BATCH_CHANGES_CODE_HOSTS,
        { variables }
    )

    const fetchMoreResults = useCallback(
        (cursor: string) =>
            response.fetchMore({
                variables: {
                    ...variables,
                    after: cursor,
                },
                updateQuery: (previousData, { fetchMoreResult }) => {
                    if (!fetchMoreResult?.node) {
                        return previousData
                    }

                    const previousCodeHosts = getBatchChangesCodeHosts(previousData)
                    const fetchMoreCodeHosts = getBatchChangesCodeHosts(fetchMoreResult)
                    fetchMoreCodeHosts.nodes = previousCodeHosts.nodes.concat(fetchMoreCodeHosts.nodes)

                    return {
                        node: {
                            __typename: 'User',
                            id: userID,
                            batchChangesCodeHosts: fetchMoreCodeHosts,
                        },
                    }
                },
            }),
        [response, variables, userID]
    )

    return <CodeHostConnectionNodesUI {...response} fetchMoreResults={fetchMoreResults} userID={userID} />
}

export const GlobalCodeHostConnectionNodes: React.FunctionComponent = () => {
    const variables = useMemo(
        () => ({
            first: 15,
            after: null,
        }),
        []
    )

    const response = useQuery<GlobalBatchChangesCodeHostsResult, GlobalBatchChangesCodeHostsVariables>(
        GLOBAL_BATCH_CHANGES_CODE_HOSTS,
        { variables }
    )

    const fetchMoreResults = useCallback(
        (cursor: string) =>
            response.fetchMore({
                variables: {
                    ...variables,
                    after: cursor,
                },
                updateQuery: (previousData, { fetchMoreResult }) => {
                    if (!fetchMoreResult) {
                        return previousData
                    }

                    const previousCodeHosts = getBatchChangesCodeHosts(previousData)
                    const fetchMoreCodeHosts = getBatchChangesCodeHosts(fetchMoreResult)
                    fetchMoreCodeHosts.nodes = previousCodeHosts.nodes.concat(fetchMoreCodeHosts.nodes)

                    return {
                        batchChangesCodeHosts: fetchMoreCodeHosts,
                    }
                },
            }),
        [response, variables]
    )

    return <CodeHostConnectionNodesUI {...response} fetchMoreResults={fetchMoreResults} userID={null} />
}

const CodeHostConnectionNodesUI: React.FunctionComponent<{
    data?: UserBatchChangesCodeHostsResult | GlobalBatchChangesCodeHostsResult
    error?: ApolloError
    loading: boolean
    userID: Scalars['ID'] | null
    fetchMoreResults: (cursor: string) => void
}> = ({ data, error, loading, userID, fetchMoreResults }) => {
    if (error) {
        throw error
    }

    if (!data) {
        return null
    }

    if (loading) {
        return <LoadingSpinner className="icon-inline" />
    }

    const { nodes, pageInfo } = getBatchChangesCodeHosts(data)

    return (
        <>
            {nodes.map((node, index) => (
                <CodeHostConnectionNode key={index} node={node} userID={userID} />
            ))}
            {pageInfo.hasNextPage && (
                <button
                    type="button"
                    className="btn btn-sm btn-link"
                    onClick={() => fetchMoreResults(pageInfo.endCursor || '')}
                >
                    Show more
                </button>
            )}
        </>
    )
}
