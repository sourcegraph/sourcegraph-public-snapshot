import { FuzzySearch, type FuzzySearchParameters, type FuzzySearchResult } from './FuzzySearch'

export class AggregateFuzzySearch extends FuzzySearch {
    public totalFileCount: number
    constructor(public readonly underlying: FuzzySearch[]) {
        super()
        this.totalFileCount =
            underlying.length === 0 ? 0 : underlying.map(search => search.totalFileCount).reduce((a, b) => a + b)
    }
    public search(parameters: FuzzySearchParameters): FuzzySearchResult {
        const result: FuzzySearchResult = {
            links: [],
            isComplete: true,
        }
        for (const search of this.underlying) {
            const searchResult = search.search(parameters)
            result.links.push(...searchResult.links)
            result.isComplete = result.isComplete && searchResult.isComplete
        }
        // Sort aggregated results based on fuzzy score.
        result.links.sort((a, b) => (b.score ?? 0) - (a.score ?? 0))
        return result
    }
}
