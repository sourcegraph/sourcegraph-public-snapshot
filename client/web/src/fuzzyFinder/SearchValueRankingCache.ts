import { SearchValue } from './SearchValue'

// Helper to cache the result of `SearchValue.historyRanking()`.  This
// optimization helps in large repositories where we sort a large number of
// files by their rank.
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
