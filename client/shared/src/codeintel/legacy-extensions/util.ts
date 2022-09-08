import stringify from 'fast-json-stable-stringify'
import LRU from 'lru-cache'

export const cache = <Arguments extends unknown[], V>(
    func: (...args: Arguments) => V,
    cacheOptions?: LRU.Options<string, V>
): ((...args: Arguments) => V) => {
    const lru = new LRU<string, V>(cacheOptions || { max: 500 })
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
