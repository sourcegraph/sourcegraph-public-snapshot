import { extendedMatch, Fzf } from 'fzf'

import { HighlightedLinkProps, RangePosition } from '../components/fuzzyFinder/HighlightedLink'

import { FuzzySearch, FuzzySearchParameters, FuzzySearchResult, SearchValue } from './FuzzySearch'
import { createUrlFunction } from './WordSensitiveFuzzySearch'

/**
 * FuzzySearch implementation that uses the original fzy filtering algorithm from https://github.com/jhawthorn/fzy.js
 */
export class CaseInsensitiveFuzzySearch extends FuzzySearch {
    public totalFileCount: number

    constructor(public readonly values: SearchValue[], private readonly createUrl: createUrlFunction) {
        super()
        this.totalFileCount = values.length
    }

    public search(parameters: FuzzySearchParameters): FuzzySearchResult {
        const fzf = new Fzf<SearchValue[]>(this.values, {
            selector: ({ text }) => text,
            limit: parameters.maxResults,
            match: extendedMatch,
            tiebreakers: [(a, b) => (b.item?.ranking ?? 0) - (a.item?.ranking ?? 0)],
        })
        const candidates = fzf.find(parameters.query)
        // this.cacheCandidates.push(new CacheCandidate(parameters.query, [...candidates.map(({ item }) => item)]))
        const isComplete = candidates.length < parameters.maxResults
        candidates.slice(0, parameters.maxResults)

        const links: HighlightedLinkProps[] = candidates.map(candidate => {
            const positions = compressedRangePositions([...candidate.positions])
            return {
                ...candidate.item,
                positions,
                url: candidate.item.url || this.createUrl?.(candidate.item.text),
            }
        })
        return {
            isComplete,
            links,
        }
    }
}

/**
 * Returns the minimal number of range positions that enclose the given offset positions.
 *
 * Consecutive offset positions get compressed into a single range position. For example, the
 * offsets [1, 2, 3, 5, 6] become two range positions [1-3, 5-6].
 */
function compressedRangePositions(offsets: number[]): RangePosition[] {
    offsets.sort((a, b) => a - b)
    const ranges: RangePosition[] = []
    let index = 0
    while (index < offsets.length) {
        const start = offsets[index]
        index++
        while (index < offsets.length && offsets[index] === offsets[index - 1] + 1) {
            index++
        }
        const end = offsets[index - 1] + 1
        ranges.push({ startOffset: start, endOffset: end, isExact: false })
    }
    return ranges
}
