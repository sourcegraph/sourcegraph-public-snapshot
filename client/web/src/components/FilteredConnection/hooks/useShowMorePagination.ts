import { useCallback, useMemo, useRef } from 'react'

import type { ApolloError, QueryResult, WatchQueryFetchPolicy } from '@apollo/client'

import { type GraphQLResult, useQuery } from '@sourcegraph/http-client'
import { useSearchParameters, useInterval } from '@sourcegraph/wildcard'

import type { Connection, ConnectionQueryArguments } from '../ConnectionType'
import { asGraphQLResult, hasNextPage, parseQueryInt } from '../utils'

import { useShowMorePaginationUrl } from './useShowMorePaginationUrl'

export interface UseShowMorePaginationResult<TResult, TData> {
    data?: TResult
    connection?: Connection<TData>
    error?: ApolloError
    fetchMore: () => void
    refetchAll: () => void
    refetchFirst: () => void
    loading: boolean
    hasNextPage: boolean
    /**
     * NOTE: USE WITH CAUTION. Polling connection data is currently not recommended as
     * Apollo does not reliably merge data from "Show More" (`fetchMore`) responses with
     * the data from polling responses when the two are in flight simultaneously.
     */
    startPolling: (pollInterval: number) => void
    stopPolling: () => void
}

interface UseShowMorePaginationConfig<TResult> {
    /** Set if query variables should be updated in and derived from the URL */
    useURL?: boolean
    /** Allows modifying how the query interacts with the Apollo cache */
    fetchPolicy?: WatchQueryFetchPolicy
    /** Allows specifying the Apollo error policy */
    errorPolicy?: 'all' | 'none' | 'ignore'
    /**
     * Set to enable polling of all the nodes currently loaded in the connection.
     *
     * NOTE: USE WITH CAUTION. Polling connection data is currently not recommended as
     * Apollo does not reliably merge data from "Show More" (`fetchMore`) responses with
     * the data from polling responses when the two are in flight simultaneously.
     */
    pollInterval?: number
    /** Allows running an optional callback on any successful request */
    onCompleted?: (data: TResult) => void
    onError?: (error: ApolloError) => void

    // useAlternateAfterCursor is used to indicate that a custom field instead of the
    // standard "after" field is used to for pagination. This is typically a
    // workaround for existing APIs where after may already be in use for
    // another field.
    useAlternateAfterCursor?: boolean
    /** Skip the query if this condition is true */
    skip?: boolean
}

interface UseShowMorePaginationParameters<TResult, TVariables, TData> {
    query: string
    variables: TVariables & ConnectionQueryArguments
    getConnection: (result: GraphQLResult<TResult>) => Connection<TData>
    options?: UseShowMorePaginationConfig<TResult>
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
export const useShowMorePagination = <TResult, TVariables extends {}, TData>({
    query,
    variables,
    getConnection: getConnectionFromGraphQLResult,
    options,
}: UseShowMorePaginationParameters<TResult, TVariables, TData>): UseShowMorePaginationResult<TResult, TData> => {
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
        // We only need these controls for the initial request. We do not care about dependency updates.
        // eslint-disable-next-line react-hooks/exhaustive-deps
        []
    )

    /**
     * Initial query of the hook.
     * Subsequent requests (such as further pagination) will be handled through `fetchMore`
     */
    const {
        data: currentData,
        previousData,
        error,
        loading,
        fetchMore,
        refetch,
    } = useQuery<TResult, TVariables>(query, {
        variables: {
            ...variables,
            ...initialControls,
        },
        notifyOnNetworkStatusChange: true, // Ensures loading state is updated on `fetchMore`
        skip: options?.skip,
        fetchPolicy: options?.fetchPolicy,
        onCompleted: options?.onCompleted,
        onError: options?.onError,
        errorPolicy: options?.errorPolicy,
    })

    /**
     * Map over Apollo results to provide type-compatible `GraphQLResult`s for consumers.
     * This ensures good interoperability between `FilteredConnection` and `useShowMorePagination`.
     */
    const getConnection = ({ data, error }: Pick<QueryResult<TResult>, 'data' | 'error'>): Connection<TData> => {
        const result = asGraphQLResult({ data, errors: error?.graphQLErrors || [] })
        return getConnectionFromGraphQLResult(result)
    }

    const data = currentData ?? previousData
    const connection = data ? getConnection({ data, error }) : undefined

    useShowMorePaginationUrl({
        enabled: options?.useURL,
        first: firstReference.current,
        visibleResultCount: connection?.nodes.length,
    })

    const fetchMoreData = async (): Promise<void> => {
        const cursor = connection?.pageInfo?.endCursor

        // Use cursor paging if possible, otherwise fallback to multiplying `first`.
        const afterVariables: { after?: string; first?: number; afterCursor?: string } = {}
        if (cursor) {
            if (options?.useAlternateAfterCursor) {
                afterVariables.afterCursor = cursor
            } else {
                afterVariables.after = cursor
            }
        } else {
            afterVariables.first = firstReference.current.actual * 2
        }
        await fetchMore({
            variables: {
                ...variables,
                ...afterVariables,
            },
            updateQuery: (previousResult, { fetchMoreResult }) => {
                if (!fetchMoreResult) {
                    return previousResult
                }

                if (cursor) {
                    // Update resultant data in the cache by prepending the `previousResult`s to the
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

    // Refetch the current nodes
    const refetchAll = useCallback(async (): Promise<void> => {
        const first = connection?.nodes.length || firstReference.current.actual

        await refetch({
            ...variables,
            first,
        })
    }, [connection?.nodes.length, refetch, variables])

    // Refetch the first page. Use this function if the number of nodes in the
    // connection might have changed since the last refetch.
    const refetchFirst = useCallback(async (): Promise<void> => {
        await refetch({
            ...variables,
            first,
        })
    }, [first, refetch, variables])

    // We use `refetchAll` to poll for all the nodes currently loaded in the
    // connection, vs. just providing a `pollInterval` to the underlying `useQuery`, which
    // would only poll for the first page of results.
    const { startExecution, stopExecution } = useInterval(refetchAll, options?.pollInterval || -1)

    return {
        data,
        connection,
        loading,
        error,
        fetchMore: fetchMoreData,
        refetchFirst,
        refetchAll,
        hasNextPage: connection ? hasNextPage(connection) : false,
        startPolling: startExecution,
        stopPolling: stopExecution,
    }
}
