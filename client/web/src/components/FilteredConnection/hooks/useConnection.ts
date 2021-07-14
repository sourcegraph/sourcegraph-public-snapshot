import { GraphQLError } from 'graphql'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useHistory, useLocation } from 'react-router'

import { GraphQLResult, useQuery } from '@sourcegraph/shared/src/graphql/graphql'
import {
    asGraphQLResult,
    getUrlQuery,
    hasNextPage,
    parseQueryInt,
} from '@sourcegraph/web/src/components/FilteredConnection/utils'

import { Connection, ConnectionQueryArguments } from '../ConnectionType'

import { useSearchParameters } from './useSearchParameters'

interface PaginationConnectionResult<TData> {
    connection?: Connection<TData>
    errors: readonly GraphQLError[]
    fetchMore: () => void
    loading: boolean
    hasNextPage: boolean
}

interface UsePaginationConnectionParameters<TResult, TVariables, TData> {
    query: string
    variables: TVariables & ConnectionQueryArguments
    getConnection: (result: GraphQLResult<TResult>) => Connection<TData>
    options?: {
        useURL?: boolean
    }
}

export const useConnection = <TResult, TVariables, TData>({
    query,
    variables,
    getConnection: getConnectionFromGraphQLResult,
    options,
}: UsePaginationConnectionParameters<TResult, TVariables, TData>): PaginationConnectionResult<TData> => {
    const location = useLocation()
    const history = useHistory()
    const searchParameters = useSearchParameters()

    const firstRequest = useRef(true)
    const after = useRef((options?.useURL && searchParameters.get('after')) || variables.after)
    const visible = (options?.useURL && parseQueryInt(searchParameters, 'visible')) || 0
    const first = useRef(
        (firstRequest.current && visible) ||
            (options?.useURL && parseQueryInt(searchParameters, 'first')) ||
            variables.first ||
            15
    )
    const { current: defaultFirst } = useRef(variables.first || 15)

    const queryarguments = useMemo(
        () => ({
            first: first.current,
            after: after.current,
            query: variables.query,
        }),
        [variables.query]
    )

    console.log(firstRequest.current, first.current)
    console.log('Args:', queryarguments)

    const { data, loading, error, fetchMore } = useQuery<TResult, TVariables>(query, {
        variables: {
            ...variables,
            ...queryarguments,
        },
        onCompleted: () => (firstRequest.current = false),
    })

    // Mapping over Apollo results to valid GraphQLResults for consumers
    const getConnection = useCallback(
        ({ data, errors }: { data: TResult; errors?: readonly GraphQLError[] }): Connection<TData> => {
            const result = asGraphQLResult({ data: data ?? null, errors: errors || [] })
            return getConnectionFromGraphQLResult(result)
        },
        [getConnectionFromGraphQLResult]
    )

    const errors = error?.graphQLErrors || []
    const connection = data ? getConnection({ data, errors }) : undefined

    const searchFragment = getUrlQuery({
        first: first.current,
        defaultFirst,
        visible: connection?.nodes.length || 0,
        location,
    })

    useEffect(() => {
        if (options?.useURL && searchFragment && location.search !== `?${searchFragment}`) {
            history.replace({
                search: searchFragment,
                hash: location.hash,
            })
        }
    }, [history, location.hash, location.search, options?.useURL, searchFragment])

    const fetchMoreData = useCallback(() => {
        const cursor = connection?.pageInfo?.endCursor

        return fetchMore({
            variables: {
                ...variables,
                // Use cursor paging if possible, otherwise fallback to multiplying `first`
                ...(cursor ? { after: cursor } : { first: first.current * 2 }),
            },
            updateQuery: (previousResult, { fetchMoreResult }) => {
                if (!fetchMoreResult) {
                    return previousResult
                }

                if (cursor) {
                    // Cursor paging so append to results
                    const previousNodes = getConnection({ data: previousResult }).nodes
                    getConnection({ data: fetchMoreResult }).nodes.unshift(...previousNodes)
                    console.log('Updating after')
                    after.current = cursor
                } else {
                    console.log('Updating first')
                    // Increment paging, update variable state for next fetch
                    first.current = first.current * 2
                }

                return fetchMoreResult
            },
        })
    }, [connection?.pageInfo?.endCursor, fetchMore, variables, getConnection])

    return {
        connection,
        loading,
        errors,
        fetchMore: fetchMoreData,
        hasNextPage: connection ? hasNextPage(connection) : false,
    }
}
