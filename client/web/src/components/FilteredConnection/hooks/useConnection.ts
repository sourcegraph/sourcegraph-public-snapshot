import { GraphQLError } from 'graphql'
import { useCallback, useEffect, useState } from 'react'

import { GraphQLResult, useQuery } from '@sourcegraph/shared/src/graphql/graphql'
import { asGraphQLResult, hasNextPage, parseQueryInt } from '@sourcegraph/web/src/components/FilteredConnection/utils'

import { Connection } from '../ConnectionType'

import { useSearchParameters } from './useSearchParameters'
import { useUrlQuery } from './useUrlQuery'

interface PaginationConnectionQueryArguments {
    first?: number
    after?: string | null
    query?: string
}

interface PaginationConnectionResult<TData> {
    connection?: Connection<TData>
    errors: readonly GraphQLError[]
    fetchMore: () => void
    loading: boolean
    hasNextPage: boolean
}

interface UsePaginationConnectionOptions {
    useURLQuery?: boolean
}
interface UsePaginationConnectionParameters<TResult, TVariables, TData> {
    query: string
    variables: TVariables & PaginationConnectionQueryArguments
    getConnection: (result: GraphQLResult<TResult>) => Connection<TData>
    options?: UsePaginationConnectionOptions
}

export const usePaginatedConnection = <TResult, TVariables, TData>({
    query,
    variables: uncontrolledVariables,
    getConnection: getConnectionFromGraphQLResult,
    options,
}: UsePaginationConnectionParameters<TResult, TVariables, TData>): PaginationConnectionResult<TData> => {
    const searchParameters = useSearchParameters()

    const [controlledVariables, setControlledVariables] = useState<TVariables & PaginationConnectionQueryArguments>({
        ...uncontrolledVariables,
        first: (options?.useURLQuery && parseQueryInt(searchParameters, 'first')) || uncontrolledVariables.first,
    })

    const { data, loading, error, fetchMore } = useQuery<TResult, TVariables>(query, { variables: controlledVariables })
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

    // Support allowing consumers to control the query variable
    useEffect(() => {
        if (uncontrolledVariables.query !== controlledVariables.query) {
            setControlledVariables(previous => ({
                ...previous,
                query: uncontrolledVariables.query,
            }))
        }
    }, [uncontrolledVariables.query, controlledVariables.query])

    useUrlQuery({
        enabled: options?.useURLQuery,
        first: uncontrolledVariables.first,
        defaultFirst: controlledVariables.first,
        query: controlledVariables.query,
    })

    const fetchMoreData = useCallback(() => {
        const cursor = connection?.pageInfo?.endCursor

        if (!cursor && !controlledVariables.first) {
            throw new Error('Cannot fetch more data with no endCursor or first variable present')
        }

        return fetchMore({
            variables: {
                ...controlledVariables,
                // Use cursor paging if possible, otherwise fallback to multiplying `first`
                ...(cursor ? { after: cursor } : { first: controlledVariables.first! * 2 }),
            },
            updateQuery: (previousResult, { fetchMoreResult }) => {
                if (!fetchMoreResult) {
                    return previousResult
                }

                if (cursor) {
                    // Cursor paging so append to results
                    const previousNodes = getConnection({ data: previousResult }).nodes
                    getConnection({ data: fetchMoreResult }).nodes.unshift(...previousNodes)
                } else {
                    // Increment paging, update variable state for next fetch
                    setControlledVariables(previous => ({
                        ...previous,
                        first: controlledVariables.first! * 2,
                    }))
                }

                return fetchMoreResult
            },
        })
    }, [connection?.pageInfo?.endCursor, fetchMore, getConnection, controlledVariables])

    return {
        connection,
        loading,
        errors: error?.graphQLErrors || [],
        fetchMore: fetchMoreData,
        hasNextPage: connection ? hasNextPage(connection) : false,
    }
}
