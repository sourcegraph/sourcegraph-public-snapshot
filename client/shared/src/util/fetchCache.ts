const cache = new Map<string, FetchCacheReturnType<any>>()
const runningRequests = new Map<string, Promise<FetchCacheReturnType<any>>>()
const INVALIDATE_TIMEOUT = 60 * 1000 // 1 minute

let isEnabled = true

interface FetchCacheReturnType<T> {
    data: T
    status: number
}

/**
 * fetch API with cache
 *
 * @description Caches same argument requests for 1 minute
 */
export const fetchCache = async <T = any>(...args: Parameters<typeof fetch>): Promise<FetchCacheReturnType<T>> => {
    // eslint-disable-next-line @typescript-eslint/explicit-function-return-type
    const doRequest = () =>
        fetch(...args).then(async response => {
            const data = (await response.json()) as T
            return { status: response.status, data }
        })

    if (!isEnabled) {
        return doRequest()
    }

    const key = JSON.stringify(args)
    if (cache.has(key)) {
        return cache.get(key) as FetchCacheReturnType<T>
    }

    if (runningRequests.has(key)) {
        return (await runningRequests.get(key)) as FetchCacheReturnType<T>
    }

    const request = doRequest()
        .then(result => {
            cache.set(key, result)
            setTimeout(() => cache.delete(key), INVALIDATE_TIMEOUT)

            return result
        })
        .finally(() => runningRequests.delete(key))

    runningRequests.set(key, request)

    return request
}
/**
 * For unit testing purposes
 */
fetchCache.enableCache = (): void => {
    isEnabled = true
}

fetchCache.disableCache = (): void => {
    isEnabled = false
}
