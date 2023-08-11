import { extendedMatch, Fzf, type FzfResultItem, type Tiebreaker } from 'fzf'

import type { HighlightedLinkProps, RangePosition } from '../components/fuzzyFinder/HighlightedLink'

import {
    FuzzySearch,
    type FuzzySearchConstructorParameters,
    type FuzzySearchParameters,
    type FuzzySearchResult,
} from './FuzzySearch'
import type { SearchValue } from './SearchValue'
import { SearchValueRankingCache } from './SearchValueRankingCache'

function sortByTiebreakers<T>(values: FzfResultItem<T>[], tiebreakers: Tiebreaker<T>[]): FzfResultItem<T>[] {
    return values.sort((a, b) => {
        for (const tiebreaker of tiebreakers) {
            const compareResult = tiebreaker(a, b, () => '')
            if (compareResult !== 0) {
                return compareResult
            }
        }
        return 0
    })
}

/**
 * FuzzySearch implementation that uses the original fzy filtering algorithm from https://github.com/jhawthorn/fzy.js
 */
export class CaseInsensitiveFuzzySearch extends FuzzySearch {
    public totalFileCount: number

    constructor(public readonly values: SearchValue[], private readonly params?: FuzzySearchConstructorParameters) {
        super()
        this.totalFileCount = values.length
    }

    public search(parameters: FuzzySearchParameters): FuzzySearchResult {
        const query = this?.params?.transformer?.modifyQuery
            ? this.params.transformer.modifyQuery(parameters.query)
            : parameters.query
        const historyCache = parameters.cache ?? new SearchValueRankingCache()
        const tiebreakers: Tiebreaker<SearchValue>[] = [
            (a, b) => historyCache.rank(b.item) - historyCache.rank(a.item),
            (a, b) => (b.item?.ranking ?? 0) - (a.item?.ranking ?? 0),
        ]
        const fzf = new Fzf<SearchValue[]>(this.values, {
            selector: ({ text }) => text,
            limit: parameters.maxResults,
            match: extendedMatch,
            casing: 'case-insensitive',
            tiebreakers,
        })
        const isEmpty = query === ''
        const candidates = isEmpty
            ? sortByTiebreakers(
                  this.values
                      .filter(value => historyCache.rank(value))
                      .map(value => ({ item: value, positions: new Set(), start: 0, end: 0, score: 0 })),
                  tiebreakers
              )
            : fzf.find(query)
        const isComplete = candidates.length < parameters.maxResults
        candidates.slice(0, parameters.maxResults)

        const links: HighlightedLinkProps[] = candidates.map<HighlightedLinkProps>(candidate => {
            const positions = compressedRangePositions([...candidate.positions])
            const url = candidate.item.url || this?.params?.createURL?.(candidate.item.text)
            return {
                ...candidate.item,
                score: candidate.score,
                positions,
                url:
                    url && this?.params?.transformer?.modifyURL
                        ? this.params?.transformer.modifyURL(parameters.query, url)
                        : url,
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
