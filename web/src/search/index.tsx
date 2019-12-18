import { escapeRegExp } from 'lodash'
import { SearchPatternType } from '../../../shared/src/graphql/schema'
import { SuggestionTypes } from '../../../shared/src/search/suggestions/util'
import { FiltersToTypeAndValue } from '../../../shared/src/search/interactive/util'

/**
 * Parses the query out of the URL search params (the 'q' parameter). In non-interactive mode, if the 'q' parameter is not present, it
 * returns undefined. When parsing for interactive mode, each filter's individual query parameter
 * will be parsed and detected.
 *
 * @param query the URL query parameters
 * @param interactiveMode whether to parse the search URL query in interactive mode, reading query params such as `repo=` and `file=`.
 * @param navbarQueryOnly whether to only parse the query for the main query input, i.e. only the value passed to the `q=`
 * URL query parameter, as this represents the query that appears in the main query input in both modes.
 *
 */
export function parseSearchURLQuery(
    query: string,
    interactiveMode: boolean,
    navbarQueryOnly?: boolean
): string | undefined {
    if (!interactiveMode || navbarQueryOnly) {
        const searchParams = new URLSearchParams(query)
        return searchParams.get('q') || undefined
    }

    return interactiveParseSearchURLQuery(query)
}

/**
 * Parses the query out of the URL search params for interactive mode. This will parse
 * each individual filter's query parameter (for example, `file=` or `repo=`) in addition
 * to the raw query parameter (`q=`)
 *
 * @param query the URL query parameters
 */
export function interactiveParseSearchURLQuery(query: string): string | undefined {
    const searchParams = new URLSearchParams(query)
    const finalQueryParts = []
    for (const suggestionType of Object.keys(SuggestionTypes)) {
        for (const filterValue of searchParams.getAll(suggestionType)) {
            finalQueryParts.push(`${suggestionType}:${filterValue}`)
        }
    }

    const querySearchParams = searchParams.get('q')

    if (querySearchParams) {
        finalQueryParts.push(querySearchParams)
    }

    if (finalQueryParts.length > 0) {
        return finalQueryParts.join(' ')
    }

    return undefined
}

/**
 * Parses the pattern type out of the URL search params (the 'patternType' parameter). If the 'pattern' parameter
 * is not present, or it is an invalid value, it returns undefined.
 */
export function parseSearchURLPatternType(query: string): SearchPatternType | undefined {
    const searchParams = new URLSearchParams(query)
    const patternType = searchParams.get('patternType')
    if (
        patternType !== SearchPatternType.literal &&
        patternType !== SearchPatternType.regexp &&
        patternType !== SearchPatternType.structural
    ) {
        return undefined
    }
    return patternType
}

export function searchQueryForRepoRev(repoName: string, rev?: string): string {
    return `repo:${quoteIfNeeded(`^${escapeRegExp(repoName)}$${rev ? `@${abbreviateOID(rev)}` : ''}`)} `
}

function abbreviateOID(oid: string): string {
    if (oid.length === 40) {
        return oid.slice(0, 7)
    }
    return oid
}

export function quoteIfNeeded(s: string): string {
    if (/["' ]/.test(s)) {
        return JSON.stringify(s)
    }
    return s
}

export interface PatternTypeProps {
    patternType: SearchPatternType
    setPatternType: (patternType: SearchPatternType) => void
}

export interface InteractiveSearchProps {
    filtersInQuery: FiltersToTypeAndValue
    onFiltersInQueryChange: (filtersInQuery: FiltersToTypeAndValue) => void
    splitSearchModes: boolean
    interactiveSearchMode: boolean
    toggleSearchMode: (event: React.MouseEvent<HTMLAnchorElement>) => void
}
