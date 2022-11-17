import { useCallback, useMemo, useRef } from 'react'

import { ApolloError, QueryResult, WatchQueryFetchPolicy } from '@apollo/client'

import { GraphQLResult, useQuery } from '@sourcegraph/http-client'
import { useSearchParameters, useInterval } from '@sourcegraph/wildcard'

import { Connection, PaginatedConnectionQueryArguments, PaginatedConnection } from '../ConnectionType'
import { asGraphQLResult, hasNextPage, parseQueryInt } from '../utils'

export interface UsePaginatedConnectionResult<TData> {
    connection?: Connection<TData>
    error?: ApolloError
    nextPage: () => void
    previousPage: () => void
    loading: boolean
    hasNextPage: boolean
    hasPreviousPage: boolean
}

interface UsePaginatedConnectionConfig<TResult> {
    /** The number of items per page, defaults to 20 */
    pageSize?: number
    /** Set if query variables should be updated in and derived from the URL */
    useURL?: boolean
    /** Allows modifying how the query interacts with the Apollo cache */
    fetchPolicy?: WatchQueryFetchPolicy
    /** Allows running an optional callback on any successful request */
    onCompleted?: (data: TResult) => void
}

interface UsePaginatedConnectionParameters<TResult, TVariables extends PaginatedConnectionQueryArguments, TData> {
    query: string
    variables: Omit<TVariables, 'first' | 'last' | 'before' | 'after'>
    getConnection: (result: GraphQLResult<TResult>) => PaginatedConnection<TData>
    options?: UsePaginatedConnectionConfig<TResult>
}

const DEFAULT_CONNECTION_ARGUMENTS: PaginatedConnectionQueryArguments = {
    first: 20,
}

/**
 * Request a GraphQL connection query and handle pagination options.
 * Valid queries should follow the connection specification at https://relay.dev/graphql/connections.htm
 *
 * @param query The GraphQL connection query
 * @param variables The GraphQL connection variables
 * @param getConnection A function that filters and returns the relevant data from the connection response.
 * @param options Additional configuration options
 */
export const usePaginatedConnection = <TResult, TVariables extends PaginatedConnectionQueryArguments, TData>({
    query,
    variables,
    getConnection: getConnectionFromGraphQLResult,
    options,
}: UsePaginatedConnectionParameters<TResult, TVariables, TData>): UsePaginatedConnectionResult<TData> => {
    const searchParameters = useSearchParameters()

    const { first = DEFAULT_FIRST, after = DEFAULT_AFTER } = variables
    const paginationRef = useRef({
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

    const initialControls: PaginatedConnectionQueryArguments = useMemo(
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
            first: (options?.useURL && parseQueryInt(searchParameters, 'visible')) || paginationRef.current.actual,
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
            ...variables,
            ...initialControls,
        },
        notifyOnNetworkStatusChange: true, // Ensures loading state is updated on `fetchMore`
        fetchPolicy: options?.fetchPolicy,
        onCompleted: options?.onCompleted,
    })

    /**
     * Map over Apollo results to provide type-compatible `GraphQLResult`s for consumers.
     * This ensures good interoperability between `FilteredConnection` and `useConnection`.
     */
    const getConnection = ({ data, error }: Pick<QueryResult<TResult>, 'data' | 'error'>): Connection<TData> => {
        const result = asGraphQLResult({ data, errors: error?.graphQLErrors || [] })
        return getConnectionFromGraphQLResult(result)
    }

    const connection = data ? getConnection({ data, error }) : undefined

    // useConnectionUrl({
    //     enabled: options?.useURL,
    //     first: paginationRef.current,
    //     visibleResultCount: connection?.nodes.length,
    // })

    const nextPage = async (): Promise<void> => {
        const cursor = connection?.pageInfo?.endCursor

        await fetchMore({
            variables: {
                ...variables,
                // Use cursor paging if possible, otherwise fallback to multiplying `first`
                ...(cursor ? { after: cursor } : { first: paginationRef.current.actual * 2 }),
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
                    paginationRef.current.actual *= 2
                }

                return fetchMoreResult
            },
        })
    }

    const refe
    // We use `refetchAll`s to poll for all of the nodes currently loaded in the
    // connection, vs. just providing a `pollInterval` to the underlying `useQuery`, which
    // would only poll for the first page of results.
    const { startExecution, stopExecution } = useInterval(refetchAll, options?.pollInterval || -1)

    return {
        connection,
        loading,
        error,
        nextPage,
        hasNextPage: connection ? hasNextPage(connection) : false,
        hasPreviousPage: connection ? hasPreviousPage(connection) : false,
        startPolling: startExecution,
        stopPolling: stopExecution,
    }
}
