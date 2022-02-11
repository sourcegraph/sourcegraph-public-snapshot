/**
 * Creates a function that memoizes the async result of func. If the Promise is rejected, the result will not be
 * cached.
 *
 * @param func The function to memoize
 * @param toKey Determines the cache key for storing the result based on the first argument provided to the memoized
 * function
 */
export function memoizeAsync<P, T>(
    func: (parameters: P) => Promise<T>,
    toKey: (parameters: P) => string
): (parameters: P) => Promise<T> {
    const valuesCache = new Map<string, T>()
    const requestCache = new Map<string, Promise<T>>()

    return (parameters: P) => {
        const key = toKey(parameters)

        const valueHit = valuesCache.get(key)
        const requestHit = requestCache.get(key)

        if (valueHit) {
            return Promise.resolve(valueHit)
        }

        if (requestHit) {
            return requestHit
        }

        const promise = func(parameters)

        promise
            .then(result => {
                requestCache.delete(key)
                valuesCache.set(key, result)

                return result
            })
            .catch(error => {
                requestCache.delete(key)

                // Re-throw error for consumers reject and catch handlers
                throw error
            })

        requestCache.set(key, promise)

        return promise
    }
}
