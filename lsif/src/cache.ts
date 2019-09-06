import { Connection, createConnection, EntityManager } from 'typeorm'
import { DocumentData } from './entities'
import { Id } from 'lsif-protocol'
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
        private disposeFunction: (value: V) => void
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
        }
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
        const newEntry = { key, promise, size: 0, readers: -1 }

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

                await this.removeNode(node, promise, size)
            }

            node = prev
        }
    }

    /**
     * Remove the given node from the list.
     *
     * @param node The node to remove.
     * @param promise The promise of the cache entry.
     * @param size The size of the promise value.
     */
    private async removeNode(node: Yallist.Node<CacheEntry<K, V>>, promise: Promise<V>, size: number): Promise<void> {
        this.size -= size
        this.lruList.removeNode(node)
        this.cache.delete(node.value.key)
        this.disposeFunction(await promise)
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
     */
    public withTransactionalEntityManager<T>(
        database: string,
        // Decorators are not possible type check
        // eslint-disable-next-line @typescript-eslint/ban-types
        entities: Function[],
        callback: (entityManager: EntityManager) => Promise<T>
    ): Promise<T> {
        return this.withConnection(database, entities, connection => connection.transaction(em => callback(em)))
    }
}

/**
 * A cache of deserialized `DocumentData` values indexed by their Identifer.
 */
export class DocumentCache extends GenericCache<Id, DocumentData> {
    /**
     * Create a new `DocumentCache` with the given maximum (soft) size for
     * all items in the cache.
     */
    constructor(max: number) {
        super(
            max,
            // TODO - determine memory size
            () => 1,
            // Let GC handle the cleanup of the object on cache eviction.
            (): void => {}
        )
    }

    /**
     * Invoke `callback` with document value obtained from the cache
     * cache or created on cache miss.
     *
     * @param documentId The identifier of the document.
     * @param factory The function used to create a document.
     * @param callback The function invoked with the document.
     */
    public withDocument<T>(
        documentId: Id,
        factory: () => Promise<DocumentData>,
        callback: (document: DocumentData) => Promise<T>
    ): Promise<T> {
        return this.withValue(documentId, factory, callback)
    }
}
