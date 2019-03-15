import * as H from 'history'
import * as GQL from '../../../shared/src/graphql/schema'
import { buildSearchURLQuery } from '../../../shared/src/util/url'
import { eventLogger } from '../tracking/eventLogger'

export function submitSearch(
    history: H.History,
    query: string,
    source: 'home' | 'nav' | 'repo' | 'tree' | 'filter'
): void {
    // Go to search results page
    const path = '/search?' + buildSearchURLQuery(query)
    eventLogger.log('SearchSubmitted', {
        code_search: {
            pattern: query,
            query,
            source,
        },
    })
    history.push(path, { ...history.location.state, query })
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
