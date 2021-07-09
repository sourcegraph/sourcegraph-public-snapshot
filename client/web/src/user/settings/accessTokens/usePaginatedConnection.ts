import { ApolloError } from '@apollo/client'
import { useCallback, useEffect, useState } from 'react'
import { useLocation } from 'react-router'

import { useQuery } from '@sourcegraph/shared/src/graphql/graphql'
import { parseQueryInt } from '@sourcegraph/web/src/components/FilteredConnection/utils'

import { Connection } from '../../../components/FilteredConnection/ConnectionType'

import { useUrlQuery } from './useUrlQuery'

interface PaginationConnectionQueryArguments {
    first?: number
    after?: string | null
    query?: string
}

interface PaginationConnectionResult<TData> {
    connection?: Connection<TData>
    loading: boolean
    error?: ApolloError
    fetchMore: () => void
}

interface UsePaginationConnectionOptions {
    useURLQuery?: boolean
}

interface UsePaginationConnectionParameters<TResult, TVariables, TData> {
    query: string
    variables: TVariables & PaginationConnectionQueryArguments
    getConnection: (result: TResult) => Connection<TData>
    options?: UsePaginationConnectionOptions
}

const useSearchParameters = (): URLSearchParams => {
    const location = useLocation()
    return new URLSearchParams(location.search)
}

export const usePaginatedConnection = <TResult, TVariables, TData>({
    query,
    variables: _variables,
    getConnection,
    options,
}: UsePaginationConnectionParameters<TResult, TVariables, TData>): PaginationConnectionResult<TData> => {
    const searchParameters = useSearchParameters()

    const [variables, setVariables] = useState<TVariables & PaginationConnectionQueryArguments>({
        ..._variables,
        first: (options?.useURLQuery && parseQueryInt(searchParameters, 'first')) || _variables.first,
    })

    const { data, loading, error, fetchMore } = useQuery<TResult, TVariables>(query, { variables })
    const connection = data ? getConnection(data) : undefined

    // Support allowing consumers to control the query variable
    useEffect(() => {
        if (_variables.query !== variables.query) {
            setVariables(previous => ({
                ...previous,
                query: _variables.query,
            }))
        }
    }, [_variables.query, variables.query])

    useUrlQuery({
        enabled: options?.useURLQuery,
        first: variables.first,
        initialFirst: _variables.first,
        query: variables.query,
    })

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
