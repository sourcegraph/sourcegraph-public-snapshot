import { CaseInsensitiveFuzzySearch } from '../../fuzzyFinder/CaseInsensitiveFuzzySearch'
import type {
    FuzzySearch,
    FuzzySearchConstructorParameters,
    IndexingFSM,
    SearchIndexing,
} from '../../fuzzyFinder/FuzzySearch'
import type { SearchValue } from '../../fuzzyFinder/SearchValue'

import type { FuzzyFSM } from './FuzzyFsm'
import type { FuzzyLocalCache, PersistableQueryResult } from './FuzzyLocalCache'

export abstract class FuzzyQuery {
    private isStaleResultsDeleted = false
    protected queries: Map<string, Promise<PersistableQueryResult[]>> = new Map()
    protected doneQueries: Set<string> = new Set()
    protected queryResults: Map<string, PersistableQueryResult> = new Map()

    constructor(
        private readonly onNamesChanged: () => void,
        private readonly cache: FuzzyLocalCache,
        private readonly fuzzySearchParams?: FuzzySearchConstructorParameters
    ) {
        this.addQueryResults(this.cache.initialValues())
    }

    protected abstract searchValues(): SearchValue[]
    protected abstract rawQuery(userQuery: string): string
    protected abstract handleRawQueryPromise(query: string): Promise<PersistableQueryResult[]>

    public async removeStaleResults(): Promise<void> {
        const fromCache = this.cache.initialValues()
        if (fromCache.length === 0) {
            // Nothing to invalidate.
            return
        }
        const toRemove = await this.cache.staleValues(fromCache)
        this.removeQueryResults(toRemove)
    }
    private removeQueryResults(toRemove: PersistableQueryResult[]): void {
        const oldSize = this.queryResults.size
        for (const result of toRemove) {
            this.queryResults.delete(result.url || result.text)
        }
        const didChange = this.queryResults.size < oldSize
        if (didChange) {
            const newValues = [...this.queryResults.values()]
            this.cache.cacheValues(newValues)
            this.onNamesChanged()
        }
    }
    private addQueryResults(results: PersistableQueryResult[]): void {
        const oldSize = this.queryResults.size
        for (const result of results) {
            this.queryResults.set(result.url || result.text, result)
        }
        const didChangeSize = this.queryResults.size > oldSize
        if (didChangeSize) {
            this.cache.cacheValues([...this.queryResults.values()])
            this.onNamesChanged()
        }
    }
    public isDoneDownloading(): boolean {
        return this.queries.size === 0
    }
    public isDownloading(): boolean {
        return this.queries.size > 0
    }
    public hasQuery(query: string): boolean {
        return this.queries.has(query) || this.doneQueries.has(query)
    }
    protected fuzzySearch(): FuzzySearch {
        return new CaseInsensitiveFuzzySearch(this.searchValues(), this.fuzzySearchParams)
    }

    public fuzzyFSM(): FuzzyFSM {
        if (this.isDownloading()) {
            return {
                key: 'downloading',
                downloading: this.indexingFSM(),
            }
        }
        return {
            key: 'ready',
            fuzzy: this.fuzzySearch(),
        }
    }

    public indexingFSM(): SearchIndexing {
        let indexingPromise: Promise<IndexingFSM> | undefined
        return {
            key: 'indexing',
            partialFuzzy: this.fuzzySearch(),
            indexedFileCount: this.queryResults.size,
            totalFileCount: this.queryResults.size + 10,
            isIndexing: () => indexingPromise !== undefined,
            continueIndexing: () => {
                if (!indexingPromise) {
                    indexingPromise = Promise.any([...this.queries.values()]).then(
                        () =>
                            this.isDoneDownloading() ? { key: 'ready', value: this.fuzzySearch() } : this.indexingFSM(),
                        () => this.indexingFSM()
                    )
                }
                return indexingPromise
            },
        }
    }

    public handleQuery(query: string): void {
        if (query === '') {
            return
        }
        const actualQuery = this.rawQuery(query)
        if (this.hasQuery(actualQuery)) {
            return
        }
        this.addQuery(actualQuery, this.handleRawQueryPromise(actualQuery))

        if (!this.isStaleResultsDeleted) {
            this.isStaleResultsDeleted = true
            this.removeStaleResults().then(
                () => {},
                () => {}
            )
        }
    }

    public addQuery(query: string, promise: Promise<PersistableQueryResult[]>): void {
        this.queries.set(query, promise)
        promise
            .then(
                result => this.addQueryResults(result),
                // eslint-disable-next-line no-console
                error => console.error(`failed to download results for query ${query}`, error)
            )
            .finally(() => {
                this.queries.delete(query)
                this.doneQueries.add(query)
                this.onNamesChanged()
            })
        return
    }
}
