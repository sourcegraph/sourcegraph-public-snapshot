import { escapeRegExp } from 'lodash'
import { replaceRange } from '../../../shared/src/util/strings'
import { discreteValueAliases, escapeSpaces } from '../../../shared/src/search/query/filters'
import { VersionContext } from '../schema/site.schema'
import { SearchPatternType } from '../../../shared/src/graphql-operations'
import { Observable } from 'rxjs'
import { ISavedSearch, ISearchContext } from '../../../shared/src/graphql/schema'
import { EventLogResult } from './backend'
import { AggregateStreamingSearchResults, StreamSearchOptions } from './stream'
import { findFilter, FilterKind } from '../../../shared/src/search/query/validate'
import { VersionContextProps } from '../../../shared/src/search/util'

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
    if (
        patternType !== SearchPatternType.literal &&
        patternType !== SearchPatternType.regexp &&
        patternType !== SearchPatternType.structural
    ) {
        return undefined
    }
    return patternType
}

/**
 * Parses the version context out of the URL search params (the 'c' parameter). If the version context
 * is not present, return undefined.
 */
function parseSearchURLVersionContext(query: string): string | undefined {
    const searchParameters = new URLSearchParams(query)
    const context = searchParameters.get('c')
    return context ?? undefined
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
    versionContext: string | undefined
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
    let finalQuery = parseSearchURLQuery(urlSearchQuery) || ''
    let patternType = parseSearchURLPatternType(urlSearchQuery)
    let caseSensitive = searchURLIsCaseSensitive(urlSearchQuery)

    const globalPatternType = findFilter(finalQuery, 'patterntype', FilterKind.Global)
    if (globalPatternType?.value && globalPatternType.value.type === 'literal') {
        // Any `patterntype:` filter in the query should override the patternType= URL query parameter if it exists.
        finalQuery = replaceRange(finalQuery, globalPatternType.range)
        patternType = globalPatternType.value.value as SearchPatternType
    }

    const globalCase = findFilter(finalQuery, 'case', FilterKind.Global)
    if (globalCase?.value && globalCase.value.type === 'literal') {
        // Any `case:` filter in the query should override the case= URL query parameter if it exists.
        finalQuery = replaceRange(finalQuery, globalCase.range)

        if (discreteValueAliases.yes.includes(globalCase.value.value)) {
            caseSensitive = true
        } else if (discreteValueAliases.no.includes(globalCase.value.value)) {
            caseSensitive = false
        }
    }

    if (appendCaseFilter) {
        // Invariant: If case:value was in the query, it is erased at this point. Add case:yes if needed.
        finalQuery = caseSensitive ? `${finalQuery} case:yes` : finalQuery
    }

    return {
        query: finalQuery,
        patternType,
        caseSensitive,
        versionContext: parseSearchURLVersionContext(urlSearchQuery),
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

export interface ParsedSearchQueryProps {
    parsedSearchQuery: string
    setParsedSearchQuery: (query: string) => void
}

export interface PatternTypeProps {
    patternType: SearchPatternType
    setPatternType: (patternType: SearchPatternType) => void
}

export interface CaseSensitivityProps {
    caseSensitive: boolean
    setCaseSensitivity: (caseSensitive: boolean) => void
}

export interface MutableVersionContextProps extends VersionContextProps {
    setVersionContext: (versionContext: string | undefined) => void
    availableVersionContexts: VersionContext[] | undefined
    previousVersionContext: string | null
}

export interface CopyQueryButtonProps {
    copyQueryButton: boolean
}

export interface RepogroupHomepageProps {
    showRepogroupHomepage: boolean
}

export interface OnboardingTourProps {
    showOnboardingTour: boolean
}

export interface SearchContextProps {
    showSearchContext: boolean
    availableSearchContexts: ISearchContext[]
    defaultSearchContextSpec: string
    selectedSearchContextSpec?: string
    setSelectedSearchContextSpec: (spec: string) => void
}

export interface ShowQueryBuilderProps {
    showQueryBuilder: boolean
}

export interface HomePanelsProps {
    showEnterpriseHomePanels: boolean
    fetchSavedSearches: () => Observable<ISavedSearch[]>
    fetchRecentSearches: (userId: string, first: number) => Observable<EventLogResult | null>
    fetchRecentFileViews: (userId: string, first: number) => Observable<EventLogResult | null>

    /** Function that returns current time (for stability in visual tests). */
    now?: () => Date
}

export interface SearchStreamingProps {
    streamSearch: (options: StreamSearchOptions) => Observable<AggregateStreamingSearchResults>
}

/**
 * Verifies whether a version context exists on an instance.
 *
 * For URLs that have a `c=$X` parameter, we must check that
 * the version $X actually exists before trying to search with it.
 *
 * If the version context doesn't exist or there are no available version contexts, return undefined to
 * use the default context.
 *
 * @param versionContext The version context to verify.
 * @param availableVersionContexts A list of all version contexts defined in site configuration.
 */
export function resolveVersionContext(
    versionContext: string | undefined,
    availableVersionContexts: VersionContext[] | undefined
): string | undefined {
    if (
        !versionContext ||
        !availableVersionContexts ||
        !availableVersionContexts.map(versionContext => versionContext.name).includes(versionContext) ||
        versionContext === 'default'
    ) {
        return undefined
    }

    return versionContext
}

export function isSearchContextSpecAvailable(spec: string, availableSearchContexts: ISearchContext[]): boolean {
    return availableSearchContexts.map(item => item.spec).includes(spec)
}

export function resolveSearchContextSpec(
    spec: string,
    availableSearchContexts: ISearchContext[],
    defaultSpec: string
): string {
    return isSearchContextSpecAvailable(spec, availableSearchContexts) ? spec : defaultSpec
}
