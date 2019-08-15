import { Connection, createConnection } from 'typeorm'
import { DocumentBlob } from './database'
import { Id } from 'lsif-protocol'
import Yallist from 'yallist'

interface CacheEntry<K, V> {
    key: K
    promise: Promise<V>
    size: number
    readers: number
}

export class GenericCache<K, V> {
    private cache = new Map<K, Yallist.Node<CacheEntry<K, V>>>()
    private lruList = new Yallist<CacheEntry<K, V>>()
    private size = 0

    constructor(
        private max: number,
        private sizeFunction: (value: V) => number,
        private disposeFunction: (value: V) => void
    ) {}

    protected async withValue<T>(key: K, factory: () => Promise<V>, callback: (value: V) => Promise<T>): Promise<T> {
        const entry = this.getEntry(key, factory)
        entry.readers++

        try {
            return callback(await entry.promise)
        } finally {
            entry.readers--
        }
    }

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

    private resolved(entry: CacheEntry<K, V>, value: V): void {
        const size = this.sizeFunction(value)
        this.size += size
        this.cache.get(entry.key)
        entry.size = size
        this.trim()
    }

    private trim(): void {
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

export class ConnectionCache extends GenericCache<string, Connection> {
    constructor(max: number) {
        super(max, ConnectionCache.sizeFunction, ConnectionCache.connectionFunction)
    }

    private static sizeFunction = (_: Connection) => 1
    private static connectionFunction = (connection: Connection) => connection.close()

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
}

export class BlobCache extends GenericCache<Id, DocumentBlob> {
    constructor(max: number) {
        super(max, BlobCache.sizeFunction, BlobCache.disposeFunction)
    }

    private static sizeFunction = (_: DocumentBlob) => 1
    private static disposeFunction = (_: DocumentBlob): void => {}

    public withBlob<T>(
        documentId: Id,
        factory: () => Promise<DocumentBlob>,
        callback: (blob: DocumentBlob) => Promise<T>
    ): Promise<T> {
        return this.withValue(documentId, factory, callback)
    }
}

export const connectionCache = new ConnectionCache(5)
export const blobCache = new BlobCache(25)
