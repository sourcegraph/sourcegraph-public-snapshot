import { useCallback, useMemo } from 'react'

import { useLocation, useNavigate } from 'react-router-dom'

import { QUERY_KEY } from '../constants'
import type { Filter, FilterValues } from '../FilterControl'
import { getFilterFromURL, parseQueryInt, urlSearchParamsForFilteredConnection } from '../utils'

import { DEFAULT_PAGE_SIZE, type PaginatedConnectionQueryArguments } from './usePageSwitcherPagination'

/**
 * The value and a setter for the value of a GraphQL connection's params.
 */
export interface UseConnectionStateResult<TState extends PaginatedConnectionQueryArguments> {
    value: TState

    /**
     * Set the {@link UseConnectionStateResult.value} value in a callback that receives the current
     * value as an argument. Usually callers to {@link UseConnectionStateResult.setValue} will
     * want to merge values (like `updateValue(prev => ({...prev, ...newValue}))`).
     */
    setValue: (valueFunc: (current: TState) => TState) => void
}

/**
 * A React hook for using the URL querystring to store the state of a paginated connection.
 */
export function useUrlSearchParamsForConnectionState<
    TState extends PaginatedConnectionQueryArguments,
    K extends keyof TState & string
>(filters?: Filter<K>[], pageSize?: number): UseConnectionStateResult<TState> {
    const location = useLocation()
    const navigate = useNavigate()

    pageSize = pageSize ?? DEFAULT_PAGE_SIZE

    const value = useMemo<TState>(() => {
        const params = new URLSearchParams(location.search)

        // The `first` and `last` params are omitted from the URL if they equal the default pageSize
        // to make the URL cleaner, so we need to resolve the actual value.
        const first =
            parseQueryInt(params, 'first') ??
            (params.has('after') || (!params.has('before') && !params.has('last')) ? pageSize : null)
        const last =
            parseQueryInt(params, 'last') ??
            (params.has('before') && !params.has('after') && !params.has('last') ? pageSize : null)

        return {
            ...(filters ? getFilterFromURL(params, filters) : undefined),
            query: params.get(QUERY_KEY) ?? '',
            first,
            last,
            after: params.get('after'),
            before: params.get('before'),
        } as unknown as TState
    }, [location.search, pageSize, filters])

    const setValue = useCallback(
        (valueFunc: (current: TState) => TState) => {
            const newValue = valueFunc(value)
            const params = urlSearchParamsForFilteredConnection({
                pagination: {
                    first: newValue.first,
                    last: newValue.last,
                    after: newValue.after,
                    before: newValue.before,
                },
                pageSize,
                filters,
                filterValues: newValue as FilterValues<K>,
                query: 'query' in newValue ? (newValue.query as string) : '',
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
        [filters, pageSize, value, location.search, location.hash, location.state, navigate]
    )

    return { value, setValue }
}
