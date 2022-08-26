import stringify from 'fast-json-stable-stringify'
import LRU from 'lru-cache'

export const cache = <Arguments extends unknown[], V>(
    func: (...args: Arguments) => V,
    cacheOptions?: LRU.Options<string, V>
): ((...args: Arguments) => V) => {
    // All the other options are optional, see the sections below for
    // documentation on what each one does.  Most of them can be
    // overridden for specific items in get()/set()
    const defaultOptions: LRU.Options<string, V> = {
        max: 500,
        maxSize: 5000,
        ttl: 1000 * 60 * 5,
    }
    const lru = new LRU<string, V>(cacheOptions || defaultOptions)
    return (...args) => {
        const key = stringify(args)
        if (lru.has(key)) {
            // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
            return lru.get(key)!
        }
        const value = func(...args)
        lru.set(key, value)
        return value
    }
}
