import * as fzy from 'fzy.js'

import { HighlightedLinkProps, RangePosition } from '../components/fuzzyFinder/HighlightedLink'

import { FuzzySearch, FuzzySearchParameters, FuzzySearchResult, SearchValue } from './FuzzySearch'
import { createUrlFunction } from './WordSensitiveFuzzySearch'

interface ScoredSearchValue extends SearchValue {
    score: number
}

class CacheCandidate {
    constructor(public readonly query: string, public readonly candidates: SearchValue[]) {}
    public matches(parameters: FuzzySearchParameters): boolean {
        return parameters.query.startsWith(this.query)
    }
}

// The 0.2 value was chosen by manually observing the behavior and confirming
// that it seems to give relevant results without too much noise.
const FZY_MINIMUM_SCORE_THRESHOLD = 0.2

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
    private spaceSeparator = new RegExp('\\s+')

    constructor(public readonly values: SearchValue[], private readonly createUrl: createUrlFunction) {
        super()
        this.totalFileCount = values.length
    }

    public search(parameters: FuzzySearchParameters): FuzzySearchResult {
        const cacheCandidate = this.nextCacheCandidate(parameters)
        const searchValues: SearchValue[] = cacheCandidate ? cacheCandidate.candidates : this.values
        const isEmptyQuery = parameters.query.length === 0
        const candidates: ScoredSearchValue[] = []
        const queryParts = parameters.query.split(this.spaceSeparator).filter(part => part.length > 0)
        if (queryParts.length === 0) {
            // Empty query, match all values
            candidates.push(...searchValues.map(value => ({ ...value, score: 0 })))
        } else {
            for (const value of searchValues) {
                let score = 0
                if (queryParts.length === 1 && queryParts[0] === value.text) {
                    score = 1 // exact match, special-cased because fzy.score returns `Infinity`
                } else {
                    for (const queryPart of queryParts) {
                        // TODO: the query 'sourcegraph' should have a higher
                        // score for the value 'sourcegraph/sourcegraph' instead
                        // of 'sourcegraph/scip'. Right now, `sourcegraph/scip` scores
                        // equally.
                        const partScore = fzy.score(queryPart, value.text)
                        score += partScore
                    }
                }
                const isAcceptableScore = !isNaN(score) && isFinite(score) && score > FZY_MINIMUM_SCORE_THRESHOLD
                if (isEmptyQuery || isAcceptableScore) {
                    candidates.push({ ...value, score })
                }
            }
        }

        this.cacheCandidates.push(new CacheCandidate(parameters.query, [...candidates]))

        const isComplete = candidates.length < parameters.maxResults
        candidates.sort((a, b) => {
            const byScore = b.score - a.score
            if (byScore < 1 && a.ranking && b.ranking) {
                return b.ranking - a.ranking
            }
            return byScore
        })
        candidates.slice(0, parameters.maxResults)

        const links: HighlightedLinkProps[] = candidates.map(candidate => {
            const offsets = new Set<number>()
            for (const queryPart of queryParts) {
                for (const offset of fzy.positions(queryPart, candidate.text)) {
                    offsets.add(offset)
                }
            }
            const positions = compressedRangePositions([...offsets])
            return {
                ...candidate,
                positions,
                url: candidate.url || this.createUrl?.(candidate.text),
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
