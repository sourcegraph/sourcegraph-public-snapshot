import { useCallback, useMemo } from 'react'

import { useLocation, useNavigate } from 'react-router-dom'

import { QUERY_KEY } from '../constants'
import type { Filter, FilterValues } from '../FilterControl'
import { getFilterFromURL, getUrlQuery } from '../utils'

import type { PaginatedConnectionQueryArguments } from './usePageSwitcherPagination'

/**
 * The value and a setter for the value of a GraphQL connection's params.
 */
export interface UseConnectionStateResult<TState extends PaginatedConnectionQueryArguments> {
    value: TState
    setValue: (value: Partial<TState>) => void
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
            first: parseNumber(params, 'first'),
            last: parseNumber(params, 'last'),
            after: params.get('after'),
            before: params.get('before'),
        } as unknown as TState
    }, [location.search, filters])

    const setValue = useCallback(
        (update: Partial<TState>) => {
            // Always clear pagination-related keys because updates of them implicitly overwrite the others.
            const search1 = new URLSearchParams(location.search)
            search1.delete('after')
            search1.delete('before')
            search1.delete('first')
            search1.delete('last')
            if (update.after) {
                search1.set('after', update.after)
            } else if (update.before) {
                search1.set('before', update.before)
            } else if (update.last) {
                search1.set('last', update.last.toString())
            } else if (update.first) {
                search1.set('first', update.first.toString())
            }

            const search = getUrlQuery({
                filters,
                filterValues: { ...value, ...update } as FilterValues<K>,
                query: 'query' in update ? (update.query as string) : '',
                search: search1.toString(),
            })
            navigate(
                {
                    search,
                    hash: location.hash,
                },
                { replace: true }
            )
        },
        [filters, value, location.search, location.hash, navigate]
    )

    return { value, setValue }
}

function parseNumber(params: URLSearchParams, name: string): number | null {
    const str = params.get(name)
    return str === null ? null : Number.parseInt(str, 10)
}
