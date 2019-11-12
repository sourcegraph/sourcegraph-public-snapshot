import * as H from 'history'
import _ from 'lodash'
import { ActivationProps } from '../../../shared/src/components/activation/Activation'
import * as GQL from '../../../shared/src/graphql/schema'
import { buildSearchURLQuery } from '../../../shared/src/util/url'
import { eventLogger } from '../tracking/eventLogger'
import { SearchType } from './results/SearchResults'
import { SearchFilterSuggestions } from './searchFilterSuggestions'
import {
    Suggestion,
    SuggestionTypes,
    FiltersSuggestionTypes,
    isolatedFuzzySearchFilters,
    filterAliases,
} from './input/Suggestion'

/**
 * @param activation If set, records the DidSearch activation event for the new user activation
 * flow.
 */
export function submitSearch(
    history: H.History,
    query: string,
    source: 'home' | 'nav' | 'repo' | 'tree' | 'filter' | 'type',
    patternType: GQL.SearchPatternType,
    activation?: ActivationProps['activation']
): void {
    // Go to search results page
    const path = '/search?' + buildSearchURLQuery(query, patternType)
    eventLogger.log('SearchSubmitted', {
        code_search: {
            pattern: query,
            query,
            source,
        },
    })
    history.push(path, { ...history.location.state, query })
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
    let idx = 0
    while (true) {
        idx = query.indexOf(scope, idx)
        if (idx === -1) {
            break
        }

        // prevent matching scopes that are substrings of other scopes
        if (idx > 0 && query[idx - 1] !== ' ') {
            idx = idx + 1
        } else {
            break
        }
    }
    return idx
}

/**
 * Toggles the given search scope by adding or removing it from the current
 * user query string.
 *
 * @param query The current user query.
 * @param searchFilter The search scope (sub query) or dynamic filter to toggle (add/remove) from the current user query.
 * @returns The new query.
 */
export function toggleSearchFilter(query: string, searchFilter: string): string {
    const idx = queryIndexOfScope(query, searchFilter)
    if (idx === -1) {
        // Scope doesn't exist in search query, so add it now.
        return [query.trim(), searchFilter].filter(s => s).join(' ') + ' '
    }

    // Scope exists in the search query, so remove it now.
    return (query.substring(0, idx).trim() + ' ' + query.substring(idx + searchFilter.length).trim()).trim()
}

