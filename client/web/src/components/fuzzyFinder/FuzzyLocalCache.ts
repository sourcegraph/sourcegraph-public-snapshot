import type { SymbolKind } from '../../graphql-operations'

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
    symbolName?: string
    stars?: number
    repoName?: string
    filePath?: string
}

export interface FuzzyLocalCache {
    initialValues(): PersistableQueryResult[]
    staleValues: StaleValuesFunction
    cacheValues(values: PersistableQueryResult[]): void
}

type StaleValuesFunction = (cachedValues: PersistableQueryResult[]) => Promise<PersistableQueryResult[]>

export const emptyFuzzyCache: FuzzyLocalCache = {
    initialValues: () => [],
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
    public initialValues(): PersistableQueryResult[] {
        const fromCache = this.storage.getItem(this.cacheKey) ?? '[]'
        try {
            return JSON.parse(fromCache) as PersistableQueryResult[]
        } catch {
            return []
        }
    }
    public cacheValues(values: PersistableQueryResult[]): void {
        this.storage.setItem(this.cacheKey, JSON.stringify(values))
    }
}
