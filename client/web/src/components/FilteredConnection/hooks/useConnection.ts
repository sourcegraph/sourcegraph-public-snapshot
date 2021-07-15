import { QueryResult } from '@apollo/client'
import { GraphQLError } from 'graphql'
import { useCallback, useRef } from 'react'

import { GraphQLResult, useQuery } from '@sourcegraph/shared/src/graphql/graphql'
import { asGraphQLResult, hasNextPage, parseQueryInt } from '@sourcegraph/web/src/components/FilteredConnection/utils'

import { Connection, ConnectionQueryArguments } from '../ConnectionType'

import { useConnectionUrl } from './useConnectionUrl'
import { useSearchParameters } from './useSearchParameters'

interface UseConnectionResult<TData> {
    connection?: Connection<TData>
    errors?: readonly GraphQLError[]
    fetchMore: () => void
    loading: boolean
    hasNextPage: boolean
}

interface UseConnectionParameters<TResult, TVariables, TData> {
    query: string
    variables: TVariables & ConnectionQueryArguments
    getConnection: (result: GraphQLResult<TResult>) => Connection<TData>
    options?: {
        useURL?: boolean
    }
}

const DEFAULT_AFTER = null
const DEFAULT_FIRST = 20

export const useConnection = <TResult, TVariables, TData>({
    query,
    variables,
    getConnection: getConnectionFromGraphQLResult,
    options,
}: UseConnectionParameters<TResult, TVariables, TData>): UseConnectionResult<TData> => {
    const searchParameters = useSearchParameters()
    const firstRequest = useRef(true)

    /**
     * The number of results that were visible from previous requests. The initial request of
     * a result set will load `visible` items, then will request `first` items on each subsequent
     * request. This has the effect of loading the correct number of visible results when a URL
     * is copied during pagination. This value is only useful with cursor-based paging.
     */
    const visible = useRef((options?.useURL && parseQueryInt(searchParameters, 'visible')) || 0)

    const defaultFirst = useRef(variables.first || DEFAULT_FIRST)
    /**
     * The number of results that we will typically want to load in the next request (unless `visible` is used).
     * This value will typically be static for cursor-based pagination, but will be dynamic for batch-based pagination.
     */
    const first = useRef(
        (options?.useURL && parseQueryInt(searchParameters, 'first')) || variables.first || defaultFirst.current
    )
    // If this is our first query and we were supplied a value for `visible`,
    // load that many results. If we weren't given such a value or this is a
    // subsequent request, only ask for one page of results.
    const queryFirst = useRef((firstRequest.current && visible.current) || first.current)

    const after = (options?.useURL && searchParameters.get('after')) || variables.after || DEFAULT_AFTER

    const { data, error, loading, fetchMore } = useQuery<TResult, TVariables>(query, {
        variables: {
            ...variables,
            first: queryFirst.current,
            after,
        },
        onCompleted: () => (firstRequest.current = false),
    })

    /**
     * Map over Apollo results to provide type-compatible GraphQLResults for consumers.
     * This ensures good interopability between FilteredConnection and useConnection.
     */
    const getConnection = useCallback(
        ({ data, error }: Pick<QueryResult<TResult>, 'data' | 'error'>): Connection<TData> => {
            const result = asGraphQLResult({ data, errors: error?.graphQLErrors || [] })
            return getConnectionFromGraphQLResult(result)
        },
        [getConnectionFromGraphQLResult]
    )

    const connection = data ? getConnection({ data, error }) : undefined

    useConnectionUrl({
        enabled: options?.useURL,
        first: first.current,
        defaultFirst: defaultFirst.current,
        visible: connection?.nodes.length || 0,
    })

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
                    // Cursor paging so append to results, update `after` for next fetch
                    const previousNodes = getConnection({ data: previousResult }).nodes
                    getConnection({ data: fetchMoreResult }).nodes.unshift(...previousNodes)
                } else {
                    // Batch-based pagination, update `first` to fetch more results next time
                    first.current = first.current * 2
                }

                return fetchMoreResult
            },
        })
    }, [connection?.pageInfo?.endCursor, fetchMore, variables, getConnection])

    return {
        connection,
        loading,
        errors: error?.graphQLErrors,
        fetchMore: fetchMoreData,
        hasNextPage: connection ? hasNextPage(connection) : false,
    }
}
