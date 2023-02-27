/**
 * This class implements a Map like interface. However for every item that is
 * stored it also stores the time on when this item was last accessed. Whenever
 * new items are added, we check if the cache contains items that are not
 * accessed for a certain amount of time. If so, we remove them from the cache.
 */
export class ExpirationCache<K, V> {
    private expirationTime: number
    private map: Map<K, { value: V; lastAccessAt: number }>

    constructor(expirationTime: number) {
        this.expirationTime = expirationTime
        this.map = new Map()
    }

    public set(key: K, value: V): void {
        this.map.set(key, { value, lastAccessAt: Date.now() })
        this.clean()
    }

    public has(key: K): boolean {
        const entry = this.map.get(key)
        if (entry !== undefined) {
            entry.lastAccessAt = Date.now()
        }
        return entry !== undefined
    }

    public get(key: K): V | undefined {
        const entry = this.map.get(key)
        if (entry === undefined) {
            return undefined
        }
        entry.lastAccessAt = Date.now()
        return entry.value
    }

    public delete(key: K): void {
        this.map.delete(key)
    }

    private clean(): void {
        const now = Date.now()
        for (const [key, entry] of this.map) {
            if (now - entry.lastAccessAt > this.expirationTime) {
                this.map.delete(key)
            }
        }
    }
}
