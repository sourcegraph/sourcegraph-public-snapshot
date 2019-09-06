import { Connection, createConnection, EntityManager } from 'typeorm'
import { DocumentData, ResultChunkData } from './entities'
import Yallist from 'yallist'

/**
 * A wrapper around a cache value promise.
 */
interface CacheEntry<K, V> {
    /**
     * The key that can retrieve this cache entry.
     */
    key: K

    /**
     * The promise that will resolve the cache value.
     */
    promise: Promise<V>

    /**
     * The size of the promise value, once resolved. This value is
     * initially zero and is updated once an appropriate can be
     * determined from the result of `promise`.
     */
    size: number

    /**
     * The number of active withValue calls referencing this entry.
     * If this value is non-zero, it should not be evict-able from the
     * cache.
     */
    readers: number

    /**
     * A function reference that should be called, if present, when
     * the reader count for an entry goes to zero. This will unblock a
     * a promise created in `bustKey` to wait for all readers to finish
     * using the cache value.
     */
    waiter: (() => void) | undefined
}

/**
 * A generic LRU cache. We use this instead of the `lru-cache` package
 * available in NPM so that we can handle async payloads in a more
 * first-class way as well as shedding some of the cruft around evictions.
 * We need to ensure database handles are closed when they are no longer
 * accessible, and we also do not want to evict any database handle while
 * it is actively being used.
 */
export class GenericCache<K, V> {
    /**
     * A map from from keys to nodes in `lruList`.
     */
    private cache = new Map<K, Yallist.Node<CacheEntry<K, V>>>()

    /**
     * A linked list of cache entires ordered by last-touch.
     */
    private lruList = new Yallist<CacheEntry<K, V>>()

    /**
     * The additive size of the items currently in the cache.
     */
    private size = 0

    /**
     * Create a new `GenericCache` with the given maximum (soft) size for
     * all items in the cache, a function that determine the size of a
     * cache item from its resolved value, and a function that is called
     * when an item falls out of the cache.
     */
    constructor(
        private max: number,
        private sizeFunction: (value: V) => number,
        private disposeFunction: (value: V) => Promise<void>
    ) {}

    /**
     * Check if `key` exists in the cache. If it does not, create a value
     * from `factory`. Once the cache value resolves, invoke `callback` and
     * return its value. This method acts as a lock around the cache entry
     * so that it may not be removed while the factory or callback functions
     * are running.
     *
     * @param key The cache key.
     * @param factory The function used to create a new value.
     * @param callback The function to invoke with the resolved cache value.
     */
    public async withValue<T>(key: K, factory: () => Promise<V>, callback: (value: V) => Promise<T>): Promise<T> {
        // Find or create the entry
        const entry = await this.getEntry(key, factory)

        // Increase the number of readers currently looking at this value.
        // While this value is not equal to zero, it wll not be skipped over
        // on the cache eviction pass.

        entry.readers++

        try {
            // Re-resolve the promise. If this is already resolved it's a fast
            // no-op. Otherwise, we got a cache entry that was under-construction
            // and will resolve shortly.

            return await callback(await entry.promise)
        } finally {
            // Unlock the cache entry
            entry.readers--

            // If we were the last reader and there's a bustKey call waiting on
            // us to finish, inform it that we're done using it. Bust away!

            if (entry.readers === 0 && entry.waiter !== undefined) {
                entry.waiter()
            }
        }
    }

