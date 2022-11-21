import { useCallback, useMemo } from 'react'

import { ApolloError, QueryResult, WatchQueryFetchPolicy } from '@apollo/client'
import { useHistory, useLocation } from 'react-router'

import { GraphQLResult, useQuery } from '@sourcegraph/http-client'

import { asGraphQLResult } from '../utils'

export interface PaginatedConnectionQueryArguments {
    first?: number | null
    last?: number | null
    after?: string | null
    before?: string | null
}

export interface PaginatedConnection<N> {
    nodes: N[]
    totalCount?: number
    pageInfo?: {
        hasNextPage: boolean
        hasPreviousPage: boolean
        startCursor?: string
        endCursor?: string
    }
    error?: string | null
}

export interface PaginationProps {
    hasNextPage: null | boolean
    hasPreviousPage: null | boolean
    goToNextPage: () => Promise<void>
    goToPreviousPage: () => Promise<void>
    goToFirstPage: () => Promise<void>
    goToLastPage: () => Promise<void>
}

export interface UsePaginatedConnectionResult<TData> extends PaginationProps {
    connection?: PaginatedConnection<TData>
    loading: boolean
    error?: ApolloError
}

interface UsePaginatedConnectionConfig<TResult> {
    // The number of items per page, defaults to 20
    pageSize?: number
    // Set if query variables should be updated in and derived from the URL
    useURL?: boolean
    // Allows modifying how the query interacts with the Apollo cache
    fetchPolicy?: WatchQueryFetchPolicy
    // Allows running an optional callback on any successful request
    onCompleted?: (data: TResult) => void
}

interface UsePaginatedConnectionParameters<TResult, TVariables extends PaginatedConnectionQueryArguments, TData> {
    query: string
    variables: Omit<TVariables, 'first' | 'last' | 'before' | 'after'>
    getConnection: (result: GraphQLResult<TResult>) => PaginatedConnection<TData>
    options?: UsePaginatedConnectionConfig<TResult>
}

const DEFAULT_PAGE_SIZE = 20

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
    const pageSize = options?.pageSize ?? DEFAULT_PAGE_SIZE
    const [initialPaginationArgs, setPaginationArgs] = useSyncUrl(!!options?.useURL, pageSize)

    // TODO(philipp-spiess): Find out why Omit<TVariables, "first" | ...> & { first: number, ... } does not work
    // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
    const initialQueryVariables: TVariables = useMemo(
        () => ({ ...variables, ...initialPaginationArgs } as any),
        // The variables object can be different every time the hook is called but we only need to react
        // to a change in the individual variable values.
        // eslint-disable-next-line react-hooks/exhaustive-deps
        [pageSize, initialPaginationArgs, ...Object.values(variables)]
    )

    const { data, error, loading, refetch } = useQuery<TResult, TVariables>(query, {
        variables: initialQueryVariables,
        notifyOnNetworkStatusChange: true, // Ensures loading state is updated on `refetch`
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

    const connection = data ? getConnection({ data, error }) : undefined

    const goToNextPage = useCallback(async (): Promise<void> => {
        const cursor = connection?.pageInfo?.endCursor
        if (!cursor) {
            throw new Error('No cursor available for next page')
        }
        const nextPageArgs = { after: cursor, first: pageSize, last: null, before: null }
        setPaginationArgs(nextPageArgs)
        await refetch({
            ...initialQueryVariables,
            ...nextPageArgs,
        })
    }, [connection?.pageInfo?.endCursor, pageSize, setPaginationArgs, refetch, initialQueryVariables])

    const goToPreviousPage = useCallback(async (): Promise<void> => {
        const cursor = connection?.pageInfo?.startCursor
        if (!cursor) {
            throw new Error('No cursor available for next page')
        }
        const previousPageArgs = {
            after: null,
            first: null,
            last: pageSize,
            before: cursor,
        }
        setPaginationArgs(previousPageArgs)
        await refetch({
            ...initialQueryVariables,
            ...previousPageArgs,
        })
    }, [connection?.pageInfo?.startCursor, pageSize, setPaginationArgs, refetch, initialQueryVariables])

    const goToFirstPage = useCallback(async (): Promise<void> => {
        const firstPageArgs = {
            after: null,
            first: pageSize,
            last: null,
            before: null,
        }
        setPaginationArgs(firstPageArgs)
        await refetch({
            ...initialQueryVariables,
            ...firstPageArgs,
        })
    }, [pageSize, setPaginationArgs, refetch, initialQueryVariables])

    const goToLastPage = useCallback(async (): Promise<void> => {
        const lastPageArgs = {
            after: null,
            first: null,
            last: pageSize,
            before: null,
        }
        setPaginationArgs(lastPageArgs)
        await refetch({
            ...initialQueryVariables,
            ...lastPageArgs,
        })
    }, [pageSize, setPaginationArgs, refetch, initialQueryVariables])

    return {
        connection,
        loading,
        error,
        hasNextPage: connection?.pageInfo?.hasNextPage ?? null,
        hasPreviousPage: connection?.pageInfo?.hasPreviousPage ?? null,
        goToNextPage,
        goToPreviousPage,
        goToFirstPage,
        goToLastPage,
    }
}

// TODO(philipp-spiess): We should make these callbacks overridable by the
// consumer of this API to allow for serialization of other query parameters in
// the URL (e.g. filters).
//
// We also need to change this if we ever want to allow users to change the page
// size and want to make it persist in the URL.
const getPaginationArgsFromSearch = (search: string, pageSize: number): PaginatedConnectionQueryArguments => {
    const searchParameters = new URLSearchParams(search)

    if (searchParameters.has('after')) {
        return { first: pageSize, last: null, after: searchParameters.get('after'), before: null }
    }
    if (searchParameters.has('before')) {
        return { first: null, last: pageSize, after: null, before: searchParameters.get('before') }
    }
    // Special case for handling the last page.
    if (searchParameters.has('last')) {
        return { first: null, last: pageSize, after: null, before: null }
    }
    return { first: pageSize, last: null, after: null, before: null }
}
const getSearchFromPaginationArgs = (paginationArgs: PaginatedConnectionQueryArguments): string => {
    const searchParameters = new URLSearchParams()
    if (paginationArgs.after) {
        searchParameters.set('after', paginationArgs.after)
        return searchParameters.toString()
    }
    if (paginationArgs.before) {
        searchParameters.set('before', paginationArgs.before)
        return searchParameters.toString()
    }
    if (paginationArgs.last) {
        searchParameters.set('last', paginationArgs.last.toString())
        return searchParameters.toString()
    }
    return ''
}

export const useSyncUrl = (
    enabled: boolean,
    pageSize: number
): [
    initialPaginationArgs: PaginatedConnectionQueryArguments,
    setPaginationArgs: (args: PaginatedConnectionQueryArguments) => void
] => {
    const location = useLocation()
    const history = useHistory()

    const initialPaginationArgs = useMemo(() => {
        if (enabled) {
            return getPaginationArgsFromSearch(location.search, pageSize)
        }
        return { first: pageSize, last: null, after: null, before: null }
        // We deliberately ignore changes to the URL after the first render
        // since we assume that these are caused by this hook.
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [enabled])

    const setPaginationArgs = useCallback(
        (paginationArgs: PaginatedConnectionQueryArguments): void => {
            if (enabled) {
                const search = getSearchFromPaginationArgs(paginationArgs)
                history.replace({ search })
            }
        },
        [enabled, history]
    )
    return [initialPaginationArgs, setPaginationArgs]
}
