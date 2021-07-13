import { GraphQLError } from 'graphql'
import { useCallback } from 'react'

import { GraphQLResult, useQuery } from '@sourcegraph/shared/src/graphql/graphql'
import { asGraphQLResult, hasNextPage } from '@sourcegraph/web/src/components/FilteredConnection/utils'
import { useDebounce } from '@sourcegraph/wildcard'

import { Connection, ConnectionQueryArguments } from '../ConnectionType'

interface PaginationConnectionResult<TData> {
    connection?: Connection<TData>
    errors: readonly GraphQLError[]
    fetchMore: (args: ConnectionQueryArguments) => void
    loading: boolean
    hasNextPage: boolean
}

interface UsePaginationConnectionParameters<TResult, TVariables, TData> {
    query: string
    variables: TVariables & ConnectionQueryArguments
    getConnection: (result: GraphQLResult<TResult>) => Connection<TData>
}

export const useConnection = <TResult, TVariables, TData>({
    query,
    variables,
    getConnection: getConnectionFromGraphQLResult,
}: UsePaginationConnectionParameters<TResult, TVariables, TData>): PaginationConnectionResult<TData> => {
    const debouncedQuery = useDebounce(variables.query, 200)
    const { data, loading, error, fetchMore } = useQuery<TResult, TVariables>(query, {
        variables: {
            ...variables,
            query: debouncedQuery,
        },
    })
    const errors = error?.graphQLErrors || []

    // Mapping over Apollo results to valid GraphQLResults for consumers
    const getConnection = useCallback(
        ({ data, errors }: { data: TResult; errors?: readonly GraphQLError[] }): Connection<TData> => {
            const result = asGraphQLResult({ data: data ?? null, errors: errors || [] })
            return getConnectionFromGraphQLResult(result)
        },
        [getConnectionFromGraphQLResult]
    )

    const connection = data ? getConnection({ data, errors }) : undefined

    const fetchMoreData = useCallback(
        (args: ConnectionQueryArguments) =>
            fetchMore({
                variables: {
                    ...variables,
                    ...args,
                },
                updateQuery: (previousResult, { fetchMoreResult }) => {
                    if (!fetchMoreResult) {
                        return previousResult
                    }

                    if (args.after) {
                        // Cursor paging so append to results
                        const previousNodes = getConnection({ data: previousResult }).nodes
                        getConnection({ data: fetchMoreResult }).nodes.unshift(...previousNodes)
                    }

                    return fetchMoreResult
                },
            }),
        [fetchMore, variables, getConnection]
    )

    return {
        connection,
        loading,
        errors: error?.graphQLErrors || [],
        fetchMore: fetchMoreData,
        hasNextPage: connection ? hasNextPage(connection) : false,
    }
}
