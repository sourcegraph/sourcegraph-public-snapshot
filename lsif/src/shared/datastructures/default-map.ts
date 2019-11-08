/**
 * An extension of `Map` that defines `getOrDefault` for a type of stunted
 * autovivification. This saves a bunch of code that needs to check if a
 * nested type within a map is undefined on first access.
 */
export class DefaultMap<K, V> extends Map<K, V> {
    /**
     * Returns a new `DefaultMap`.
     *
     * @param defaultFactory The factory invoked when an undefined value is accessed.
     */
    constructor(private defaultFactory: () => V) {
        super()
    }

    /**
     * Get a key from the map. If the key does not exist, the default factory produces
     * a value and inserted into the map before being returned.
     *
     * @param key The key to retrieve.
     */
    public getOrDefault(key: K): V {
        let value = super.get(key)
        if (value !== undefined) {
            return value
        }

        value = this.defaultFactory()
        this.set(key, value)
        return value
    }
}