    /**
     * Remove a key from the cache. This blocks until all current readers
     * of the cached value have completed, then calls the dispose function.
     *
     * Do NOT call this function while holding the same: you will deadlock.
     *
     * @param key The cache key.
     */
    public async bustKey(key: K): Promise<void> {
        const node = this.cache.get(key)
        if (!node) {
            return
        }

        const {
            value: { promise, size, readers },
        } = node

        // Immediately remove from cache so that another reader cannot get
        // ahold of the value, and so that another bust attempt cannot call
        // dispose twice on the same value.

        this.removeNode(node, size)

        // Wait for the value to resolve. We do this first in case the value
        // was still under construction. This simplifies the rest of the logic
        // below, as readers can never be negative once the promise value has
        // resolved.

        const value = await promise

        if (readers !== 0) {
            // There's someone holding the cache value. Create a barrier promise
            // and stash the function that can unlock it. When the reader count
            // for an entry is decremented, the waiter function, if present, is
            // invoked. This basically forms a condition variable.

            const { wait, done } = createBarrierPromise()
            node.value.waiter = done
            await wait
        }

        // We have the resolved value, removed from the cache, which is no longer
        // used by any reader. It's safe to dispose now.
        await this.disposeFunction(value)
    }

    /**
     * Check if `key` exists in the cache. If it does not, create a value
     * from `factory` and add it to the cache. In either case, update the
     * cache entry's place in `lruCache` and return the entry. If a new
     * value was created, then it may trigger a cache eviction once its
     * value resolves.
     *
     * @param key The cache key.
     * @param factory The function used to create a new value.
     */
    private async getEntry(key: K, factory: () => Promise<V>): Promise<CacheEntry<K, V>> {
        const node = this.cache.get(key)
        if (node) {
            // Found, move to head of list
            this.lruList.unshiftNode(node)
            return node.value
        }

        // Create promise and the entry that wraps it. We don't know
        // the effective size of the value until the promise resolves,
        // so we put zero. We have a reader count of -1, which is the
        // value that denotes that the cache entry is currently under
        // construction. We don't want to block here while waiting for
        // the promise value to resolve, otherwise a second request for
        // the same key will create a duplicate cache entry.

        const promise = factory()
        const newEntry = { key, promise, size: 0, readers: -1, waiter: undefined }

        // Add to head of list
        this.lruList.unshift(newEntry)

        // Grab the head of the list we just pushed and store it
        // in the map. We need the node that the unshift method
        // creates so we can unlink it in constant time.
        const head = this.lruList.head
        if (head) {
            this.cache.set(key, head)
        }

        // Now that another call to getEntry will find the cache entry
        // and early-out, we can block here and wait to resolve the
        // value, then update the entry and cache sizes.

        const value = await promise
        await this.resolved(newEntry, value)

        // Remove the under-construction value from the reader count.
        // Callers of this method end up incrementing this value again
        // a second time before calling a user callback function.
        newEntry.readers++

        return newEntry
    }

    /**
     * Determine the size of the resolved value and update the size of the
     * entry as well as `size`. While the total cache size exceeds `max`,
     * try to evict the least recently used cache entries that do not have
     * a non-zero `readers` count.
     *
     * @param entry The cache entry.
     * @param value The cache entry's resolved value.
     */
    private async resolved(entry: CacheEntry<K, V>, value: V): Promise<void> {
        const size = this.sizeFunction(value)
        this.size += size
        entry.size = size

        let node = this.lruList.tail
        while (this.size > this.max && node) {
            const {
                prev,
                value: { promise, size, readers },
            } = node

            if (readers === 0) {
                // If readers < 0, then we're under construction and we
                // don't have anything yet to discard. If readers > 0, then
                // it may be actively used by another part of the code that
                // hit a portion of their critical section that returned
                // control to the event loop. We don't want to mess with
                // those if we can help it.

                this.removeNode(node, size)
                await this.disposeFunction(await promise)
            }

            node = prev
        }
    }

    /**
     * Remove the given node from the list and update the cache size.
     *
     * @param node The node to remove.
     * @param size The size of the promise value.
     */
    private removeNode(node: Yallist.Node<CacheEntry<K, V>>, size: number): void {
        this.size -= size
        this.lruList.removeNode(node)
        this.cache.delete(node.value.key)
    }
}

/**
 * A cache of SQLite database connections indexed by database filenames.
 */
