import { HighlightedTextProps } from './HighlightedText'

export interface FuzzySearchParameters {
    value: string
    maxResults: number
    createUrl?: (value: string) => string
    onClick?: () => void
}
export interface FuzzySearchResult {
    values: HighlightedTextProps[]
    isComplete: boolean
    elapsedMilliseconds?: number
    falsePositiveRatio?: number
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
