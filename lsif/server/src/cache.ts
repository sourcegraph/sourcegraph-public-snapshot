import LRU from 'lru-cache'
import { Backend, QueryRunner } from './backend'
import { CacheStats, CreateRunnerStats, timeit } from './stats'
import { readEnvInt } from './env'

/**
 * Soft limit on the total amount of storage occupied by LSIF data loaded in
 * memory. The actual amount can exceed this if a single LSIF file is larger
 * than this limit, otherwise memory will be kept under this limit. Defaults to
 * 100MB.
 *
 * Empirically based on github.com/sourcegraph/codeintellify, each byte of
 * storage (uncompressed newline-delimited JSON) expands to 3 bytes in memory.
 */
const SOFT_MAX_STORAGE_IN_MEMORY = readEnvInt({
    key: 'LSIF_SOFT_MAX_STORAGE_IN_MEMORY',
    defaultValue: 100 * 1024 * 1024,
})

/**
 * The cache holds a promise that resolves to a query runner, some hint about
 * how many resources keeping this runner 'hot' costs, and a callback invoked
 * when the entry is evicted from memory.
 */
interface LRUDBEntry {
    promise: Promise<{ queryRunner: QueryRunner; createRunnerStats: CreateRunnerStats }>
    length: number
    dispose: () => void
}

/**
 * A wrapper around an LRU cache for Database values.
 */
export class Cache {
    private lru: LRU<string, LRUDBEntry>

    constructor() {
        this.lru = new LRU<string, LRUDBEntry>({
            max: SOFT_MAX_STORAGE_IN_MEMORY,
            length: (entry, _) => entry.length,
            dispose: (_, entry) => entry.dispose(),
        })
    }

    /**
     * Runs the given `action` with the `Database` associated with the given
     * repository@commit. Internally, it either gets a handle to the database
     * from the LRU cache or loads it from a secondary storage.
     */
    public async withDB<T, R extends QueryRunner>(
        backend: Backend<R>,
        repository: string,
        commit: string,
        action: (db: QueryRunner) => Promise<T>
    ): Promise<{ result: T; cacheStats: CacheStats; createRunnerStats?: CreateRunnerStats }> {
        const key = makeKey(repository, commit)

        let hit = true
        let entry = this.lru.get(key)

        if (!entry) {
            const promise = backend.createRunner(repository, commit)

            entry = {
                promise,
                length: 1, // TODO(efritz) - get length from backend
                dispose: () => promise.then(({ queryRunner }) => queryRunner.close()),
            }

            hit = false
            this.lru.set(key, entry)
        }

        const {
            result: { queryRunner, createRunnerStats },
            elapsed,
        } = await timeit(async () => {
            return await (<LRUDBEntry>entry).promise
        })

        // Wait for entry promise to resolve - will already
        // be resolved if this lookup was a cache hit.
        const result = await action(queryRunner)

        return {
            result,
            cacheStats: {
                cacheHit: hit,
                elapsedMs: elapsed,
            },
            // Only return createRunnerStats if it wasn't a cache hit
            createRunnerStats: (!hit || undefined) && createRunnerStats,
        }
    }

    /**
     * Remove the entry associated with the given key from the cache.
     */
    public delete(repository: string, commit: string): void {
        this.lru.del(makeKey(repository, commit))
    }

    /**
     * Remove all entries from the cache.
     */
    public reset(): void {
        this.lru.reset()
    }
}

/**
 * Computes a cache key from the given repository and commit hash.
 */
function makeKey(repository: string, commit: string): string {
    return `${repository}@${commit}`
}
