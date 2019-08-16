import { Connection, createConnection } from 'typeorm'
import { DocumentBlob } from './models'
import { Id } from 'lsif-protocol'
import Yallist from 'yallist'

/**
 * `CacheEntry` is a wrapper around a cache value promise.
 */
interface CacheEntry<K, V> {
    /**
     * `key` is the key that can retrieve this cache entry.
     */
    key: K

    /**
     * `promise` is the promise that will resolve the cache value.
     */
    promise: Promise<V>

    /**
     * `size` is the size of the promise value, once resolved. This
     * value is initially zero and is updated once an appropriate can
     * be determined from the result of `promise`.
     */
    size: number

    /**
     * `reader` is the number of active withValue calls referencing
     * this entry. If this value is non-zero, it should not be evictable
     * from the cache.
     */
    readers: number
}

/**
 * `GenericCache` is a generic LRU cache.
 */
class GenericCache<K, V> {
    /**
     * `cache` is a map from from keys to nodes in `lruList`.
     */
    private cache = new Map<K, Yallist.Node<CacheEntry<K, V>>>()

    /**
     * `lruList` is a linked list of cache entires ordered by last-touch.
     */
    private lruList = new Yallist<CacheEntry<K, V>>()

    /**
     * `size` is the additive size of the items currently in the cache.
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
            this.lruList.unshiftNode(node)
            return node.value
        }

        const promise = factory()
        const newEntry = { key: key, promise, size: 0, readers: 0 }
        promise.then(value => this.resolved(newEntry, value))
        this.lruList.unshift(newEntry)
        this.cache.set(key, this.lruList.head!)
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

        let node = this.lruList.tail
        while (this.size > this.max && node) {
            const {
                prev,
                value: { promise, size, readers },
            } = node

            if (readers === 0) {
                this.size -= size
                this.lruList.removeNode(node)
                this.cache.delete(node.value.key)
                promise.then(value => this.disposeFunction(value))
            }

            node = prev
        }
    }
}

/**
 * `ConnectionCache` is a cache of SQLite database connections indexed
 * by database filenames.
 */
export class ConnectionCache extends GenericCache<string, Connection> {
    /**
     * Create a new `ConnectionCache` with the given maximum (soft) size for
     * all items in the cache.
     */
    constructor(max: number) {
        super(max, ConnectionCache.sizeFunction, ConnectionCache.connectionFunction)
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
        entities: any[],
        callback: (connection: Connection) => Promise<T>
    ): Promise<T> {
        const factory = () =>
            createConnection({
                type: 'sqlite',
                name: database,
                database: database,
                entities: entities,
                synchronize: true,
                // logging: 'all',
            })

        return this.withValue(database, factory, callback)
    }

    // Each handle is roughly the same size.
    private static sizeFunction = (_: Connection) => 1

    // Close the underlying file handle on cache eviction.
    private static connectionFunction = (connection: Connection) => connection.close()
}

/**
 * `BlobCache` is a cache of deserialized `DocumentBlob` values indexed
 * by their Identifer.
 */
export class BlobCache extends GenericCache<Id, DocumentBlob> {
    /**
     * Create a new `BlobCache` with the given maximum (soft) size for
     * all items in the cache.
     */
    constructor(max: number) {
        super(max, BlobCache.sizeFunction, BlobCache.disposeFunction)
    }

    /**
     * Invoke `callback` with document value obtained from the cache
     * cache or created on cache miss.
     *
     * @param documentId The identifier of the document.
     * @param factory The function used to create a document.
     * @param callback The function invoked with the document.
     */
    public withBlob<T>(
        documentId: Id,
        factory: () => Promise<DocumentBlob>,
        callback: (blob: DocumentBlob) => Promise<T>
    ): Promise<T> {
        return this.withValue(documentId, factory, callback)
    }

    // TODO - determine memory size
    private static sizeFunction = (_: DocumentBlob) => 1

    // Let GC handle the cleanup of the object on cache eviction.
    private static disposeFunction = (_: DocumentBlob): void => {}
}

// TODO - make non-globals
export const connectionCache = new ConnectionCache(5)
export const blobCache = new BlobCache(25)
