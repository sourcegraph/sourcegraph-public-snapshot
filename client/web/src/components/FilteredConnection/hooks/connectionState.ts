import { useCallback, useMemo, useRef, useState } from 'react'

import { useLocation, useNavigate } from 'react-router-dom'

import { QUERY_KEY } from '../constants'
import type { Filter } from '../FilterControl'
import { getFilterFromURL, parseQueryInt, urlSearchParamsForFilteredConnection } from '../utils'

import type { PaginatedConnectionQueryArguments } from './usePageSwitcherPagination'

/**
 * The value and a setter for the value of a GraphQL connection's params.
 */
export type UseConnectionStateResult<TState extends PaginatedConnectionQueryArguments> = [
    connectionState: TState,

    /**
     * Set the {@link UseConnectionStateResult.connectionState} value in a callback that receives the current
     * value as an argument. Usually callers to {@link UseConnectionStateResult.setConnectionState} will
     * want to merge values (like `updateValue(prev => ({...prev, ...newValue}))`).
     */
    setConnectionState: (valueFunc: (current: TState) => TState) => void
]

/**
 * A React hook for using the URL querystring to store the state of a paginated connection,
 * including both pagination parameters (such as `first` and `after`) and other custom filter
 * parameters.
 */
export function useUrlSearchParamsForConnectionState<TFilterKeys extends string>(
    filters?: Filter<TFilterKeys>[]
): UseConnectionStateResult<
    Partial<Record<TFilterKeys, string>> & { query?: string } & PaginatedConnectionQueryArguments
> {
    type TState = Partial<Record<TFilterKeys, string>> & { query?: string } & PaginatedConnectionQueryArguments

    const location = useLocation()
    const navigate = useNavigate()

    // Use a ref that is set on each render so that our `setValue` callback can access the latest
    // value without having the value as one of its deps, which can cause render cycles. Note that
    // this is how `useState` works as well (the setter's function value does not change when the
    // value changes).
    const value = useRef<TState>()
    value.current = useMemo<TState>((): TState => {
        const params = new URLSearchParams(location.search)

        const pgParams: PaginatedConnectionQueryArguments = {
            first: parseQueryInt(params, 'first'),
            last: parseQueryInt(params, 'last'),
            after: params.get('after') ?? undefined,
            before: params.get('before') ?? undefined,
        }
        const filterParams: Partial<Record<TFilterKeys, string>> = filters
            ? getFilterFromURL<TFilterKeys>(params, filters)
            : {}
        return {
            query: params.get(QUERY_KEY) ?? '',
            ...pgParams,
            ...filterParams,
        }
    }, [location.search, filters])

    const locationRef = useRef<typeof location>(location)
    locationRef.current = location
    const setValue = useCallback(
        (valueFunc: (current: TState) => TState) => {
            const location = locationRef.current
            const newValue = valueFunc(value.current!)
            const params = urlSearchParamsForFilteredConnection({
                pagination: {
                    first: newValue.first,
                    last: newValue.last,
                    after: newValue.after,
                    before: newValue.before,
                },
                filters,
                filterValues: newValue,
                query: 'query' in newValue ? newValue.query : '',
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
        [filters, navigate]
    )

    return [value.current, setValue]
}

/**
 * A React hook for using the provided connection state (usually from
 * {@link useUrlSearchParamsForConnectionState}) if defined, or otherwise falling back to an
 * in-memory connection state implementation that does not read from and write to the URL.
 */
export function useConnectionStateOrMemoryFallback<
    TFilterKeys extends string,
    TState extends PaginatedConnectionQueryArguments = Record<TFilterKeys | 'query', string> &
        PaginatedConnectionQueryArguments
>(state: UseConnectionStateResult<TState> | undefined): UseConnectionStateResult<TState> {
    const memoryState = useState<TState>({} as TState)
    return state ?? memoryState
}

/**
 * A React hook that wraps the provided {@link UseConnectionStateResult} so that `?first` and
 * `?last` URL parameters are omitted if they are equal to the default page size. This makes the
 * URLs look nicer.
 */
export function useConnectionStateWithImplicitPageSize<
    TFilterKeys extends string,
    TState extends PaginatedConnectionQueryArguments = Record<TFilterKeys | 'query', string> &
        PaginatedConnectionQueryArguments
>(state: UseConnectionStateResult<TState>, pageSize: number): UseConnectionStateResult<TState> {
    const [value, setValue] = state

    // The resolved value has explicit `first` and `last`.
    const resolvedValue = useMemo<TState>(
        () => ({
            ...value,
            first: value.first ?? (!value.before && !value.last ? pageSize : null),
            last: value.last ?? (value.before && !value.after && !value.first ? pageSize : null),
        }),
        [value, pageSize]
    )

    // The setter removes `first` and `last` if they are equal to the default page size and
    // otherwise implicit.
    const setValueWithImplicits = useCallback(
        (valueFunc: (current: TState) => TState) => {
            setValue(prev => {
                const newValue = valueFunc(prev)
                return {
                    ...newValue,
                    first:
                        newValue.first === pageSize && !newValue.before && !newValue.last ? undefined : newValue.first,
                    last:
                        newValue.last === pageSize && newValue.before && !newValue.after && !newValue.first
                            ? undefined
                            : newValue.last,
                }
            })
        },
        [pageSize, setValue]
    )

    return [resolvedValue, setValueWithImplicits]
}
