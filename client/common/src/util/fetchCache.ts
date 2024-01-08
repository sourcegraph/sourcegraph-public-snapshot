interface CacheItem {
    createdAt: number
    request: Promise<FetchCacheResponse<any>>
}

export interface FetchCacheResponse<T> {
    data?: T
    status: number
    headers: Record<string, string>
}

export interface FetchCacheRequest extends RequestInit {
    /**
     * URL to fetch
     */
    url: string
    /**
     * maximum allowed cached item age in milliseconds
     * @example "60000" will allow using cached item within last minute
     */
    cacheMaxAge: number
}

const cache = new Map<string, CacheItem>()
let isEnabled = true

/**
 * fetch API with cache
 * @description Caches same argument requests for 1 minute
 */
export const fetchCache = async <T = any>({
    cacheMaxAge,
    url,
    ...requestInit
}: FetchCacheRequest): Promise<FetchCacheResponse<T>> => {
    const fetchData = (): Promise<FetchCacheResponse<T>> =>
        fetch(url, requestInit).then(async response => {
            const result = { status: response.status, headers: Object.fromEntries(response.headers.entries()) }

            try {
                const data = (await response.json()) as T
                return { ...result, data }
            } catch {
                return result
            }
        })

    if (!isEnabled || cacheMaxAge <= 0) {
        return fetchData()
    }

    const key = JSON.stringify({ url, ...requestInit })
    const minCreatedAt = Date.now() - cacheMaxAge

    if (!cache.has(key) || cache.get(key)!.createdAt < minCreatedAt) {
        const request = fetchData().catch(error => {
            cache.delete(key)
            throw error
        })
        cache.set(key, { createdAt: Date.now(), request })
    }

    return cache.get(key)!.request as Promise<FetchCacheResponse<T>>
}
/**
 * For unit testing purposes only
 */
export const enableFetchCache = (): void => {
    isEnabled = true
}

/**
 * For unit testing purposes only
 */
export const disableFetchCache = (): void => {
    isEnabled = false
}

/**
 * For unit testing purposes only
 */
export const clearFetchCache = (): void => cache.clear()
