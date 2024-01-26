import { useCallback, useMemo } from 'react'

import { useLocation, useNavigate } from 'react-router-dom'

export interface URLStateOptions<State, SerializedState> {
    urlKey: string
    deserializer: (value: SerializedState | null) => State
    serializer: (state: State) => string | null
    replace?: boolean
}

export type UpdatedSearchQuery = string

export interface UpdateSearchQueryOptions {
    replace: boolean
}

export type SetStateResult<State> = [
    state: State,
    dispatch: (state: State, options?: UpdateSearchQueryOptions) => UpdatedSearchQuery,
    serialize: (state: State) => UpdatedSearchQuery
]

/**
 * React hook analog standard react useState hook but with synced value with URL
 * through URL query parameter.
 */
export function useSyncedWithURLState<State, SerializedState>(
    options: URLStateOptions<State, SerializedState>
): SetStateResult<State> {
    const { urlKey, serializer, deserializer, replace = true } = options
    const navigate = useNavigate()
    const { search } = useLocation()

    const urlSearchParameters = useMemo(() => new URLSearchParams(search), [search])
    const queryParameter = useMemo(
        () => deserializer(urlSearchParameters.get(urlKey) as unknown as SerializedState | null),
        [urlSearchParameters, urlKey, deserializer]
    )

    const setNextState = useCallback(
        (nextState: State, options?: UpdateSearchQueryOptions) => {
            const serializedValue = serializer(nextState)

            if (serializedValue === null) {
                urlSearchParameters.delete(urlKey)
            } else {
                urlSearchParameters.set(urlKey, serializedValue)
            }

            const search = `?${urlSearchParameters.toString()}`
            navigate({ search }, { replace: options?.replace ?? replace })

            return search
        },
        [navigate, serializer, urlKey, urlSearchParameters, replace]
    )

    const serializeURL = useCallback(
        (nextState: State) => {
            const serializedValue = serializer(nextState)

            if (serializedValue === null) {
                urlSearchParameters.delete(urlKey)
            } else {
                urlSearchParameters.set(urlKey, serializedValue)
            }

            return `?${urlSearchParameters.toString()}`
        },
        [serializer, urlKey, urlSearchParameters]
    )

    return [queryParameter, setNextState, serializeURL]
}
