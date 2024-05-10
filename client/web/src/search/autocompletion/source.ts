export interface CompletionResult<T> {
    /**
     * Initial/synchronous result.
     */
    result: T[]
    /**
     * Function to be called to load additional results if necessary.
     */
    next: () => Promise<{ result: T[] }>
}

interface CachedAsyncCompletionSourceConfig<QueryResult, FilterResult, ExtraArgs extends any[] = []> {
    /**
     * Returns a string that uniquely identifies this query (which is often just
     * the query itself). If the same request is made again the existing result
     * is reused.
     */
    queryKey(value: string, dataCacheKey?: string): string
    /**
     * Fetch data. queryKey is the value return by the queryKey function and
     * value is the term that's currently completed. Returns a list of [key,
     * value] tuples. The key of these tuples is used to uniquly identify a
     * value the data cache.
     */
    query(queryKey: string, value: string): Promise<[string, QueryResult][]>
    /**
     * This function filters and ranks all cache values (entries) by value.
     */
    filter(entries: QueryResult[], value: string): FilterResult[]
    /**
     * If provided data values are bucketed into different "cache groups", keyed
     * by the return value of this function.
     */
    dataCacheKey?(...extraArgs: ExtraArgs): string
}

/**
 * This class handles creating suggestion results that include cached values (if
 * available) and updates the cache with new results from new queries.
 */
export class CachedAsyncCompletionSource<QueryResult, FilterResult, ExtraArgs extends any[] = []> {
    private queryCache = new Map<string, Promise<void>>()
    private dataCache = new Map<string, QueryResult>()
    private dataCacheByQuery = new Map<string, Map<string, QueryResult>>()

    constructor(private config: CachedAsyncCompletionSourceConfig<QueryResult, FilterResult, ExtraArgs>) {}

    public query<MappedResult>(
        value: string,
        mapper: (values: FilterResult[]) => MappedResult[],
        ...extraArgs: ExtraArgs
    ): CompletionResult<MappedResult> {
        // The dataCacheKey could possibly just be an argument to query. However
        // that would require callsites to remember to pass the value. Doing it
        // this way we get a bit more type safety.
        const dataCacheKey = this.config.dataCacheKey?.(...extraArgs)
        const queryKey = this.config.queryKey(value, dataCacheKey)
        let dataCache = this.dataCache
        if (dataCacheKey) {
            dataCache = this.dataCacheByQuery.get(dataCacheKey) ?? new Map<string, QueryResult>()
            if (!this.dataCacheByQuery.has(dataCacheKey)) {
                this.dataCacheByQuery.set(dataCacheKey, dataCache)
            }
        }
        return {
            result: mapper(this.cachedData(value, dataCache)),
            next: () => {
                let result = this.queryCache.get(queryKey)

                if (!result) {
                    result = this.config.query(queryKey, value).then(entries => {
                        for (const [key, entry] of entries) {
                            if (!dataCache.has(key)) {
                                dataCache.set(key, entry)
                            }
                        }
                    })

                    this.queryCache.set(queryKey, result)
                }

                return result.then(() => ({ result: mapper(this.cachedData(value, dataCache)) }))
            },
        }
    }

    private cachedData(value: string, cache = this.dataCache): FilterResult[] {
        return this.config.filter(Array.from(cache.values()), value)
    }
}
