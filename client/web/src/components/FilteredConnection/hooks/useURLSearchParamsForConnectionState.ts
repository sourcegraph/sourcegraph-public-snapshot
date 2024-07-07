import { useCallback, useMemo } from 'react'

import { useLocation, useNavigate } from 'react-router-dom'

import { QUERY_KEY } from '../constants'
import type { Filter, FilterValues } from '../FilterControl'
import { getFilterFromURL, parseQueryInt, urlSearchParamsForFilteredConnection } from '../utils'

import type { PaginatedConnectionQueryArguments } from './usePageSwitcherPagination'

/**
 * The value and a setter for the value of a GraphQL connection's params.
 */
export interface UseConnectionStateResult<TState extends PaginatedConnectionQueryArguments> {
    value: TState

    /**
     * Update the {@link UseConnectionStateResult.value} value with the given partial state value.
     * The final value is equivalent to `{...prevValue, ...updates}`.
     */
    updateValue: (updates: Partial<TState>) => void
}

/**
 * A React hook for using the URL querystring to store the state of a paginated connection.
 */
export function useUrlSearchParamsForConnectionState<
    TState extends PaginatedConnectionQueryArguments,
    K extends keyof TState & string
>(filters: Filter<K>[]): UseConnectionStateResult<TState> {
    const location = useLocation()
    const navigate = useNavigate()

    const value = useMemo<TState>(() => {
        const params = new URLSearchParams(location.search)
        return {
            ...getFilterFromURL(params, filters),
            query: params.get(QUERY_KEY) ?? '',
            first: parseQueryInt(params, 'first'),
            last: parseQueryInt(params, 'last'),
            after: params.get('after'),
            before: params.get('before'),
        } as unknown as TState
    }, [location.search, filters])

    const updateValue = useCallback(
        (update: Partial<TState>) => {
            const params = urlSearchParamsForFilteredConnection({
                pagination: {
                    first: update.first,
                    last: update.last,
                    after: update.after,
                    before: update.before,
                },
                filters,
                filterValues: { ...value, ...update } as FilterValues<K>,
                query: 'query' in update ? (update.query as string) : '',
                search: location.search,
            })
            navigate(
                {
                    search: params.toString(),
                    hash: location.hash,
                },
                {
                    replace: true,
                    state: location.state, // Preserve flash messages.
                }
            )
        },
        [filters, value, location.search, location.hash, location.state, navigate]
    )

    return { value, updateValue }
}
