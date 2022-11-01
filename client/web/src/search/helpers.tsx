import { SubmitSearchParameters } from '@sourcegraph/search'
import { appendContextFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'

import { eventLogger } from '../tracking/eventLogger'

import { AGGREGATION_MODE_URL_KEY, AGGREGATION_UI_MODE_URL_KEY } from './results/components/aggregation/constants'

/**
 * By default {@link submitSearch} overrides all existing query parameters.
 * This breaks all functionality that is built on top of URL query params and history
 * state. This list of query keys will be preserved between searches.
 */
const PRESERVED_QUERY_PARAMETERS = ['trace', AGGREGATION_MODE_URL_KEY, AGGREGATION_UI_MODE_URL_KEY]

/**
 * @param activation If set, records the DidSearch activation event for the new user activation
 * flow.
 */
export function submitSearch({
    history,
    query,
    patternType,
    caseSensitive,
    selectedSearchContextSpec,
    source,
    searchParameters,
    addRecentSearch,
}: SubmitSearchParameters): void {
    let searchQueryParameter = buildSearchURLQuery(
        query,
        patternType,
        caseSensitive,
        selectedSearchContextSpec,
        searchParameters
    )

    const existingParameters = new URLSearchParams(history.location.search)

    for (const key of PRESERVED_QUERY_PARAMETERS) {
        const queryParameter = existingParameters.get(key)

        if (queryParameter !== null) {
            const parameters = new URLSearchParams(searchQueryParameter)
            parameters.set(key, queryParameter)
            searchQueryParameter = parameters.toString()
        }
    }

    // Go to search results page
    const path = '/search?' + searchQueryParameter

    const queryWithContext = appendContextFilter(query, selectedSearchContextSpec)
    eventLogger.log(
        'SearchSubmitted',
        {
            query: queryWithContext,
            source,
        },
        { source }
    )
    addRecentSearch?.(queryWithContext)
    history.push(path, { ...(typeof history.location.state === 'object' ? history.location.state : null), query })
}
