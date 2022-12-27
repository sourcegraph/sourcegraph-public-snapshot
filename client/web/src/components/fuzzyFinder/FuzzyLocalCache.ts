import { SymbolKind } from '../../graphql-operations'

/**
 * PersistableQueryResult must be safe to cache with JSON.stringify().
 *
 * It's OK to add new fields to this interface, but you must ensure that the
 * values are preserved with a `JSON.parse(JSON.stringify(value))` roundtrip.
 * For example, it's not OK to use JSX.Element.
 */
export interface PersistableQueryResult {
    text: string
    url?: string
    symbolKind?: SymbolKind
    stars?: number
    repoName?: string
    filePath?: string
}

export interface FuzzyLocalCache {
    initialValues(): Promise<PersistableQueryResult[]>
    staleValues: StaleValuesFunction
    cacheValues(values: PersistableQueryResult[]): void
}

type StaleValuesFunction = (cachedValues: PersistableQueryResult[]) => Promise<PersistableQueryResult[]>

export const emptyFuzzyCache: FuzzyLocalCache = {
    initialValues: () => Promise.resolve([]),
    staleValues: () => Promise.resolve([]),
    cacheValues: () => {},
}

/**
 * Implementation of `FuzzyLocalCache` that uses `Storage` such as `window.localStorage`
 */
export class FuzzyStorageCache implements FuzzyLocalCache {
    constructor(
        private readonly storage: Storage,
        private readonly cacheKey: string,
        public readonly staleValues: StaleValuesFunction
    ) {}
    public initialValues(): Promise<PersistableQueryResult[]> {
        const fromCache = this.storage.getItem(this.cacheKey) ?? '[]'
        try {
            return Promise.resolve(JSON.parse(fromCache) as PersistableQueryResult[])
        } catch (error) {
            return Promise.resolve([])
        }
    }
    public cacheValues(values: PersistableQueryResult[]): void {
        this.storage.setItem(this.cacheKey, JSON.stringify(values))
    }
}

/**
 * Implementation of `FuzzyLocalCache` that uses the browser's `caches` API.
 */
export class FuzzyWebCache implements FuzzyLocalCache {
    constructor(private readonly cacheKey: string, public readonly staleValues: StaleValuesFunction) {}
    public async initialValues(): Promise<PersistableQueryResult[]> {
        const cacheAvailable = 'caches' in self
        if (!cacheAvailable) {
            return []
        }
        const cache = await caches.open(this.cacheKey)
        const fromCache = await cache.match(new Request(this.cacheKey))
        if (!fromCache) {
            return []
        }
        return JSON.parse(await fromCache.text()) as PersistableQueryResult[]
    }
    public cacheValues(values: PersistableQueryResult[]): void {
        this.cacheValuesPromise(values).then(
            () => {},
            () => {}
        )
    }
    private async cacheValuesPromise(values: PersistableQueryResult[]): Promise<void> {
        const cacheAvailable = 'caches' in self
        if (!cacheAvailable) {
            return
        }
        const cache = await caches.open(this.cacheKey)
        await cache.put(new Request(this.cacheKey), new Response(JSON.stringify(values)))
    }
}
