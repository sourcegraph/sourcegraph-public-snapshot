import { useCallback, useMemo } from 'react'

import type { ApolloError, WatchQueryFetchPolicy } from '@apollo/client'

import { useQuery, type GraphQLResult } from '@sourcegraph/http-client'

import { asGraphQLResult } from '../utils'

import {
    useConnectionStateOrMemoryFallback,
    useConnectionStateWithImplicitPageSize,
    type UseConnectionStateResult,
} from './connectionState'

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
    /** The number of items per page. Defaults to 20. */
    pageSize?: number

    /** Allows modifying how the query interacts with the Apollo cache. */
    fetchPolicy?: WatchQueryFetchPolicy

    /** Allows running an optional callback on any successful request. */
    onCompleted?: (data: TResult) => void

    /** Allows to provide polling interval to useQuery. */
    pollInterval?: number
}

export type PaginationKeys = 'first' | 'last' | 'before' | 'after'

interface UsePaginatedConnectionParameters<
    TResult,
    TVariables extends PaginatedConnectionQueryArguments,
    TNode,
    TState extends PaginatedConnectionQueryArguments
> {
    query: string
    variables: Omit<TVariables, PaginationKeys>
    getConnection: (result: GraphQLResult<TResult>) => PaginatedConnection<TNode> | undefined
    options?: UsePaginatedConnectionConfig<TResult>

    /**
     * The value and setter for the state parameters (such as `first`, `after`, `before`, and
     * filters).
     */
    state?: UseConnectionStateResult<TState>
}

export const DEFAULT_PAGE_SIZE = 20

/**
 * Request a GraphQL connection query and handle pagination options.
 * Valid queries should follow the connection specification at https://relay.dev/graphql/connections.htm
 * @param query The GraphQL connection query
 * @param variables The GraphQL connection variables
 * @param getConnection A function that filters and returns the relevant data from the connection response.
 * @param options Additional configuration options
 */
export const usePageSwitcherPagination = <
    TResult,
    TVariables extends PaginatedConnectionQueryArguments,
    TNode,
    TState extends PaginatedConnectionQueryArguments = PaginatedConnectionQueryArguments &
        Partial<Record<string | 'query', string>>
>({
    query,
    variables,
    getConnection,
    options,
    state,
}: UsePaginatedConnectionParameters<TResult, TVariables, TNode, TState>): UsePaginatedConnectionResult<
    TResult,
    TVariables,
    TNode
> => {
    const pageSize = options?.pageSize ?? DEFAULT_PAGE_SIZE
    const [connectionState, setConnectionState] = useConnectionStateWithImplicitPageSize(
        useConnectionStateOrMemoryFallback(state),
        pageSize
    )

    const queryVariables = {
        ...variables,

        // Pagination
        first: connectionState.first ?? null,
        last: connectionState.last ?? null,
        after: connectionState.after ?? null,
        before: connectionState.before ?? null,
    } as TVariables

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
            setConnectionState(prev => ({ ...prev, ...nextPageArgs }))
            await refetch(nextPageArgs as Partial<TVariables>)
        },
        [refetch, setConnectionState]
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
