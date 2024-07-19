import { useCallback, useRef } from 'react'

import type { ApolloError, QueryResult, WatchQueryFetchPolicy } from '@apollo/client'

import { useQuery, type GraphQLResult } from '@sourcegraph/http-client'
import { useInterval } from '@sourcegraph/wildcard'

import type { Connection } from '../ConnectionType'
import { asGraphQLResult, hasNextPage } from '../utils'

import {
    useConnectionStateOrMemoryFallback,
    useConnectionStateWithImplicitPageSize,
    type UseConnectionStateResult,
} from './connectionState'
import { DEFAULT_PAGE_SIZE } from './usePageSwitcherPagination'

export interface ShowMoreConnectionQueryArguments {
    first?: number | null
    after?: string | null
}

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
    /** The number of items per page. Defaults to 20. */
    pageSize?: number

    /** Allows modifying how the query interacts with the Apollo cache. */
    fetchPolicy?: WatchQueryFetchPolicy

    /** Allows specifying the Apollo error policy. */
    errorPolicy?: 'all' | 'none' | 'ignore'

    /**
     * Set to enable polling of all the nodes currently loaded in the connection.
     *
     * NOTE: USE WITH CAUTION. Polling connection data is currently not recommended as
     * Apollo does not reliably merge data from "Show More" (`fetchMore`) responses with
     * the data from polling responses when the two are in flight simultaneously.
     */
    pollInterval?: number

    /** Allows running an optional callback on any successful request. */
    onCompleted?: (data: TResult) => void
    onError?: (error: ApolloError) => void

    /**
     * useAlternateAfterCursor is used to indicate that a custom field instead of the
     * standard "after" field is used to for pagination. This is typically a
     * workaround for existing APIs where after may already be in use for
     * another field.
     */
    useAlternateAfterCursor?: boolean

    /** Skip the query if this condition is true. */
    skip?: boolean
}

interface UseShowMorePaginationParameters<TResult, TVariables, TData, TState extends ShowMoreConnectionQueryArguments> {
    query: string
    variables: Omit<TVariables, keyof ShowMoreConnectionQueryArguments | 'afterCursor'>
    getConnection: (result: GraphQLResult<TResult>) => Connection<TData>
    options?: UseShowMorePaginationConfig<TResult>

    /**
     * The value and setter for the state parameters (such as `first`, `after`, `before`, and
     * filters).
     */
    state?: UseConnectionStateResult<TState>
}

/**
 * Request a GraphQL connection query and handle pagination options. When the user presses "show
 * more", all of the previous items still remain visible. This is for GraphQL connections that only
 * support fetching results in one direction (support for `first` is required, and support for
 * `after`/`endCursor` is optional) and/or where this "show more" behavior is desirable.
 *
 * For paginated behavior (where the user can press "next page" and see a different set of results),
 * and if the GraphQL connection supports full
 * `endCursor`/`startCursor`/`after`/`before`/`first`/`last`, use {@link usePageSwitcherPagination}
 * instead.
 *
 * Valid queries should follow the connection specification at
 * https://relay.dev/graphql/connections.htm.
 * @param query The GraphQL connection query
 * @param variables The GraphQL connection variables
 * @param getConnection A function that filters and returns the relevant data from the connection
 * response.
 * @param options Additional configuration options
 */
export const useShowMorePagination = <
    TResult,
    TVariables extends ShowMoreConnectionQueryArguments,
    TData,
    TState extends ShowMoreConnectionQueryArguments = ShowMoreConnectionQueryArguments &
        Partial<Record<string | 'query', string>>
>({
    query,
    variables,
    getConnection: getConnectionFromGraphQLResult,
    options,
    state,
}: UseShowMorePaginationParameters<TResult, TVariables, TData, TState>): UseShowMorePaginationResult<
    TResult,
    TData
> => {
    const pageSize = options?.pageSize ?? DEFAULT_PAGE_SIZE
    const [connectionState, setConnectionState] = useConnectionStateWithImplicitPageSize(
        useConnectionStateOrMemoryFallback(state),
        pageSize
    )

    const first = connectionState.first ?? pageSize

    /**
     * Map over Apollo results to provide type-compatible `GraphQLResult`s for consumers.
     * This ensures good interoperability between `FilteredConnection` and `useShowMorePagination`.
     */
    const getConnection = ({ data, error }: Pick<QueryResult<TResult>, 'data' | 'error'>): Connection<TData> => {
        const result = asGraphQLResult({ data, errors: error?.graphQLErrors || [] })
        return getConnectionFromGraphQLResult(result)
    }

    // These will change when the user clicks "show more", but we want those fetches to go through
    // `fetchMore` and not through `useQuery` noticing that its variables have changed, so
    // use a ref to achieve that.
    const initialPaginationArgs = useRef({
        first,
        after: connectionState.after,
    })

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
            ...initialPaginationArgs.current,
        } as TVariables,
        notifyOnNetworkStatusChange: true, // Ensures loading state is updated on `fetchMore`
        skip: options?.skip,
        fetchPolicy: options?.fetchPolicy,
        onCompleted: options?.onCompleted,
        onError: options?.onError,
        errorPolicy: options?.errorPolicy,
    })

    const data = currentData ?? previousData
    const connection = data ? getConnection({ data, error }) : undefined

    const fetchMoreData = async (): Promise<void> => {
        const cursor = connection?.pageInfo?.endCursor

        // Use cursor paging if possible, otherwise fallback to increasing `first`.
        const afterVariables: { after?: string; first?: number; afterCursor?: string } = {}
        if (cursor) {
            if (options?.useAlternateAfterCursor) {
                afterVariables.afterCursor = cursor
            } else {
                afterVariables.after = cursor
            }
            afterVariables.first = pageSize
        } else {
            afterVariables.first = first + pageSize
        }

        // Don't reflect `after` in the URL because the page shows *all* items from the beginning,
        // not just those after the `after` cursor. The cursor is only used in the GraphQL request.
        setConnectionState(prev => ({ ...prev, first: first + pageSize }))

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
                    // With batch-based pagination, we have all the results already in
                    // `fetchMoreResult`. We already updated `first` via `setConnectionState` above
                    // to fetch more results next time.
                }

                return fetchMoreResult
            },
        })
    }

    // Refetch the current nodes
    const refetchAll = useCallback(async (): Promise<void> => {
        // No change in connection state (`state.setValue`) needed.
        await refetch({
            ...variables,
            first,
        } as Partial<TVariables>)
    }, [first, refetch, variables])

    // Refetch the first page. Use this function if the number of nodes in the
    // connection might have changed since the last refetch.
    const refetchFirst = useCallback(async (): Promise<void> => {
        // Reset connection state to just fetch the first page.
        setConnectionState(prev => ({ ...prev, first: pageSize }))
        await refetch({
            ...variables,
            first,
        } as Partial<TVariables>)
    }, [first, pageSize, refetch, setConnectionState, variables])

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
