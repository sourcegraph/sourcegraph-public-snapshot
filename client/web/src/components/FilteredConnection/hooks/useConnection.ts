import { useCallback, useMemo, useRef, useState } from 'react'

import { ApolloError, QueryResult, WatchQueryFetchPolicy } from '@apollo/client'

import { GraphQLResult, useQuery } from '@sourcegraph/http-client'
import { useSearchParameters, useInterval, useDeepMemo } from '@sourcegraph/wildcard'

import { Connection, ConnectionQueryArguments } from '../ConnectionType'
import { asGraphQLResult, hasNextPage, parseQueryInt } from '../utils'

import { useConnectionUrl } from './useConnectionUrl'

export interface UseConnectionResult<TData> {
    connection?: Connection<TData>
    error?: ApolloError
    fetchMore: () => void
    refetchAll: () => void
    loading: boolean
    hasNextPage: boolean
    startPolling: (pollInterval: number) => void
    stopPolling: () => void
}

interface UseConnectionConfig<TResult> {
    /** Set if query variables should be updated in and derived from the URL */
    useURL?: boolean
    /** Allows modifying how the query interacts with the Apollo cache */
    fetchPolicy?: WatchQueryFetchPolicy | 'component-cache-and-network'
    /** Set to enable polling of all the nodes currently loaded in the connection */
    pollInterval?: number
    /** Allows running an optional callback on any successful request */
    onCompleted?: (data: TResult) => void
}

interface UseConnectionParameters<TResult, TVariables, TData> {
    query: string
    variables: TVariables & ConnectionQueryArguments
    getConnection: (result: GraphQLResult<TResult>) => Connection<TData>
    options?: UseConnectionConfig<TResult>
}

const DEFAULT_AFTER: ConnectionQueryArguments['after'] = undefined
const DEFAULT_FIRST: ConnectionQueryArguments['first'] = 20

/**
 * Request a GraphQL connection query and handle pagination options.
 * Valid queries should follow the connection specification at https://relay.dev/graphql/connections.htm
 *
 * @param query The GraphQL connection query
 * @param variables The GraphQL connection variables
 * @param getConnection A function that filters and returns the relevant data from the connection response.
 * @param options Additional configuration options
 */
