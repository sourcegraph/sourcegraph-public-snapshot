interface CacheItem {
    createdAt: number
    result: FetchCacheReturnType<any>
}

export interface FetchCacheReturnType<T> {
    data: T
    status: number
    headers: Headers
}

const cache = new Map<string, CacheItem>()
const fetchesInProgress = new Map<string, Promise<FetchCacheReturnType<any>>>()

let isEnabled = true

/**
 * fetch API with cache
 *
 * @param maxAge maximum allowed age in milliseconds for cached result
 * @param args fetch API parameters
 * @description Caches same argument requests for 1 minute
 */
export const fetchCache = async <T = any>(
    maxAge: number,
    ...args: Parameters<typeof fetch>
): Promise<FetchCacheReturnType<T>> => {
    const fetchData = (): Promise<FetchCacheReturnType<T>> =>
        fetch(...args).then(async response => {
            const data = (await response.json()) as T
            return { status: response.status, data, headers: response.headers }
        })

    if (!isEnabled || maxAge <= 0) {
        return fetchData()
    }

    const key = JSON.stringify(args)
    const minCreatedAt = Date.now() - maxAge
    if (cache.has(key) && cache.get(key)!.createdAt > minCreatedAt) {
        return cache.get(key)!.result as FetchCacheReturnType<T>
    }

    if (fetchesInProgress.has(key)) {
        return (await fetchesInProgress.get(key)) as FetchCacheReturnType<T>
    }

    const request = fetchData()
        .then(result => {
            cache.set(key, { result, createdAt: Date.now() })
            return result
        })
        .finally(() => fetchesInProgress.delete(key))

    fetchesInProgress.set(key, request)

    return request
}
/**
 * For unit testing purposes
 */
fetchCache.enable = (): void => {
    isEnabled = true
}

fetchCache.disable = (): void => {
    isEnabled = false
}

fetchCache.clear = (): void => cache.clear()