export function getSearchTypeFromQuery(query: string): SearchType {
    // RegExp to match `type:$TYPE` in any part of a query.
    const getTypeName = /\btype:(?<type>diff|commit|symbol|repo|path)\b/
    const matches = query.match(getTypeName)

    if (matches && matches.groups && matches.groups.type) {
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

    if (match[0] === `type:${searchType}`) {
        /** Query already contains correct search type */
        return query
    }

    return query.replace(match[0], searchType ? `type:${searchType}` : '')
}

/** Returns true if the given value is of the GraphQL SearchResults type */
export const isSearchResults = (val: any): val is GQL.ISearchResults =>
    val && typeof val === 'object' && val.__typename === 'SearchResults'

/**
 * Toggles the given search scope by adding it or removing it from the current string, and removes `repogroup:sample`
 * from the query if it exists in the query, and the search scope being added contains a `repogroup:` filter.
 *
 * @param query the current user query
 * @param searchFilter the search scope (sub query) or dynamic filter to toggle (add/remove from the current user query)
 * @returns The new query
 */
export const toggleSearchFilterAndReplaceSampleRepogroup = (query: string, searchFilter: string): string => {
    const newQuery = toggleSearchFilter(query, searchFilter)
    // RegExp to replace `repogroup:sample` without removing leading whitespace.
    const replaceSampleRepogroupRegexp = /(\b|^)repogroup:sample(\s|$)/
    // RegExp to match `repogroup:sample` in any part of a query.
    const matchSampleRepogroupRegexp = /(\s*|^)repogroup:sample(\s*|$)/
    if (
        /\brepogroup:/.test(searchFilter) &&
        matchSampleRepogroupRegexp.test(newQuery) &&
        !matchSampleRepogroupRegexp.test(searchFilter)
    ) {
        return newQuery.replace(replaceSampleRepogroupRegexp, '')
    }
    return newQuery
}

export const isValidFilter = (filter: string = ''): filter is FiltersSuggestionTypes =>
    Object.prototype.hasOwnProperty.call(SuggestionTypes, filter) ||
    Object.prototype.hasOwnProperty.call(filterAliases, filter)

const isValidFilterAlias = (alias: string): alias is keyof typeof filterAliases =>
    Object.prototype.hasOwnProperty.call(filterAliases, alias)

/**
 * If a filter value is being typed, try to get its filter type.
 * E.g: with "|"" being the cursor: "repo:| lang:go" => "repo"
 * Checks if the matched word is a valid filter.
 */
export const lastFilterAndValueBeforeCursor = (
    queryState: QueryState
): {
    filter: SuggestionTypes,
    filterAndValue: string
    value: string
} | null => {
    const firstPart = queryState.query.substring(0, queryState.cursorPosition)
    // get string before ":" char until a space is found or start of string
    const [filterAndValue, filter] = firstPart.match(/([^\s:]+)?(:(\S?)+)?$/) || []
    const value = filterAndValue?.split(':')[1]?.trim()
    const absoluteFilter = filter?.replace(/^-/, '')
    const resolvedFilter = isValidFilterAlias(absoluteFilter) ? filterAliases[absoluteFilter] : absoluteFilter
    return isValidFilter(resolvedFilter) ? { filter: resolvedFilter, filterAndValue: filterAndValue.trim(), value } : null
}

/**
 * Returns suggestions for a given search query but only at the last typed word.
 * If the word does not contain ":" then it returns filter types as suggestions
 * If the word contains ":" then it returns suggestions for the typed filter.
 * For query "case:| archived:" where "|" is the cursor position, it
 * returns suggestions (filter values) for the "case" filter.
 */
export const filterStaticSuggestions = (
    queryState: QueryState,
    suggestions: SearchFilterSuggestions
): Suggestion[] => {
    const filterAndValueBeforeCursor = lastFilterAndValueBeforeCursor(queryState)

    if (!filterAndValueBeforeCursor) {
        return []
    }

    const { filter, value, filterAndValue } = filterAndValueBeforeCursor

    if (
        // suggest values for selected filter
        isValidFilter(filter) &&
        filter !== SuggestionTypes.filters &&
        (value || filterAndValue.endsWith(':'))
    ) {
        const suggestionsToShow = suggestions[filter] ?? []
        return suggestionsToShow.values.filter(suggestion => suggestion.value.startsWith(value))
    }

    // Suggest filter types
    return suggestions.filters.values.filter(({ value }) => value.startsWith(filter))
}

/**
 * The search query and cursor position of where the last character was inserted.
 * Cursor position is used to correctly insert the suggestion when it's selected,
 * and set the cursor to the end of where the suggestion was inserted.
 */
export interface QueryState {
    query: string
    cursorPosition: number
}

/**
 * Used to decide if the search is for a filter value or a fuzzy-search word.
 * "l:go yes" => true
 * "l:go archived:" => false
 */
export const isTypingWordAndNotFilterValue = (value: string): boolean => Boolean(value.match(/\s+([^:]?)+$/))

/**
 * Adds suggestions value to search query where cursor was positioned.
 * ('a test: query', 'suggestion', 7) => 'a test:suggestion query'
 */
export const insertSuggestionInQuery = (queryToInsertIn: string, selectedSuggestion: Suggestion, cursorPosition: number): QueryState => {
    const firstPart = queryToInsertIn.substring(0, cursorPosition)
    const lastPart = queryToInsertIn.substring(firstPart.length)
    const isFiltersSuggestion = selectedSuggestion.type === SuggestionTypes.filters
    // Know where to place the suggestion later on
    const separatorIndex = firstPart.lastIndexOf(!isFiltersSuggestion ? ':' : ' ')
    // If a filter value or separate word suggestion was selected, then append a whitespace
    const valueToAppend = selectedSuggestion.value + (isFiltersSuggestion ? '' : ' ')

    const newFirstPart = (() => {
        const lastWordOfFirstPartMatch = firstPart.match(/\s+(\S?)+$/)
        const isSeparateWordSuggestion = isTypingWordAndNotFilterValue(firstPart)

        // A fuzzy-search suggestion was selected but it doesn't have a URL...
        // This prevents the selected suggestion replacing a previous value in query.
        // e.g: (with "|" being the cursor)
        // without: "archived:Yes Query|" -> selection -> "archived:QueryInput"
        // with: "archived:Yes Query|" -> selection -> "archived:Yes QueryInput"
        if (
            !isFiltersSuggestion &&
            isSeparateWordSuggestion &&
            lastWordOfFirstPartMatch &&
            lastWordOfFirstPartMatch.index
        ) {
            // adds a space because a separate word was being typed
            return firstPart.substring(0, lastWordOfFirstPartMatch.index) + ' ' + valueToAppend + lastPart
        }

        return firstPart.substring(0, separatorIndex + 1) + valueToAppend
    })()

    return {
        // .replace() to remove excess whitespace in query
        query: (newFirstPart + lastPart).replace(/\s+/g, ' '),
        cursorPosition: newFirstPart.length,
    }
}

/**
 * Returns true if word being typed is not a filter value.
 * E.g: where "|" is cursor
 *     "QueryInput lang:|" => false
 *     "archived:Yes QueryInp|" => true
 */
export const isFuzzyWordSearch = (queryCursor: QueryState): boolean => {
    const firstPart = queryCursor.query.substring(0, queryCursor.cursorPosition)
    const isTypingFirstWord = Boolean(firstPart.match(/^(\s?)+[^:\s]+$/))
    return isTypingFirstWord || isTypingWordAndNotFilterValue(firstPart)
}

/**
 * Makes any modification to the query which will only be used
 * for fetching suggesitons, and should not mutate the query in state.
 */
export const formatQueryForFuzzySearch = (queryState: QueryState): string => {
    const filterAndValueBeforeCursor = lastFilterAndValueBeforeCursor(queryState)
    // Check if filter shold have its suggestions searched without influence from the rest of the query
    if (
        filterAndValueBeforeCursor &&
        isolatedFuzzySearchFilters.includes(filterAndValueBeforeCursor.filter)
    ) {
        return filterAndValueBeforeCursor.filterAndValue
    }
    // Use search results from `file` filter when suggesting for `repohasfile` filter.
    // Also requires the filter type to be in ./Suggestion.tsx->fuzzySearchFilters
    return queryState.query.replace(SuggestionTypes.repohasfile, 'file')
}
