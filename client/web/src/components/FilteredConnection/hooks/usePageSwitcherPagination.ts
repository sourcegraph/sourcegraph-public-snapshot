import { useCallback, useMemo } from 'react'

import type { ApolloError, WatchQueryFetchPolicy } from '@apollo/client'
import { useNavigate, useLocation } from 'react-router-dom'

import { type GraphQLResult, useQuery } from '@sourcegraph/http-client'

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
        startCursor?: string | null
        endCursor?: string | null
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

export interface UsePaginatedConnectionResult<TResult, TVariables, TNode> extends PaginationProps {
    data: TResult | undefined | null
    variables: TVariables
    connection?: PaginatedConnection<TNode>
    loading: boolean
    error?: ApolloError
    refetch: (variables?: TVariables) => any
    startPolling: (pollInterval: number) => void
    stopPolling: () => void
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
    // Allows to provide polling interval to useQuery
    pollInterval?: number
}

interface UsePaginatedConnectionParameters<TResult, TVariables extends PaginatedConnectionQueryArguments, TNode> {
    query: string
    variables: Omit<TVariables, 'first' | 'last' | 'before' | 'after'>
    getConnection: (result: GraphQLResult<TResult>) => PaginatedConnection<TNode> | undefined
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
export const usePageSwitcherPagination = <TResult, TVariables extends PaginatedConnectionQueryArguments, TNode>({
    query,
    variables,
    getConnection,
    options,
}: UsePaginatedConnectionParameters<TResult, TVariables, TNode>): UsePaginatedConnectionResult<
    TResult,
    TVariables,
    TNode
> => {
    const pageSize = options?.pageSize ?? DEFAULT_PAGE_SIZE
    const [initialPaginationArgs, setPaginationArgs] = useSyncPaginationArgsWithUrl(!!options?.useURL, pageSize)

    // TODO(philipp-spiess): Find out why Omit<TVariables, "first" | ...> & { first: number, ... }
    // does not work here and get rid of the any cast.

    const queryVariables: TVariables = {
        ...variables,
        ...initialPaginationArgs,
    } as any

    const {
        data: currentData,
        previousData,
        error,
        loading,
        refetch,
        startPolling: startPollingFunction,
        stopPolling: stopPollingFunction,
    } = useQuery<TResult, TVariables>(query, {
        variables: queryVariables,
        fetchPolicy: options?.fetchPolicy,
        onCompleted: options?.onCompleted,
        pollInterval: options?.pollInterval,
    })

    const data = currentData ?? previousData

    const connection = useMemo(() => {
        if (!data) {
            return undefined
        }
        const result = asGraphQLResult({ data, errors: error?.graphQLErrors || [] })
        return getConnection(result)
    }, [data, error, getConnection])

    const updatePagination = useCallback(
        async (nextPageArgs: PaginatedConnectionQueryArguments): Promise<void> => {
            setPaginationArgs(nextPageArgs)
            await refetch(nextPageArgs as Partial<TVariables>)
        },
        [refetch, setPaginationArgs]
    )

    const goToNextPage = useCallback(async (): Promise<void> => {
        const cursor = connection?.pageInfo?.endCursor
        if (!cursor) {
            throw new Error('No cursor available for next page')
        }
        await updatePagination({ after: cursor, first: pageSize, last: null, before: null })
    }, [connection?.pageInfo?.endCursor, updatePagination, pageSize])

    const goToPreviousPage = useCallback(async (): Promise<void> => {
        const cursor = connection?.pageInfo?.startCursor
        if (!cursor) {
            throw new Error('No cursor available for previous page')
        }
        await updatePagination({ after: null, first: null, last: pageSize, before: cursor })
    }, [connection?.pageInfo?.startCursor, updatePagination, pageSize])

    const goToFirstPage = useCallback(async (): Promise<void> => {
        await updatePagination({ after: null, first: pageSize, last: null, before: null })
    }, [updatePagination, pageSize])

    const goToLastPage = useCallback(async (): Promise<void> => {
        await updatePagination({ after: null, first: null, last: pageSize, before: null })
    }, [updatePagination, pageSize])

    const startPolling = useCallback(
        (pollInterval: number): void => {
            startPollingFunction(pollInterval)
        },
        [startPollingFunction]
    )

    const stopPolling = useCallback((): void => {
        stopPollingFunction()
    }, [stopPollingFunction])

    return {
        data,
        variables: queryVariables,
        connection,
        loading,
        error,
        refetch,
        hasNextPage: connection?.pageInfo?.hasNextPage ?? null,
        hasPreviousPage: connection?.pageInfo?.hasPreviousPage ?? null,
        goToNextPage,
        goToPreviousPage,
        goToFirstPage,
        goToLastPage,
        startPolling,
        stopPolling,
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

const useSyncPaginationArgsWithUrl = (
    enabled: boolean,
    pageSize: number
): [
    initialPaginationArgs: PaginatedConnectionQueryArguments,
    setPaginationArgs: (args: PaginatedConnectionQueryArguments) => void
] => {
    const location = useLocation()
    const navigate = useNavigate()

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
                navigate({ search }, { replace: true })
            }
        },
        [enabled, navigate]
    )
    return [initialPaginationArgs, setPaginationArgs]
}
