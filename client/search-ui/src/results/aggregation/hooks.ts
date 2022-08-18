import { useCallback, useMemo } from 'react'

import { useHistory, useLocation } from 'react-router'

import { AGGREGATION_MODE_URL_KEY, AGGREGATION_UI_MODE_URL_KEY } from './constants'
import { AggregationMode, AggregationUIMode } from './types'

interface URLStateOptions<State, SerializedState> {
    urlKey: string
    deserializer: (value: SerializedState | null) => State
    serializer: (state: State) => string
}

type SetStateResult<State> = [state: State, dispatch: (state: State) => void]

/**
 * React hook analog standard react useState hook but with synced value with URL
 * through URL query parameter.
 */
function useSyncedWithURLState<State, SerializedState>(
    options: URLStateOptions<State, SerializedState>
): SetStateResult<State> {
    const { urlKey, serializer, deserializer } = options
    const history = useHistory()
    const { search } = useLocation()

    const urlSearchParameters = useMemo(() => new URLSearchParams(search), [search])
    const queryParameter = useMemo(
        () => deserializer((urlSearchParameters.get(urlKey) as unknown) as SerializedState),
        [urlSearchParameters, urlKey, deserializer]
    )

    const setNextState = useCallback(
        (nextState: State) => {
            urlSearchParameters.set(urlKey, serializer(nextState))

            history.replace({ search: `?${urlSearchParameters.toString()}` })
        },
        [history, serializer, urlKey, urlSearchParameters]
    )

    return [queryParameter, setNextState]
}

type SerializedAggregationMode = `${AggregationMode}`

const aggregationModeSerializer = (mode: AggregationMode): SerializedAggregationMode => `${mode}`

const aggregationModeDeserializer = (serializedValue: SerializedAggregationMode | null): AggregationMode => {
    switch (serializedValue) {
        case 'repo':
            return AggregationMode.Repository
        case 'file':
            return AggregationMode.FilePath
        case 'author':
            return AggregationMode.Author
        case 'captureGroup':
            return AggregationMode.CaptureGroups

        default:
            return AggregationMode.Repository
    }
}

/**
 * Shared state hook for syncing aggregation type state between different UI trough
 * ULR query param {@link AGGREGATION_MODE_URL_KEY}
 */
export const useAggregationSearchMode = (): SetStateResult<AggregationMode> => {
    const [aggregationMode, setAggregationMode] = useSyncedWithURLState({
        urlKey: AGGREGATION_MODE_URL_KEY,
        serializer: aggregationModeSerializer,
        deserializer: aggregationModeDeserializer,
    })

    return [aggregationMode, setAggregationMode]
}

type SerializedAggregationUIMode = `${AggregationUIMode}`
const aggregationUIModeSerializer = (uiMode: AggregationUIMode): SerializedAggregationUIMode => `${uiMode}`

const aggregationUIModeDeserializer = (serializedValue: SerializedAggregationUIMode | null): AggregationUIMode => {
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
