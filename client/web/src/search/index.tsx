import { escapeRegExp } from 'lodash'
import { Observable } from 'rxjs'

import { replaceRange } from '@sourcegraph/common'
import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { discreteValueAliases, escapeSpaces } from '@sourcegraph/shared/src/search/query/filters'
import { stringHuman } from '@sourcegraph/shared/src/search/query/printer'
import { findFilter, FilterKind } from '@sourcegraph/shared/src/search/query/query'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
import { createLiteral } from '@sourcegraph/shared/src/search/query/token'
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
