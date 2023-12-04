import type { FuzzyTabKey } from '../components/fuzzyFinder/FuzzyTabs'
import type { HighlightedLinkProps } from '../components/fuzzyFinder/HighlightedLink'

import type { SearchValueRankingCache } from './SearchValueRankingCache'

export interface FuzzySearchParameters {
    query: string
    maxResults: number
    cache?: SearchValueRankingCache
    fuzzyFinderTab?: FuzzyTabKey
}

export interface FuzzySearchResult {
    links: HighlightedLinkProps[]
    isComplete: boolean
    elapsedMilliseconds?: number
    falsePositiveRatio?: number
}

export enum SearchIconKind {
    codeHost,
    symbol,
}

export type IndexingFSM = SearchIndexing | SearchReady
export interface SearchIndexing {
    key: 'indexing'
    indexedFileCount: number
    totalFileCount: number
    partialFuzzy: FuzzySearch
    isIndexing: () => boolean
    continueIndexing: () => Promise<IndexingFSM>
}
export interface SearchReady {
    key: 'ready'
    value: FuzzySearch
}

/**
 * Superclass for different fuzzy finding algorithms.
 *
 * Currently, there is only one implementation that is case sensitive.  This
 * implementation is specifically tailored for large repos that have >400k
 * source files. Most users will likely prefer case-insensitive fuzzy filtering,
 * which is easy to support for small repos (<20k files) but it's not clear how
 * to support that in larger repos without sacrificing latency.
 *
 * Tracking issue to add case-insensitive search: https://github.com/sourcegraph/sourcegraph/issues/21201
 */
export abstract class FuzzySearch {
    public abstract totalFileCount: number
    public abstract search(parameters: FuzzySearchParameters): FuzzySearchResult
}

export interface FuzzySearchConstructorParameters {
    createURL?: createUrlFunction
    transformer?: FuzzySearchTransformer
}

export type createUrlFunction = undefined | ((value: string) => string)

export interface FuzzySearchTransformer {
    modifyQuery?: (query: string) => string
    modifyURL?: (query: string, url: string) => string
}
