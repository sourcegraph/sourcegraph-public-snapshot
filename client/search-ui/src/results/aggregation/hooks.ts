import { useCallback, useEffect, useMemo } from 'react'

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
        () => deserializer((urlSearchParameters.get(urlKey) as unknown) as SerializedState | null),
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

type SerializedAggregationMode = SearchAggregationMode | ''

const aggregationModeSerializer = (mode: SearchAggregationMode | null): SerializedAggregationMode => mode ?? ''

const aggregationModeDeserializer = (
    serializedValue: SerializedAggregationMode | null
): SearchAggregationMode | null => {
    switch (serializedValue) {
        case 'REPO':
            return SearchAggregationMode.REPO
        case 'PATH':
            return SearchAggregationMode.PATH
        case 'AUTHOR':
            return SearchAggregationMode.AUTHOR
        case 'CAPTURE_GROUP':
            return SearchAggregationMode.CAPTURE_GROUP

        // TODO Return null FE default value instead REPO when aggregation type
        // will be provided by the backend.
        // see https://github.com/sourcegraph/sourcegraph/issues/40425
        default:
            return SearchAggregationMode.REPO
    }
}

/**
 * Shared state hook for syncing aggregation type state between different UI trough
 * ULR query param {@link AGGREGATION_MODE_URL_KEY}
 */
export const useAggregationSearchMode = (): SetStateResult<SearchAggregationMode | null> => {
    const [aggregationMode, setAggregationMode] = useSyncedWithURLState<
        SearchAggregationMode | null,
        SerializedAggregationMode
    >({
        urlKey: AGGREGATION_MODE_URL_KEY,
        serializer: aggregationModeSerializer,
        deserializer: aggregationModeDeserializer,
    })

    return [aggregationMode, setAggregationMode]
}

type SerializedAggregationUIMode = AggregationUIMode
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

    query GetSearchAggregation(
        $query: String!
        $patternType: SearchPatternType!
        $mode: SearchAggregationMode
        $limit: Int!
    ) {
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
    aggregationMode: SearchAggregationMode | null
    limit: number
}

type SearchAggregationResults =
    | { data: undefined; loading: true; error: undefined }
    | { data: undefined; loading: false; error: Error }
    | { data: GetSearchAggregationResult; loading: false; error: undefined }

export const useSearchAggregationData = (input: SearchAggregationDataInput): SearchAggregationResults => {
    const { query, patternType, aggregationMode, limit } = input

    const [, setAggregationMode] = useAggregationSearchMode()
    const { data, error, loading } = useQuery<GetSearchAggregationResult, GetSearchAggregationVariables>(
        AGGREGATION_SEARCH_QUERY,
        {
            fetchPolicy: 'cache-first',
            variables: { query, patternType, mode: aggregationMode, limit },
        }
    )

    const calculatedAggregationMode = getCalculatedAggregationMode(data)

    useEffect(() => {
        // When we load the search result page in the first time we don't have picked
        // aggregation mode yet (unless we open the search result page with predefined
        // aggregation mode in the page URL)
        // In case when we don't have set aggregation mode on the FE, BE will
        // calculate this mode based on query that we pass to the aggregation
        // query (see AGGREGATION_SEARCH_QUERY).
        // When this happens we should take calculated aggregation mode and set its
        // value on the frontend (UI controls, update URL value of aggregation mode)

        // Catch initial page mount when aggregation mode isn't set on the FE and BE
        // calculated aggregation mode automatically on the backend based on given query
        if (calculatedAggregationMode && aggregationMode === null) {
            setAggregationMode(calculatedAggregationMode)
        }
    }, [setAggregationMode, calculatedAggregationMode, aggregationMode])

    if (loading) {
        return { data: undefined, error: undefined, loading: true }
    }

    const calculatedError = getAggregationError(error, data)

    if (calculatedError) {
        return { data: undefined, error: calculatedError, loading: false }
    }

    return {
        data: data as GetSearchAggregationResult,
        error: undefined,
        loading: false,
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

export function getAggregationData(response: GetSearchAggregationResult): SearchAggregationDatum[] {
    const aggregationResult = response.searchQueryAggregate?.aggregations

    switch (aggregationResult?.__typename) {
        case 'ExhaustiveSearchAggregationResult':
        case 'NonExhaustiveSearchAggregationResult':
            return aggregationResult.groups

        default:
            return []
    }
}

function getCalculatedAggregationMode(response?: GetSearchAggregationResult): SearchAggregationMode | null {
    if (!response) {
        return null
    }

    const aggregationResult = response.searchQueryAggregate?.aggregations

    return aggregationResult?.mode ?? null
}

export function getOtherGroupCount(response: GetSearchAggregationResult): number {
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
