import { ApolloError } from '@apollo/client'
import { useCallback, useState } from 'react'

import { useQuery } from '@sourcegraph/shared/src/graphql/graphql'

import { Connection } from '../../../components/FilteredConnection/ConnectionType'

interface PaginationConnectionQueryArguments {
    first?: number
    after?: string | null
}

interface PaginationConnectionResult<TData> {
    connection?: Connection<TData>
    loading: boolean
    error?: ApolloError
    fetchMore: () => void
}

interface UsePaginationConnectionParameters<TResult, TVariables, TData> {
    query: string
    variables: TVariables & PaginationConnectionQueryArguments
    getConnection: (result: TResult) => Connection<TData>
}

export const usePaginatedConnection = <TResult, TVariables, TData>({
    query,
    variables: _variables,
    getConnection,
}: UsePaginationConnectionParameters<TResult, TVariables, TData>): PaginationConnectionResult<TData> => {
    const [variables, setVariables] = useState<TVariables & PaginationConnectionQueryArguments>(_variables)
    const { data, loading, error, fetchMore } = useQuery<TResult, TVariables>(query, { variables })
    const connection = data ? getConnection(data) : undefined

    const fetchMoreData = useCallback(() => {
        const cursor = connection?.pageInfo?.endCursor

        if (!cursor && !variables.first) {
            throw new Error('Cannot fetch more data with no endCursor or first variable present')
        }

        return fetchMore({
            variables: {
                ...variables,
                // Use cursor paging if possible, otherwise fallback to multiplying `first`
                ...(cursor ? { after: cursor } : { first: variables.first! * 2 }),
            },
            updateQuery: (previousResult, { fetchMoreResult }) => {
                if (!fetchMoreResult) {
                    return previousResult
                }

                if (cursor) {
                    // Cursor paging so append to results
                    const previousNodes = getConnection(previousResult).nodes
                    getConnection(fetchMoreResult).nodes.unshift(...previousNodes)
                } else {
                    // Increment paging, update variable state for next fetch
                    setVariables(previous => ({
                        ...previous,
                        first: variables.first! * 2,
                    }))
                }

                return fetchMoreResult
            },
        })
    }, [connection?.pageInfo?.endCursor, fetchMore, getConnection, variables])

    return {
        connection,
        loading,
        error,
        fetchMore: fetchMoreData,
    }
}
