import { useCallback, useMemo } from 'react'

import { ApolloError, gql, useQuery } from '@apollo/client'
import { useHistory, useLocation } from 'react-router'

import { SearchAggregationMode } from '@sourcegraph/shared/src/graphql-operations'
import { SearchPatternType } from '@sourcegraph/shared/src/schema'

import {
    GetSearchAggregationResult,
    GetSearchAggregationVariables,
    SearchAggregationDatum,
} from '../../graphql-operations'

import { AGGREGATION_MODE_URL_KEY, AGGREGATION_UI_MODE_URL_KEY } from './constants'
import { AggregationUIMode } from './types'

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

type SerializedAggregationMode = `${SearchAggregationMode}`

const aggregationModeSerializer = (mode: SearchAggregationMode): SerializedAggregationMode => mode

const aggregationModeDeserializer = (serializedValue: SerializedAggregationMode | null): SearchAggregationMode => {
    switch (serializedValue) {
        case 'REPO':
            return SearchAggregationMode.REPO
        case 'PATH':
            return SearchAggregationMode.PATH
        case 'AUTHOR':
            return SearchAggregationMode.AUTHOR
        case 'CAPTURE_GROUP':
            return SearchAggregationMode.CAPTURE_GROUP

        default:
            return SearchAggregationMode.REPO
    }
}

/**
 * Shared state hook for syncing aggregation type state between different UI trough
 * ULR query param {@link AGGREGATION_MODE_URL_KEY}
 */
export const useAggregationSearchMode = (): SetStateResult<SearchAggregationMode> => {
    const [aggregationMode, setAggregationMode] = useSyncedWithURLState({
        urlKey: AGGREGATION_MODE_URL_KEY,
        serializer: aggregationModeSerializer,
        deserializer: aggregationModeDeserializer,
    })

    return [aggregationMode, setAggregationMode]
}

type SerializedAggregationUIMode = `${AggregationUIMode}`
const aggregationUIModeSerializer = (uiMode: AggregationUIMode): SerializedAggregationUIMode => uiMode

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

export const AGGREGATION_SEARCH_QUERY = gql`
    fragment SearchAggregationModeAvailability on AggregationModeAvailability {
        __typename
        mode
        available
        reasonUnavailable
    }

    fragment SearchAggregationDatum on AggregationGroup {
        __typename
        label
        count
        query
    }

    query GetSearchAggregation($query: String!, $patternType: SearchPatternType!, $mode: SearchAggregationMode, $limit: Int!) {
        searchQueryAggregate(query: $query, patternType: $patternType) {
            aggregations(mode: $mode, limit: $limit) {
                __typename
                ... on ExhaustiveSearchAggregationResult {
                    mode
                    groups {
                        ...SearchAggregationDatum
                    }
                    otherGroupCount
                }

                ... on NonExhaustiveSearchAggregationResult {
                    mode
                    groups {
                        ...SearchAggregationDatum
                    }
                    approximateOtherGroupCount
                }

                ... on SearchAggregationNotAvailable {
                    reason
                    mode
                }
            }
            modeAvailability {
                ...SearchAggregationModeAvailability
            }
        }
    }
`

interface SearchAggregationDataInput {
    query: string
    patternType: SearchPatternType
    aggregationMode: SearchAggregationMode
    limit?: number
}

interface SearchAggregationResults {
    data: GetSearchAggregationResult | undefined
    loading: boolean
    error: Error | undefined
}

export const useSearchAggregationData = (input: SearchAggregationDataInput): SearchAggregationResults => {
    const { query, patternType, aggregationMode, limit = 10 } = input

    const { data, error, loading } = useQuery<GetSearchAggregationResult, GetSearchAggregationVariables>(
        AGGREGATION_SEARCH_QUERY,
        {
            fetchPolicy: 'cache-first',
            variables: { query, patternType, mode: aggregationMode, limit },
        }
    )

    return {
        data,
        loading,
        // We need to handle error properly for cases when we got errors
        // in error field (network request error, gql error) or in data
        // response (data aggregation error)
        error: getAggregationError(error, data),
    }
}

function getAggregationError(apolloError?: ApolloError, response?: GetSearchAggregationResult): Error | undefined {
    if (apolloError) {
        return apolloError
    }

    const aggregationData = response?.searchQueryAggregate?.aggregations

    if (aggregationData?.__typename === 'SearchAggregationNotAvailable') {
        return new Error(aggregationData.reason)
    }

    return
}

export function getAggregationData(response?: GetSearchAggregationResult | null): SearchAggregationDatum[] {
    if (!response) {
        return []
    }

    const aggregationResult = response.searchQueryAggregate?.aggregations

    switch (aggregationResult?.__typename) {
        case 'ExhaustiveSearchAggregationResult':
        case 'NonExhaustiveSearchAggregationResult':
            return aggregationResult.groups

        default:
            return []
    }
}

export function getOtherGroupCount(response?: GetSearchAggregationResult): number {
    if (!response) {
        return 0
    }

    const aggregationResult = response.searchQueryAggregate?.aggregations

    switch (aggregationResult?.__typename) {
        case 'ExhaustiveSearchAggregationResult':
            return aggregationResult.otherGroupCount ?? 0
        case 'NonExhaustiveSearchAggregationResult':
            return aggregationResult.approximateOtherGroupCount ?? 0

        default:
            return 0
    }
}
