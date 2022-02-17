import { ApolloError, QueryResult, WatchQueryFetchPolicy } from '@apollo/client'
import { useCallback } from 'react'

import { GraphQLResult, useQuery } from '@sourcegraph/http-client'
import { asGraphQLResult } from '@sourcegraph/web/src/components/FilteredConnection/utils'
import { useInterval } from '@sourcegraph/wildcard'

import { ConnectionQueryArguments } from '../components/FilteredConnection/ConnectionType'
import { RefPanelLsifDataFields } from '../graphql-operations'

export interface UsePreciseCodeIntelResult {
    lsifData?: RefPanelLsifDataFields
    error?: ApolloError
    refetchAll: () => void
    loading: boolean

    referencesHasNextPage: boolean
    implementationsHasNextPage: boolean
    fetchMoreReferences: () => void
    fetchMoreImplementations: () => void

    startPolling: (pollInterval: number) => void
    stopPolling: () => void
}

interface UsePreciseCodeIntelConfig {
    /** Set if query variables should be updated in and derived from the URL */
    useURL?: boolean
    /** Allows modifying how the query interacts with the Apollo cache */
    fetchPolicy?: WatchQueryFetchPolicy
    /** Set to enable polling of all the nodes currently loaded in the connection */
    pollInterval?: number
}

interface UsePreciseCodeIntelParameters<TResult, TVariables> {
    query: string
    variables: TVariables & ConnectionQueryArguments
    getConnection: (result: GraphQLResult<TResult>) => RefPanelLsifDataFields
    options?: UsePreciseCodeIntelConfig
}

export const usePreciseCodeIntel = <TResult, TVariables, TData>({
    query,
    variables,
    getConnection: getLsifDataFromGraphQLResult,
    options,
}: UsePreciseCodeIntelParameters<TResult, TVariables>): UsePreciseCodeIntelResult => {
    const { data, error, loading, fetchMore, refetch } = useQuery<TResult, TVariables>(query, {
        variables,
        notifyOnNetworkStatusChange: true, // Ensures loading state is updated on `fetchMore`
        fetchPolicy: options?.fetchPolicy,
    })

    /**
     * Map over Apollo results to provide type-compatible `GraphQLResult`s for consumers.
     * This ensures good interoperability between `FilteredConnection` and `useConnection`.
     */
    const getLsifData = ({ data, error }: Pick<QueryResult<TResult>, 'data' | 'error'>): RefPanelLsifDataFields => {
        const result = asGraphQLResult({ data, errors: error?.graphQLErrors || [] })
        return getLsifDataFromGraphQLResult(result)
    }

    const lsifData = data ? getLsifData({ data, error }) : undefined

    const fetchMoreReferences = async (): Promise<void> => {
        const cursor = lsifData?.references.pageInfo?.endCursor

        await fetchMore({
            variables: {
                ...variables,
                ...{ afterReferences: cursor },
            },
            updateQuery: (previousResult, { fetchMoreResult }) => {
                if (!fetchMoreResult) {
                    return previousResult
                }

                if (cursor) {
                    const previousData = getLsifData({ data: previousResult })
                    const previousImplementationNodes = previousData.implementations.nodes
                    const previousReferencesNodes = previousData.references.nodes

                    const fetchMoreData = getLsifData({ data: fetchMoreResult })
                    fetchMoreData.implementations.nodes = previousImplementationNodes
                    fetchMoreData.references.nodes.unshift(...previousReferencesNodes)
                }

                return fetchMoreResult
            },
        })
    }

    const fetchMoreImplementations = async (): Promise<void> => {
        const cursor = lsifData?.implementations.pageInfo?.endCursor

        await fetchMore({
            variables: {
                ...variables,
                ...{ afterImplementations: cursor },
            },
            updateQuery: (previousResult, { fetchMoreResult }) => {
                if (!fetchMoreResult) {
                    return previousResult
                }

                if (cursor) {
                    const previousData = getLsifData({ data: previousResult })
                    const previousImplementationNodes = previousData.implementations.nodes
                    const previousReferencesNodes = previousData.references.nodes

                    const fetchMoreData = getLsifData({ data: fetchMoreResult })
                    fetchMoreData.implementations.nodes.unshift(...previousImplementationNodes)
                    fetchMoreData.references.nodes = previousReferencesNodes
                }

                return fetchMoreResult
            },
        })
    }

    const refetchAll = useCallback(async (): Promise<void> => {
        const referencesLength = lsifData?.references.nodes.length ?? 50
        const implementationsLength = lsifData?.implementations.nodes.length ?? 50
        const first = Math.max(referencesLength, implementationsLength)

        await refetch({
            ...variables,
            first,
        })
    }, [lsifData?.references.nodes.length, lsifData?.implementations.nodes.length, refetch, variables])

    // We use `refetchAll` to poll for all of the nodes currently loaded in the
    // connection, vs. just providing a `pollInterval` to the underlying `useQuery`, which
    // would only poll for the first page of results.
    const { startExecution, stopExecution } = useInterval(refetchAll, options?.pollInterval || -1)

    return {
        lsifData,
        loading,
        error,
        fetchMoreReferences,
        fetchMoreImplementations,
        refetchAll,
        referencesHasNextPage: lsifData ? lsifData.references.pageInfo.endCursor !== null : false,
        implementationsHasNextPage: lsifData ? lsifData.implementations.pageInfo.endCursor !== null : false,
        startPolling: startExecution,
        stopPolling: stopExecution,
    }
}
