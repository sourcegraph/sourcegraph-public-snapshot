import type { SearchValue } from './SearchValue'

// Helper to cache the result of `SearchValue.historyRanking()`, which is both
// useful for 1) good UX to keep the ranking stable as long as the query is
// unchanged and 2) performance in large repositories where we sort a large
// number of files by their rank score.
export class SearchValueRankingCache {
    private cache: Map<SearchValue, number> = new Map()
    public rank(value: SearchValue): number {
        if (value.historyRanking === undefined) {
            return 0
        }
        const fromCache = this.cache.get(value)
        if (fromCache !== undefined) {
            return fromCache
        }
        const ranking = value.historyRanking() ?? 0
        this.cache.set(value, ranking)
        return ranking
    }
}
