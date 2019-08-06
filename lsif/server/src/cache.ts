import LRU from 'lru-cache'
import { Backend } from './backend'
import { Database } from './ms/database'
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
 * A `Database`, the size of the LSIF file it was loaded from, and a callback to
 * dispose of it when evicted from the cache.
 */
interface LRUDBEntry {
    dbPromise: Promise<Database>
    /**
     * The size of the underlying LSIF file. This directly contributes to the
     * size of the cache. Ideally, this would be set to the amount of memory
     * that the `Database` uses, but calculating the memory usage is difficult
     * so this uses the file size as a rough heuristic.
     */
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
            length: (entry, key) => entry.length,
            dispose: (key, entry) => entry.dispose(),
        })
    }

    /**
     * Runs the given `action` with the `Database` associated with the given
     * repository@commit. Internally, it either gets the `Database` from the LRU
     * cache or loads it from storage.
     */
    public async withDB<T>(backend: Backend, key: string, action: (db: Database) => Promise<T>): Promise<T> {
        let entry = this.lru.get(key)
        if (!entry) {
            const dbPromise = backend.loadDB(key)
            const length = 1 // TODO(efritz) - get length from backend
            const dispose = () => dbPromise.then(db => db.close())

            entry = { dbPromise, length, dispose }
            this.lru.set(key, entry)
        }

        return await action(await entry.dbPromise)
    }

    /**
     * Remove the entry associated with the given key from the cache.
     */
    public delete(key: string): void {
        this.lru.del(key)
    }
}
