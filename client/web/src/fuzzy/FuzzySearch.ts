import { HighlightedTextProps } from './HighlightedText'

export interface FuzzySearchParameters {
    query: string
    maxResults: number
    createUrl?: (value: string) => string
    onClick?: () => void
}

export interface FuzzySearchResult {
    results: HighlightedTextProps[]
    isComplete: boolean
    elapsedMilliseconds?: number
    falsePositiveRatio?: number
}

export interface SearchValue {
    text: string
}

export type IndexingFSM = SearchIndexing | SearchReady
export interface SearchIndexing {
    key: 'indexing'
    indexedFileCount: number
    totalFileCount: number
    partialValue: FuzzySearch
    continue: () => Promise<IndexingFSM>
}
export interface SearchReady {
    key: 'ready'
    value: FuzzySearch
}

/**
 * Superclass for different fuzzy finding algorithms.
 *
 * NOTE: Currently, there is only one implementation that uses bloom filters.
 * This implementation is specifically tailored for large repos that have >400k
 * source files. It could be that some users prefer a different fuzzy-finding
 * algorithm that is more fuzzy (relaxed with capitalization) but has worse
 * performance in XXL repos. Consider this superclass as an open invitation to
 * contribute a different fuzzy finding implementation that suits your personal
 * preferences :)
 */
export abstract class FuzzySearch {
    public abstract totalFileCount: number
    public abstract search(parameters: FuzzySearchParameters): FuzzySearchResult
}
