interface CacheItem {
    createdAt: number
    result: FetchCacheReturnType<any>
}

export interface FetchCacheReturnType<T> {
    data: T
    status: number
    headers: Record<string, string>
}

const cache = new Map<string, CacheItem>()
// NOTE: try to use single promises cache
const fetchesInProgress = new Map<string, Promise<FetchCacheReturnType<any>>>()

let isEnabled = true

const fetchData = <T>(url: string, requestInit: RequestInit): Promise<FetchCacheReturnType<T>> =>
    fetch(url, requestInit).then(async response => {
        const data = (await response.json()) as T
        return { status: response.status, data, headers: Object.fromEntries(response.headers.entries()) }
    })

/**
 * fetch API with cache
 *
 * @param maxAge maximum allowed age in milliseconds for cached result
 * @param args fetch API parameters
 * @description Caches same argument requests for 1 minute
 */
export const fetchCache = async <T = any>({
    cacheMaxAge,
    url,
    ...requestInit
}: RequestInit & { url: string; cacheMaxAge: number }): Promise<FetchCacheReturnType<T>> => {
    if (!isEnabled || cacheMaxAge <= 0) {
        return fetchData<T>(url, requestInit)
    }

    const key = JSON.stringify({ url, ...requestInit })
    const minCreatedAt = Date.now() - cacheMaxAge

    if (cache.has(key) && cache.get(key)!.createdAt >= minCreatedAt) {
        return cache.get(key)!.result as FetchCacheReturnType<T>
    }

    if (fetchesInProgress.has(key)) {
        return (await fetchesInProgress.get(key)) as FetchCacheReturnType<T>
    }

    const request = fetchData<T>(url, requestInit)
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
export const enableFetchCache = (): void => {
    isEnabled = true
}

export const disableFetchCache = (): void => {
    isEnabled = false
}

export const clearFetchCache = (): void => cache.clear()
