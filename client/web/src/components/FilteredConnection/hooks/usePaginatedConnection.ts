import { useCallback } from 'react'

import { ApolloError, QueryResult, WatchQueryFetchPolicy } from '@apollo/client'

import { GraphQLResult, useQuery } from '@sourcegraph/http-client'
import { useSearchParameters } from '@sourcegraph/wildcard'

import { Connection, PaginatedConnectionQueryArguments, PaginatedConnection } from '../ConnectionType'
import { asGraphQLResult } from '../utils'

export interface UsePaginatedConnectionResult<TData> {
    connection?: Connection<TData>
    error?: ApolloError
    nextPage: () => void
    previousPage: () => void
    loading: boolean
    firstPage: () => void
    lastPage: () => void
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
    const pageSize = options?.pageSize ?? 20

    const searchParameters = useSearchParameters()

    // const { first = DEFAULT_FIRST, after = DEFAULT_AFTER } = variables
    // const paginationRef = useRef({
    //     /**
    //      * The number of results that we will typically want to load in the next request (unless `visible` is used).
    //      * This value will typically be static for cursor-based pagination, but will be dynamic for batch-based pagination.
    //      */
    //     actual: (options?.useURL && parseQueryInt(searchParameters, 'first')) || first,
    //     /**
    //      * Primarily used to determine original request state for URL search parameter logic.
    //      */
    //     default: first,
    // })

    // const initialControls: PaginatedConnectionQueryArguments = useMemo(
    //     () => ({
    //         /**
    //          * The `first` variable for our **initial** query.
    //          * If this is our first query and we were supplied a value for `visible` load that many results.
    //          * If we weren't given such a value or this is a subsequent request, only ask for one page of results.
    //          *
    //          * 'visible' is the number of results that were visible from previous requests. The initial request of
    //          * a result set will load `visible` items, then will request `first` items on each subsequent
    //          * request. This has the effect of loading the correct number of visible results when a URL
    //          * is copied during pagination. This value is only useful with cursor-based paging for the initial request.
    //          */
    //         first: (options?.useURL && parseQueryInt(searchParameters, 'visible')) || paginationRef.current.actual,
    //         /**
    //          * The `after` variable for our **initial** query.
    //          * Subsequent requests through `fetchMore` will use a valid `cursor` value here, where possible.
    //          */
    //         after: (options?.useURL && searchParameters.get('after')) || after,
    //     }),
    //     // We only need these controls for the inital request. We do not care about dependency updates.
    //     // eslint-disable-next-line react-hooks/exhaustive-deps
    //     []
    // )

    /**
     * Initial query of the hook.
     * Subsequent requests (such as further pagination) will be handled through `fetchMore`
     */
    const { data, error, loading, refetch } = useQuery<TResult, TVariables>(query, {
        variables: {
            ...variables,
            ...{
                first: pageSize,
                after: null,
                last: null,
                before: null,
            },
        },
        notifyOnNetworkStatusChange: true, // Ensures loading state is updated on `fetchMore`
        fetchPolicy: options?.fetchPolicy,
        onCompleted: options?.onCompleted,
    })

    /**
     * Map over Apollo results to provide type-compatible `GraphQLResult`s for consumers.
     * This ensures good interoperability between `FilteredConnection` and `useConnection`.
     */
    const getConnection = ({
        data,
        error,
    }: Pick<QueryResult<TResult>, 'data' | 'error'>): PaginatedConnection<TData> => {
        const result = asGraphQLResult({ data, errors: error?.graphQLErrors || [] })
        return getConnectionFromGraphQLResult(result)
    }

    const connection =useMemo(() => {
        const rawOrderedConnection = data ? getConnection({ data, error }) : undefined

        if (rawOrderedConnection!== undefined) {
            // Detect a backward pagination request and fix the order
            if (rawOrderedConnection.)
        }

    })


    // useConnectionUrl({
    //     enabled: options?.useURL,
    //     first: paginationRef.current,
    //     visibleResultCount: connection?.nodes.length,
    // })

    const nextPage = useCallback(async (): Promise<void> => {
        const cursor = connection?.pageInfo?.endCursor
        if (!cursor) {
            throw new Error('No cursor available for next page')
        }
        await refetch({
            ...variables,
            ...{ after: cursor, first: pageSize, last: null, before: null },
        })
    }, [connection?.pageInfo?.endCursor, pageSize, refetch, variables])
    const previousPage = useCallback(async (): Promise<void> => {
        const cursor = connection?.pageInfo?.endCursor
        if (!cursor) {
            throw new Error('No cursor available for next page')
        }
        await refetch({
            ...variables,
            ...{ after: null, first: null, last: pageSize, before: cursor },
        })
    }, [connection?.pageInfo?.startCursor, pageSize, refetch, variables])
    const firstPage = useCallback(async (): Promise<void> => {
        await refetch({
            ...variables,
            ...{ after: null, first: pageSize, last: null, before: null },
        })
    }, [pageSize, refetch, variables])
    const lastPage = useCallback(async (): Promise<void> => {
        await refetch({
            ...variables,
            ...{ after: null, first: null, last: pageSize, before: null },
        })
    }, [pageSize, refetch, variables])

    return {
        connection,
        loading,
        error,
        nextPage,
        previousPage,
        firstPage,
        lastPage,
    }
}
