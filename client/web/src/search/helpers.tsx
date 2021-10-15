import * as GQL from '@sourcegraph/shared/src/graphql/schema'
import { SubmitSearchParameters } from '@sourcegraph/shared/src/search/helpers'
import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { appendContextFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { SearchType } from '@sourcegraph/shared/src/search/stream'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'

import { eventLogger } from '../tracking/eventLogger'

const SUBMITTED_SEARCHES_COUNT_KEY = 'submitted-searches-count'

export function getSubmittedSearchesCount(): number {
    return parseInt(localStorage.getItem(SUBMITTED_SEARCHES_COUNT_KEY) || '0', 10)
}

export function canSubmitSearch(query: string, selectedSearchContextSpec?: string): boolean {
    // A standalone context: filter is also a valid search query
    return query !== '' || !!selectedSearchContextSpec
}

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
    activation,
    source,
    searchParameters,
}: SubmitSearchParameters): void {
    let searchQueryParameter = buildSearchURLQuery(
        query,
        patternType,
        caseSensitive,
        selectedSearchContextSpec,
        searchParameters
    )

    // Check if `trace` is set in the query parameters, and retain it if present.
    const existingParameters = new URLSearchParams(history.location.search)
    const traceParameter = existingParameters.get('trace')
    if (traceParameter !== null) {
        const parameters = new URLSearchParams(searchQueryParameter)
        parameters.set('trace', traceParameter)
        searchQueryParameter = parameters.toString()
    }

    // Go to search results page
    const path = '/search?' + searchQueryParameter
    eventLogger.log(
        'SearchSubmitted',
        {
            query: appendContextFilter(query, selectedSearchContextSpec),
            source,
        },
        { source }
    )
    localStorage.setItem(SUBMITTED_SEARCHES_COUNT_KEY, JSON.stringify(getSubmittedSearchesCount() + 1))
    history.push(path, { ...history.location.state, query })
    if (activation) {
        activation.update({ DidSearch: true })
    }
}

export function getSearchTypeFromQuery(query: string): SearchType {
    // RegExp to match `type:$TYPE` in any part of a query.
    const getTypeName = /\btype:(?<type>diff|commit|symbol|repo|path)\b/
    const matches = query.match(getTypeName)

    if (matches?.groups?.type) {
        // In an edge case where multiple `type:` filters are used, if
        // `type:symbol` is included, symbol results be returned, regardless of order,
        // so we must check for `type:symbol`. For other types,
        // the first `type` filter appearing in the query is applied.
        const symbolTypeRegex = /\btype:symbol\b/
        const symbolMatches = query.match(symbolTypeRegex)
        if (symbolMatches) {
            return 'symbol'
        }
        return matches.groups.type as SearchType
    }

    return null
}

/**
 * Adds the given search type (as a `type:` filter) into a query. This function replaces an existing `type:` filter,
 * appends a `type:` filter, or returns the initial query, in order to apply the correct type
 * to the query.
 *
 * @param query The search query to be mutated.
 * @param searchType The search type to be applied.
 */
export function toggleSearchType(query: string, searchType: SearchType): string {
    const match = query.match(/\btype:\w*\b/)
    if (!match) {
        return searchType ? `${query} type:${searchType}` : query
    }

    if (searchType !== null && match[0] === `type:${searchType}`) {
        // Query already contains correct search type
        return query
    }

    return query.replace(match[0], searchType ? `type:${searchType}` : '')
}

/** Returns true if the given value is of the GraphQL SearchResults type */
export const isSearchResults = (value: any): value is GQL.ISearchResults =>
    value && typeof value === 'object' && value.__typename === 'SearchResults'

/**
 * Some filters should use an alias just for search so they receive the expected suggestions.
 * See `./Suggestion.tsx->fuzzySearchFilters`.
 * E.g: `repohasfile` expects a file name as a value, so we should show `file` suggestions
 */
export const filterAliasForSearch: Record<string, FilterType | undefined> = {
    [FilterType.repohasfile]: FilterType.file,
}
