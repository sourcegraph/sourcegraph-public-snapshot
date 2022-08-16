import { useCallback, useMemo } from 'react'

import { useHistory, useLocation } from 'react-router'

import { AggregationModes, AggregationUIMode } from './types'

interface URLStateOptions<State> {
    urlKey: string
    deserializer: (value: string | null) => State
    serializer: (state: State) => string
}

type SetStateResult<State> = [state: State, dispatch: (state: State) => void]

/**
 * React hook analog standard react useState hook but with synced value with URL
 * through URL query parameter.
 */
function useSyncedWithURLState<State>(options: URLStateOptions<State>): SetStateResult<State> {
    const { urlKey, serializer, deserializer } = options
    const history = useHistory()
    const { search } = useLocation()

    const urlSearchParameters = useMemo(() => new URLSearchParams(search), [search])
    const queryParameter = useMemo(() => deserializer(urlSearchParameters.get(urlKey)), [
        urlSearchParameters,
        urlKey,
        deserializer,
    ])

    const setNextState = useCallback(
        (nextState: State) => {
            urlSearchParameters.set(urlKey, serializer(nextState))

            history.replace({ search: `?${urlSearchParameters.toString()}` })
        },
        [history, serializer, urlKey, urlSearchParameters]
    )

    return [queryParameter, setNextState]
}

const AGGREGATION_MODE_URL_KEY = 'groupBy'

const aggregationModeSerializer = (mode: AggregationModes): string => mode.toString()

const aggregationModeDeserializer = (serializedValue: string | null): AggregationModes => {
    switch (serializedValue) {
        case 'repo':
            return AggregationModes.Repository
        case 'file':
            return AggregationModes.FilePath
        case 'author':
            return AggregationModes.Author
        case 'captureGroup':
            return AggregationModes.CaptureGroups

        default:
            return AggregationModes.Repository
    }
}

/**
 * Shared state hook for syncing aggregation type state between different UI trough
 * ULR query param {@link AGGREGATION_MODE_URL_KEY}
 */
export const useAggregationSearchMode = (): SetStateResult<AggregationModes> => {
    const [aggregationMode, setAggregationMode] = useSyncedWithURLState({
        urlKey: AGGREGATION_MODE_URL_KEY,
        serializer: aggregationModeSerializer,
        deserializer: aggregationModeDeserializer,
    })

    return [aggregationMode, setAggregationMode]
}

const AGGREGATION_UI_MODE_URL_KEY = 'groupByUI'

const aggregationUIModeSerializer = (uiMode: AggregationUIMode): string => uiMode.toString()

const aggregationUIModeDeserializer = (serializedValue: string | null): AggregationUIMode => {
    switch (serializedValue) {
        case 'searchPage':
            return AggregationUIMode.SearchPage

        default:
            return AggregationUIMode.Sidebar
    }
}

/**
 * Shared state hook for syncing aggregation UI mode state between different UI trough
 * ULR query param {@link AGGREGATION_UI_MODE_URL_KEY}
 */
export const useAggregationUIMode = (): SetStateResult<AggregationUIMode> => {
    const [aggregationMode, setAggregationMode] = useSyncedWithURLState({
        urlKey: AGGREGATION_UI_MODE_URL_KEY,
        serializer: aggregationUIModeSerializer,
        deserializer: aggregationUIModeDeserializer,
    })

    return [aggregationMode, setAggregationMode]
}
