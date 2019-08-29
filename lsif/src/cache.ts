import promClient from 'prom-client'
import Yallist from 'yallist'
import { Connection, createConnection, EntityManager } from 'typeorm'
import { DocumentData } from './entities'
import { Id } from 'lsif-protocol'
import {
    ConnectionCacheEvictionCounter,
    ConnectionCacheHitCounter,
    ConnectionCacheSizeGauge,
    DocumentCacheEvictionCounter,
    DocumentCacheHitCounter,
    DocumentCacheSizeGauge,
} from './metrics'

/**
 * A wrapper around a cache value promise.
 */
interface CacheEntry<K, V> {
    // The key that can retrieve this cache entry.
    key: K

    // The promise that will resolve the cache value.
    promise: Promise<V>

    /**
     * The size of the promise value, once resolved. This value is
     * initially zero and is updated once an appropriate can be
     * determined from the result of `promise`.
     */
    size: number

    /**
     * The number of active withValue calls referencing this entry.
     * If this value is non-zero, it should not be evictable from the
     * cache.
     */
    readers: number
}

/**
 * A bag of prometheus metric objects that apply to a particular
 * instance of `GenericCache`.
 */
interface CacheMetrics {
    // A metric incremented on each cache hit or cache miss.
    hitCounter: promClient.Counter

    /**
     * A metric  incremented on each cache eviction, and a special label
     * is applied if an entry cannot be evicted because it is currently
     * being used.
     */
    evictionCounter: promClient.Counter

    /**
     * A metric incremented on each cache insertion and decremented on
     * each cache eviction.
     */
    sizeGauge: promClient.Gauge
}

/**
 * A generic LRU cache. We use this instead of the `lru-cache` apckage
 * available in NPM so that we can handle async payloads in a more
 * first-class way as well as shedding some of the cruft around evictions.
 * We need to ensure database handles are closed when they are no longer
 * accessible, and we also do not want to evict any database handle while
 * it is actively being used.
 */
class GenericCache<K, V> {
    // A map from from keys to nodes in `lruList`.
    private cache = new Map<K, Yallist.Node<CacheEntry<K, V>>>()

    // A linked list of cache entires ordered by last-touch.
    private lruList = new Yallist<CacheEntry<K, V>>()

    // The additive size of the items currently in the cache.
    private size = 0

    /**
     * Create a new `GenericCache` with the given maximum (soft) size for
     * all items in the cache, a function that determine the size of a
     * cache item from its resolved value, and a function that is called
     * when an item falls out of the cache.
     *
     * @param max The maximum size of the cache before an eviction.
     * @param sizeFunction A function that determines the size of a cache item.
     * @param disposeFunction A function that disposes of evicted cache items.
     * @param metrics The bag of metrics to use for this instance of the cache.
     */
    constructor(
        private max: number,
        private sizeFunction: (value: V) => number,
        private disposeFunction: (value: V) => void,
        private metrics: CacheMetrics
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
    protected async withValue<T>(key: K, factory: () => Promise<V>, callback: (value: V) => Promise<T>): Promise<T> {
        const entry = this.getEntry(key, factory)
        entry.readers++

        try {
            return await callback(await entry.promise)
        } finally {
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
    private getEntry(key: K, factory: () => Promise<V>): CacheEntry<K, V> {
        const node = this.cache.get(key)
        if (node) {
            this.metrics.hitCounter.labels('hit').inc()
            this.lruList.unshiftNode(node)
            return node.value
        }

        this.metrics.hitCounter.labels('miss').inc()
        const promise = factory()
        const newEntry = { key, promise, size: 0, readers: 0 }
        promise.then(value => this.resolved(newEntry, value), () => {})
        this.lruList.unshift(newEntry)
        const head = this.lruList.head
        if (head) {
            this.cache.set(key, head)
        }

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
    private resolved(entry: CacheEntry<K, V>, value: V): void {
        const size = this.sizeFunction(value)
        this.size += size
        entry.size = size
        this.metrics.sizeGauge.inc(size)

        let node = this.lruList.tail
        while (this.size > this.max && node) {
            const {
                prev,
                value: { promise, size, readers },
            } = node

            if (readers === 0) {
                this.size -= size
                this.metrics.evictionCounter.labels('evict').inc()
                this.metrics.sizeGauge.dec(size)
                this.lruList.removeNode(node)
                this.cache.delete(node.value.key)
                promise.then(value => this.disposeFunction(value), () => {})
            } else {
                this.metrics.evictionCounter.labels('locked').inc()
            }

            node = prev
        }
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
            (connection: Connection) => connection.close(),
            {
                hitCounter: ConnectionCacheHitCounter,
                sizeGauge: ConnectionCacheSizeGauge,
                evictionCounter: ConnectionCacheEvictionCounter,
            }
        )
    }

    /**
     * Invoke `callback` with a SQLite connection object obtained from the
     * cache or created on cache miss. This connection is guranteed not to
     * be disposed by cache eviction while the callback is active.
     *
     * @param database The database filename.
     * @param entities The set of expected entities present in this schema.
     * @param callback The function invoke with the SQLite connection.
     */
    public withConnection<T>(
        database: string,
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
                logging: ['error', 'warn'],
                maxQueryExecutionTime: 1000,
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
            (): void => {},
            {
                hitCounter: DocumentCacheHitCounter,
                sizeGauge: DocumentCacheSizeGauge,
                evictionCounter: DocumentCacheEvictionCounter,
            }
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
