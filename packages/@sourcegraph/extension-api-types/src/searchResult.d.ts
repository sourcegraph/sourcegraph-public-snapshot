import * as sourcegraph from 'sourcegraph'
import { Range } from './location'

/**
 * A search result.
 *
 * @see module:sourcegraph.SearchResult
 */
export interface SearchResult
    extends Pick<sourcegraph.SearchResult, Exclude<keyof sourcegraph.SearchResult, 'matches'>> {
    /** The highlight ranges in the match. */
    readonly matches: SearchResultMatch[]
}

/**
 * A search result match.
 *
 * @see module:sourcegraph.SearchResultMatch
 */
export interface SearchResultMatch
    extends Pick<sourcegraph.SearchResultMatch, Exclude<keyof sourcegraph.SearchResultMatch, 'highlights'>> {
    /** The highlight ranges in the match. */
    readonly highlights: Range[]
}