export const useConnection = <TResult, TVariables, TData>({
    query,
    variables,
    getConnection: getConnectionFromGraphQLResult,
    options,
}: UseConnectionParameters<TResult, TVariables, TData>): UseConnectionResult<TData> => {
    /**
     * Local connections is intended to track pagination _without_ using the Apollo cache.
     * This state will only be used if options.fetchPolicy is 'component-cache-and-network'
     *
     * Note: We use a `Map` here in order to keep track of connections against provided variables.
     * This is so that consumers are able to modify the connection request without having
     * to deal with potentially state data.
     * This is especially important when using polling, as there may already be in-flight requests.
     */
    const [localConnections, setLocalConnections] = useState<Map<TVariables, TResult>>(new Map())
    const skipApolloCache = options?.fetchPolicy === 'component-cache-and-network'

    const searchParameters = useSearchParameters()

    const { first = DEFAULT_FIRST, after = DEFAULT_AFTER } = variables
    const firstReference = useRef({
        /**
         * The number of results that we will typically want to load in the next request (unless `visible` is used).
         * This value will typically be static for cursor-based pagination, but will be dynamic for batch-based pagination.
         */
        actual: (options?.useURL && parseQueryInt(searchParameters, 'first')) || first,
        /**
         * Primarily used to determine original request state for URL search parameter logic.
         */
        default: first,
    })

    const initialVariables = useDeepMemo(variables)
    const initialControls = useMemo(
        () => ({
            /**
             * The `first` variable for our **initial** query.
             * If this is our first query and we were supplied a value for `visible` load that many results.
             * If we weren't given such a value or this is a subsequent request, only ask for one page of results.
             *
             * 'visible' is the number of results that were visible from previous requests. The initial request of
             * a result set will load `visible` items, then will request `first` items on each subsequent
             * request. This has the effect of loading the correct number of visible results when a URL
             * is copied during pagination. This value is only useful with cursor-based paging for the initial request.
             */
            first: (options?.useURL && parseQueryInt(searchParameters, 'visible')) || firstReference.current.actual,
            /**
             * The `after` variable for our **initial** query.
             * Subsequent requests through `fetchMore` will use a valid `cursor` value here, where possible.
             */
            after: (options?.useURL && searchParameters.get('after')) || after,
        }),
        // We only need these controls for the inital request. We do not care about dependency updates.
        // eslint-disable-next-line react-hooks/exhaustive-deps
        []
    )

    /**
     * Initial query of the hook.
     * Subsequent requests (such as further pagination) will be handled through `fetchMore`
     */
    const { data, error, loading, fetchMore, refetch } = useQuery<TResult, TVariables>(query, {
        variables: {
            ...initialVariables,
            ...initialControls,
        },
        notifyOnNetworkStatusChange: true, // Ensures loading state is updated on `fetchMore`
        fetchPolicy: skipApolloCache ? 'no-cache' : (options?.fetchPolicy as WatchQueryFetchPolicy),
        onCompleted: options?.onCompleted,
    })

    /**
     * Map over Apollo results to provide type-compatible `GraphQLResult`s for consumers.
     * This ensures good interoperability between `FilteredConnection` and `useConnection`.
     */
    const getConnection = useCallback(
        ({ data, error }: Pick<QueryResult<TResult>, 'data' | 'error'>): Connection<TData> => {
            const result = asGraphQLResult({ data, errors: error?.graphQLErrors || [] })
            return getConnectionFromGraphQLResult(result)
        },
        [getConnectionFromGraphQLResult]
    )

    let connection: Connection<TData> | undefined

    if (skipApolloCache && localConnections.has(initialVariables)) {
        // `Data` cannot be relied on without the Apollo cache, we use our stateful value instead.
        connection = getConnection({ data: localConnections.get(initialVariables) })
    } else if (data) {
        connection = getConnection({ data, error })
    }

    useConnectionUrl({
        enabled: options?.useURL,
        first: firstReference.current,
        visibleResultCount: connection?.nodes.length,
    })

    const fetchMoreData = async (): Promise<void> => {
        const cursor = connection?.pageInfo?.endCursor
        const queryVariables = {
            ...variables,
            // Use cursor paging if possible, otherwise fallback to multiplying `first`
            ...(cursor ? { after: cursor } : { first: firstReference.current.actual * 2 }),
        }

        if (skipApolloCache) {
            /**
             * We're using the local cache.
             * We do not use `updateQuery` as it is used to update the Apollo cache,
             * instead we will update our local cache in state using the resolved value directly.
             */
            const nextResult = await fetchMore({ variables: queryVariables })

            if (cursor) {
                setLocalConnections(connections => {
                    // Get the previous result from the local cache.
                    // Note: If this is the first subsequent request, the local cache will not yet be populated.
                    // Instead, we fall back to data as that is the latest response.
                    const previousResult = connections.get(initialVariables) || data
                    const previousNodes = previousResult ? getConnection({ data: previousResult }).nodes : []
                    // Update resultant data in the local cache by prepending the `previousResult`s to the
                    // `fetchMoreResult`s. We must rely on the consumer-provided `getConnection` here in
                    // order to access and modify the actual `nodes` in the connection response because we
                    // don't know the exact response structure
                    getConnection({ data: nextResult.data }).nodes.unshift(...previousNodes)
                    return connections.set(initialVariables, nextResult.data)
                })
            } else {
                // With batch-based pagination, we have all the results already in `fetchMoreResult`,
                // we just need to update `first` to fetch more results next time
                firstReference.current.actual *= 2
            }
        } else {
            /**
             * We're using the Apollo cache.
             * We will use `updateQuery` to allow us to directly modify the current cache item
             * and append the new results globally.
             */
            await fetchMore({
                variables: queryVariables,
                updateQuery: (previousResult, { fetchMoreResult }) => {
                    if (!fetchMoreResult) {
                        return previousResult
                    }

                    if (cursor) {
                        // Update resultant data in the Apollo cache by prepending the `previousResult`s to the
                        // `fetchMoreResult`s. We must rely on the consumer-provided `getConnection` here in
                        // order to access and modify the actual `nodes` in the connection response because we
                        // don't know the exact response structure
                        const previousNodes = getConnection({ data: previousResult }).nodes
                        getConnection({ data: fetchMoreResult }).nodes.unshift(...previousNodes)
                    } else {
                        // With batch-based pagination, we have all the results already in `fetchMoreResult`,
                        // we just need to update `first` to fetch more results next time
                        firstReference.current.actual *= 2
                    }

                    return fetchMoreResult
                },
            })
        }
    }

    const refetchAll = useCallback(async (): Promise<void> => {
        const first = connection?.nodes.length || firstReference.current.actual

        const { data: refetchedData } = await refetch({
            ...variables,
            first,
        })

        if (skipApolloCache) {
            /**
             * We're using the local cache.
             * We need to overwrite the latest cached value.
             */
            setLocalConnections(connections => {
                // Note: If this is the first subsequent request, the local cache will not yet be populated.
                // Instead, we fall back to data as that is the latest response.
                const previousResult = connections.get(initialVariables) || data
                const refetchedConnection = getConnection({ data: refetchedData })
                if (
                    previousResult &&
                    getConnection({ data: previousResult }).nodes.length !== refetchedConnection.nodes.length
                ) {
                    // We expect the refetched request to have the same amount of results as the previous one.
                    // If this is not the case, we should fallback to the previous result.
                    // Note: This will typically happen if a user clicks "Show more" as the polled request is initiated.
                    return connections
                }

                // Overwrite the local cache to use the latest data.
                return connections.set(initialVariables, refetchedData)
            })
        }
    }, [connection?.nodes.length, data, getConnection, initialVariables, refetch, skipApolloCache, variables])

    // We use `refetchAll` to poll for all of the nodes currently loaded in the
    // connection, vs. just providing a `pollInterval` to the underlying `useQuery`, which
    // would only poll for the first page of results.
    const { startExecution, stopExecution } = useInterval(refetchAll, options?.pollInterval || -1)

    return {
        connection,
        loading,
        error,
        fetchMore: fetchMoreData,
        refetchAll,
        hasNextPage: connection ? hasNextPage(connection) : false,
        startPolling: startExecution,
        stopPolling: stopExecution,
    }
}
