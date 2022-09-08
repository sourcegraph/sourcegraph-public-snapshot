import { Location } from 'history'
import { escapeRegExp, memoize } from 'lodash'
import { combineLatest, from, Observable, of } from 'rxjs'
import { startWith, switchMap, map, distinctUntilChanged } from 'rxjs/operators'

import { memoizeObservable, replaceRange } from '@sourcegraph/common'
import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { discreteValueAliases, escapeSpaces } from '@sourcegraph/shared/src/search/query/filters'
import { stringHuman } from '@sourcegraph/shared/src/search/query/printer'
import { findFilter, FilterKind, getGlobalSearchContextFilter } from '@sourcegraph/shared/src/search/query/query'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
import { createLiteral } from '@sourcegraph/shared/src/search/query/token'
import { omitFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { AggregateStreamingSearchResults, StreamSearchOptions } from '@sourcegraph/shared/src/search/stream'

/**
 * Parses the query out of the URL search params (the 'q' parameter). In non-interactive mode, if the 'q' parameter is not present, it
 * returns undefined. When parsing for interactive mode, each filter's individual query parameter
 * will be parsed and detected.
 *
 * @param query the URL query parameters
 *
 */
export function parseSearchURLQuery(query: string): string | undefined {
    const searchParameters = new URLSearchParams(query)
    return searchParameters.get('q') || undefined
}

/**
 * Parses the pattern type out of the URL search params (the 'patternType' parameter). If the 'pattern' parameter
 * is not present, or it is an invalid value, it returns undefined.
 */
export function parseSearchURLPatternType(query: string): SearchPatternType | undefined {
    const searchParameters = new URLSearchParams(query)
    const patternType = searchParameters.get('patternType')
    switch (patternType) {
        case SearchPatternType.literal:
        case SearchPatternType.standard:
        case SearchPatternType.regexp:
        case SearchPatternType.structural:
        case SearchPatternType.lucky:
        case SearchPatternType.keyword:
            return patternType
    }
    return undefined
}

function searchURLIsCaseSensitive(query: string): boolean {
    const globalCase = findFilter(parseSearchURLQuery(query) || '', 'case', FilterKind.Global)
    if (globalCase?.value && globalCase.value.type === 'literal') {
        // if `case:` filter exists in the query, override the existing case: query param
        return discreteValueAliases.yes.includes(globalCase.value.value)
    }
    const searchParameters = new URLSearchParams(query)
    const caseSensitive = searchParameters.get('case')
    return discreteValueAliases.yes.includes(caseSensitive || '')
}

export interface ParsedSearchURL {
    query: string | undefined
    patternType: SearchPatternType | undefined
    caseSensitive: boolean
}

/**
 * parseSearchURL takes a URL's search querystring and returns
 * an object containing:
 * - the canonical, user-visible query (with `patternType` and `case` filters excluded),
 * - the effective pattern type, and
 * - the effective case sensitivity of the query.
 *
 * @param urlSearchQuery a URL's query string.
 */
export function parseSearchURL(
    urlSearchQuery: string,
    { appendCaseFilter = false }: { appendCaseFilter?: boolean } = {}
): ParsedSearchURL {
    let queryInput = parseSearchURLQuery(urlSearchQuery) || ''
    let patternTypeInput = parseSearchURLPatternType(urlSearchQuery)
    let caseSensitive = searchURLIsCaseSensitive(urlSearchQuery)

    const globalPatternType = findFilter(queryInput, 'patterntype', FilterKind.Global)
    if (globalPatternType?.value && globalPatternType.value.type === 'literal') {
        // Any `patterntype:` filter in the query should override the patternType= URL query parameter if it exists.
        queryInput = replaceRange(queryInput, globalPatternType.range)
        patternTypeInput = globalPatternType.value.value as SearchPatternType
    }

    let query = queryInput
    const { queryInput: newQuery, patternTypeInput: patternType } = literalSearchCompatibility({
        queryInput,
        patternTypeInput: patternTypeInput || SearchPatternType.standard,
    })
    query = newQuery

    const globalCase = findFilter(query, 'case', FilterKind.Global)
    if (globalCase?.value && globalCase.value.type === 'literal') {
        // Any `case:` filter in the query should override the case= URL query parameter if it exists.
        query = replaceRange(query, globalCase.range)

        if (discreteValueAliases.yes.includes(globalCase.value.value)) {
            caseSensitive = true
        } else if (discreteValueAliases.no.includes(globalCase.value.value)) {
            caseSensitive = false
        }
    }

    if (appendCaseFilter) {
        // Invariant: If case:value was in the query, it is erased at this point. Add case:yes if needed.
        query = caseSensitive ? `${query} case:yes` : query
    }

    return {
        query,
        patternType,
        caseSensitive,
    }
}

export function repoFilterForRepoRevision(repoName: string, globbing: boolean, revision?: string): string {
    if (globbing) {
        return `${escapeSpaces(`${repoName}${revision ? `@${abbreviateOID(revision)}` : ''}`)}`
    }
    return `${escapeSpaces(`^${escapeRegExp(repoName)}$${revision ? `@${abbreviateOID(revision)}` : ''}`)}`
}

export function searchQueryForRepoRevision(repoName: string, globbing: boolean, revision?: string): string {
    return `repo:${repoFilterForRepoRevision(repoName, globbing, revision)} `
}

function abbreviateOID(oid: string): string {
    if (oid.length === 40) {
        return oid.slice(0, 7)
    }
    return oid
}

export function quoteIfNeeded(string: string): string {
    if (/[ "']/.test(string)) {
        return JSON.stringify(string)
    }
    return string
}

interface QueryCompatibility {
    queryInput: string
    patternTypeInput: SearchPatternType
}

export function literalSearchCompatibility({ queryInput, patternTypeInput }: QueryCompatibility): QueryCompatibility {
    if (patternTypeInput !== SearchPatternType.literal) {
        return { queryInput, patternTypeInput }
    }
    const tokens = scanSearchQuery(queryInput, false, SearchPatternType.standard)
    if (tokens.type === 'error') {
        return { queryInput, patternTypeInput }
    }

    if (!tokens.term.find(token => token.type === 'pattern' && token.delimited)) {
        // If no /.../ pattern exists in this literal search, just return the query as-is.
        return { queryInput, patternTypeInput: SearchPatternType.standard }
    }

    const newQueryInput = stringHuman(
        tokens.term.map(token =>
            token.type === 'pattern' && token.delimited
                ? {
                      type: 'filter',
                      range: { start: 0, end: 0 },
                      field: createLiteral('content', { start: 0, end: 0 }, false),
                      value: createLiteral(`/${token.value}/`, { start: 0, end: 0 }, true),
                      negated: false /** if `NOT` was used on this pattern, it's already preserved */,
                  }
                : token
        )
    )

    return {
        queryInput: newQueryInput,
        patternTypeInput: SearchPatternType.standard,
    }
}

export interface HomePanelsProps {
    /** Function that returns current time (for stability in visual tests). */
    now?: () => Date
}

export interface SearchStreamingProps {
    streamSearch: (
        queryObservable: Observable<string>,
        options: StreamSearchOptions
    ) => Observable<AggregateStreamingSearchResults>
}

/**
 * getQueryStateFromLocation listens to history changes (via the location
 * observable) and extracts search query information while also handling and
 * validating search context information.
 * It returns an observable that emits the parsed query information, the
 * extracted search context (if any) and a processed version of the query
 * (without context filter)
 */
export function getQueryStateFromLocation({
    location,
    showSearchContext,
    isSearchContextAvailable,
}: {
    location: Observable<Location>
    /**
     * Whether or not the search context should be shown or not. This is usually
     * controlled by user settings.
     */
    showSearchContext: Observable<boolean>
    /**
     * Resolves to true if the provided search context exists for the user.
     */
    isSearchContextAvailable: (searchContextSpec: string) => Promise<boolean>
}): Observable<{
    /**
     * Query state from URL
     */
    parsedSearchURL: ParsedSearchURL
    /**
     * Search context extracted from query.
     */
    searchContextSpec: string | undefined
    /**
     * Cleaned up query (if search contexts are enabled and URL query contains a
     * search context, this property contains the query without the context
     * filter)
     */
    processedQuery: string
}> {
    // Memoized function to extract the `context:...` filter from a given
    // search query (avoids reparsing the query)
    const memoizedGetGlobalSearchContextSpec = memoize((query: string) => getGlobalSearchContextFilter(query))

    const memoizedIsSearchContextAvailable = memoizeObservable(
        (spec: string) =>
            spec
                ? // While we wait for the result of the `isSearchContextSpecAvailable` call, we assume the context is available
                  // to prevent flashing and moving content in the query bar. This optimizes for the most common use case where
                  // user selects a search context from the dropdown.
                  // See https://github.com/sourcegraph/sourcegraph/issues/19918 for more info.
                  from(isSearchContextAvailable(spec)).pipe(startWith(true), distinctUntilChanged())
                : of(false),
        spec => spec ?? ''
    )

    // This subscription handles updating the global query state store from
    // the URL.
    return combineLatest([
        // Extract information from URL and validate search context if
        // available.
        location.pipe(
            switchMap(
                (location): Observable<[ReturnType<typeof parseSearchURL>, boolean]> => {
                    const parsedQuery = parseSearchURL(location.search)
                    return memoizedIsSearchContextAvailable(
                        memoizedGetGlobalSearchContextSpec(parsedQuery.query ?? '')?.spec ?? ''
                    ).pipe(map(isSearchContextAvailable => [parsedQuery, isSearchContextAvailable]))
                }
            )
        ),
        showSearchContext.pipe(distinctUntilChanged()),
    ]).pipe(
        map(([[parsedSearchURL, isSearchContextAvailable], showSearchContext]) => {
            const query = parsedSearchURL.query ?? ''
            const globalSearchContextSpec = memoizedGetGlobalSearchContextSpec(query)
            const cleanQuery =
                // If a global search context spec is available to the user, we omit it from the
                // query and move it to the search contexts dropdown
                globalSearchContextSpec && isSearchContextAvailable && showSearchContext
                    ? omitFilter(query, globalSearchContextSpec.filter)
                    : query

            return { parsedSearchURL, searchContextSpec: globalSearchContextSpec?.spec, processedQuery: cleanQuery }
        })
    )
}
