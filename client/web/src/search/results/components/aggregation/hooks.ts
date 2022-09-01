import { useCallback, useEffect, useMemo, useRef, useState } from 'react'

import { gql, useQuery } from '@apollo/client'
import { useHistory, useLocation } from 'react-router'

import { SearchAggregationMode } from '@sourcegraph/shared/src/graphql-operations'
import { SearchPatternType } from '@sourcegraph/shared/src/schema'

import { GetSearchAggregationResult, GetSearchAggregationVariables } from '../../../../graphql-operations'

import { AGGREGATION_MODE_URL_KEY, AGGREGATION_UI_MODE_URL_KEY } from './constants'
import { AggregationUIMode } from './types'

interface URLStateOptions<State, SerializedState> {
    urlKey: string
    deserializer: (value: SerializedState | null) => State
    serializer: (state: State) => string | null
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
            const serializedValue = serializer(nextState)

            if (serializedValue === null) {
                urlSearchParameters.delete(urlKey)
            } else {
                urlSearchParameters.set(urlKey, serializedValue)
            }

            history.replace({ search: `?${urlSearchParameters.toString()}` })
        },
        [history, serializer, urlKey, urlSearchParameters]
    )

    return [queryParameter, setNextState]
}

type SerializedAggregationMode = 'repo' | 'path' | 'author' | 'group' | ''

const aggregationModeSerializer = (mode: SearchAggregationMode | null): SerializedAggregationMode => {
    switch (mode) {
        case SearchAggregationMode.REPO:
            return 'repo'
        case SearchAggregationMode.PATH:
            return 'path'
        case SearchAggregationMode.AUTHOR:
            return 'author'
        case SearchAggregationMode.CAPTURE_GROUP:
            return 'group'

        default:
            return ''
    }
}

const aggregationModeDeserializer = (
    serializedValue: SerializedAggregationMode | null
): SearchAggregationMode | null => {
    switch (serializedValue) {
        case 'repo':
            return SearchAggregationMode.REPO
        case 'path':
            return SearchAggregationMode.PATH
        case 'author':
            return SearchAggregationMode.AUTHOR
        case 'group':
            return SearchAggregationMode.CAPTURE_GROUP

        default:
            return null
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

/**
 * Serialized UI mode values
 * '' means that we use query param key existence as a sign that = foo="bar"&extended
 * null means that we remove mode value key form the URL = foo="bar"
 */
type SerializedAggregationUIMode = '' | null

const aggregationUIModeSerializer = (uiMode: AggregationUIMode): SerializedAggregationUIMode => {
    switch (uiMode) {
        case AggregationUIMode.SearchPage:
            return ''
        // Null means here that we will delete uiMode query param from the URL
        case AggregationUIMode.Sidebar:
            return null
    }
}

const aggregationUIModeDeserializer = (serializedValue: SerializedAggregationUIMode | null): AggregationUIMode => {
    switch (serializedValue) {
        case '':
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
        $skipAggregation: Boolean!
    ) {
        searchQueryAggregate(query: $query, patternType: $patternType) {
            aggregations(mode: $mode, limit: $limit) @skip(if: $skipAggregation) {
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
    proactive?: boolean
}

type SearchAggregationResults =
    | { data: undefined; loading: true; error: undefined }
    | { data: GetSearchAggregationResult | undefined; loading: false; error: Error }
    | { data: GetSearchAggregationResult; loading: false; error: undefined }

export const useSearchAggregationData = (input: SearchAggregationDataInput): SearchAggregationResults => {
    const { query, patternType, aggregationMode, limit, proactive } = input

    const calculatedAggregationModeRef = useRef<SearchAggregationMode | null>(null)
    const [, setAggregationMode] = useAggregationSearchMode()

    const [data, setData] = useState<GetSearchAggregationResult | undefined>()
    const { error, loading } = useQuery<GetSearchAggregationResult, GetSearchAggregationVariables>(
        AGGREGATION_SEARCH_QUERY,
        {
            fetchPolicy: 'cache-first',
            variables: {
                query,
                patternType,
                mode: aggregationMode,
                limit,
                skipAggregation: aggregationMode === null && !proactive,
            },

            // Skip extra API request when we had no aggregation mode, and then
            // we got calculated aggregation mode from the BE. We should update
            // FE aggregationMode but this shouldn't trigger AGGREGATION_SEARCH_QUERY
            // fetching.
            skip: aggregationMode !== null && calculatedAggregationModeRef.current === aggregationMode,
            onError: () => {
                calculatedAggregationModeRef.current = null
                setData(undefined)
            },
            onCompleted: data => {
                const calculatedAggregationMode = getCalculatedAggregationMode(data)

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
                if (calculatedAggregationMode !== aggregationMode) {
                    setAggregationMode(calculatedAggregationMode)
                }

                // Preserve calculated aggregation mode in order to use it for skipping
                // extra API calls in useQuery "skip" field.
                calculatedAggregationModeRef.current = calculatedAggregationMode

                // skip: true resets data field in the useQuery hook, in order to use previously
                // saved data we use useState to store data outside useQuery hook
                setData(data)
            },
        }
    )

    useEffect(() => {
        // If query or pattern type have been changed we should "reset" our assumptions
        // about calculated aggregation mode and make another api call to determine it
        calculatedAggregationModeRef.current = null
    }, [query, patternType])

    if (loading) {
        return { data: undefined, error: undefined, loading: true }
    }

    if (error) {
        return { data, error, loading: false }
    }

    return {
        data: data as GetSearchAggregationResult,
        error: undefined,
        loading: false,
    }
}

function getCalculatedAggregationMode(response?: GetSearchAggregationResult): SearchAggregationMode | null {
    if (!response) {
        return null
    }

    const aggregationResult = response.searchQueryAggregate?.aggregations

    return aggregationResult?.mode ?? null
}
