import { extendedMatch, Fzf } from 'fzf'

import { HighlightedLinkProps, RangePosition } from '../components/fuzzyFinder/HighlightedLink'

import { FuzzySearch, FuzzySearchParameters, FuzzySearchResult, SearchValue } from './FuzzySearch'
import { createUrlFunction } from './WordSensitiveFuzzySearch'

class CacheCandidate {
    constructor(public readonly query: string, public readonly candidates: SearchValue[]) {}
    public matches(parameters: FuzzySearchParameters): boolean {
        return parameters.query.startsWith(this.query)
    }
}

/**
 * FuzzySearch implementation that uses the original fzy filtering algorithm from https://github.com/jhawthorn/fzy.js
 */
export class CaseInsensitiveFuzzySearch extends FuzzySearch {
    public totalFileCount: number
    // Optimization: stack of results from the previous queries. For example,
    // when the user types ["P", "r", "o"] the stack contains the matching
    // results for the queries ["P", "Pr", "Pro"]. When we get the query "Prov"
    // we fuzzy match against the cached candidates for the query "Pro", which
    // is most likely faster compared to fuzzy matching against the entire
    // filename corpu. We cache all prefixes of the query instead of only the
    // last query to allow the user to quickly delete // multiple characters
    // from the query.
    private cacheCandidates: CacheCandidate[] = []

    constructor(public readonly values: SearchValue[], private readonly createUrl: createUrlFunction) {
        super()
        this.totalFileCount = values.length
    }

    public search(parameters: FuzzySearchParameters): FuzzySearchResult {
        const cacheCandidate = this.nextCacheCandidate(parameters)
        const searchValues: SearchValue[] = cacheCandidate ? cacheCandidate.candidates : this.values
        const fzf = new Fzf<SearchValue[]>(searchValues, {
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

    /**
     * Returns the results from the last query, if any, that is a prefix of the current query.
     *
     * Removes cached candidates that are no longer a prefix of the current
     * query.
     */
    private nextCacheCandidate(parameters: FuzzySearchParameters): CacheCandidate | undefined {
        if (parameters.query === '') {
            this.cacheCandidates = []
            return undefined
        }
        let cacheCandidate = this.lastCacheCandidate()
        while (cacheCandidate && !cacheCandidate.matches(parameters)) {
            this.cacheCandidates.pop()
            cacheCandidate = this.lastCacheCandidate()
        }
        return cacheCandidate
    }

    private lastCacheCandidate(): CacheCandidate | undefined {
        if (this.cacheCandidates.length > 0) {
            return this.cacheCandidates[this.cacheCandidates.length - 1]
        }
        return undefined
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
