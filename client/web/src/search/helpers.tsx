import { SubmitSearchParameters } from '@sourcegraph/search'
import * as GQL from '@sourcegraph/shared/src/schema'
import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { appendContextFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { SearchType } from '@sourcegraph/shared/src/search/stream'
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
    eventLogger.log(
        'SearchSubmitted',
        {
            query: appendContextFilter(query, selectedSearchContextSpec),
            source,
        },
        { source }
    )
    history.push(path, { ...(typeof history.location.state === 'object' ? history.location.state : null), query })
    if (activation) {
        activation.update({ DidSearch: true })
    }
}

/**
 * Returns the index that a given search scope occurs in a given search query.
 * Attempts to not match a scope that is a substring of another scope.
 *
 * @param query The full query string
 * @param scope A scope (sub query) that is searched for in `query`
 * @returns The index in `query`, or `-1` if not found
 */
export function queryIndexOfScope(query: string, scope: string): number {
    let index = 0
    while (true) {
        index = query.indexOf(scope, index)
        if (index === -1) {
            break
        }

        const boundAtStart = index === 0 || query[index - 1] === ' '
        const boundAtEnd = index + scope.length === query.length || query[index + scope.length] === ' '

        // prevent matching scopes that are substrings of other scopes
        if (!boundAtStart || !boundAtEnd) {
            index = index + 1
        } else {
            break
        }
    }
    return index
}

/**
 * Toggles the given search scope by adding or removing it from the current
 * user query string.
 *
 * @param query The current user query.
 * @param searchFilter The search scope (sub query) or dynamic filter to toggle (add/remove) from the current user query.
 * @returns The new query.
 */
export function toggleSubquery(query: string, searchFilter: string): string {
    const index = queryIndexOfScope(query, searchFilter)
    if (index === -1) {
        // Scope doesn't exist in search query, so add it now.
        return [query.trim(), searchFilter].filter(string => string).join(' ') + ' '
    }

    // Scope exists in the search query, so remove it now.
    return (query.slice(0, index).trim() + ' ' + query.slice(index + searchFilter.length).trim()).trim()
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