export class ConnectionCache extends GenericCache<string, Connection> {
    /**
     * Create a new `ConnectionCache` with the given maximum (soft) size for
     * all items in the cache.
     */
    constructor(max: number) {
        super(
            max,
            // Each handle is roughly the same size.
            () => 1,
            // Close the underlying file handle on cache eviction.
            (connection: Connection) => connection.close()
        )
    }

    /**
     * Invoke `callback` with a SQLite connection object obtained from the
     * cache or created on cache miss. This connection is guaranteed not to
     * be disposed by cache eviction while the callback is active.
     *
     * @param database The database filename.
     * @param entities The set of expected entities present in this schema.
     * @param callback The function invoke with the SQLite connection.
     */
    public withConnection<T>(
        database: string,
        // Decorators are not possible type check
        // eslint-disable-next-line @typescript-eslint/ban-types
        entities: Function[],
        callback: (connection: Connection) => Promise<T>
    ): Promise<T> {
        const factory = (): Promise<Connection> =>
            createConnection({
                database,
                entities,
                type: 'sqlite',
                name: database,
                synchronize: true,
                // logging: 'all',
            })

        return this.withValue(database, factory, callback)
    }

    /**
     * Like `withConnection`, but will open a transaction on the connection
     * before invoking the callback.
     *
     * @param database The database filename.
     * @param entities The set of expected entities present in this schema.
     * @param callback The function invoke with a SQLite transaction connection.
     * @param pragmaHook The function called with connection before the transaction begins.
     */
    public withTransactionalEntityManager<T>(
        database: string,
        // Decorators are not possible type check
        // eslint-disable-next-line @typescript-eslint/ban-types
        entities: Function[],
        callback: (entityManager: EntityManager) => Promise<T>,
        pragmaHook?: (connection: Connection) => Promise<void>
    ): Promise<T> {
        return this.withConnection(database, entities, async connection => {
            if (pragmaHook) {
                await pragmaHook(connection)
            }

            return await connection.transaction(em => callback(em))
        })
    }
}

/**
 * A wrapper around a cache value that retains its encoded size. In order to keep
 * the in-memory limit of these decoded items, we use this value as the cache entry
 * size. This assumes that the size of the encoded text is a good proxy for the size
 * of the in-memory representation.
 */
export interface EncodedJsonCacheValue<T> {
    /**
     * The size of the encoded value.
     */
    size: number

    /**
     * The decoded value.
     */
    data: T
}

/**
 * A cache of decoded values encoded as JSON and gzipped in a SQLite database.
 */
class EncodedJsonCache<K, V> extends GenericCache<K, EncodedJsonCacheValue<V>> {
    /**
     * Create a new `EncodedJsonCache` with the given maximum (soft) size for
     * all items in the cache.
     */
    constructor(max: number) {
        super(
            max,
            v => v.size,
            // Let GC handle the cleanup of the object on cache eviction.
            (): Promise<void> => Promise.resolve()
        )
    }
}

/**
 * A cache of deserialized `DocumentData` values indexed by a string containing
 * the database path and the path of the document.
 */
export class DocumentCache extends EncodedJsonCache<string, DocumentData> {
    /**
     * Create a new `DocumentCache` with the given maximum (soft) size for
     * all items in the cache.
     */
    constructor(max: number) {
        super(max)
    }
}

/**
 * A cache of deserialized `ResultChunkData` values indexed by a string containing
 * the database path and the chunk index.
 */
export class ResultChunkCache extends EncodedJsonCache<string, ResultChunkData> {
    /**
     * Create a new `ResultChunkCache` with the given maximum (soft) size for
     * all items in the cache.
     */
    constructor(max: number) {
        super(max)
    }
}

/**
 * Return a promise and a function pair. The promise resolves once the function is called.
 */
export function createBarrierPromise(): { wait: Promise<void>; done: () => void } {
    let done!: () => void
    const wait = new Promise<void>(resolve => (done = resolve))
    return { wait, done }
}
