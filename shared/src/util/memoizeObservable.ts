import { Observable } from 'rxjs'
import { publishReplay, refCount, tap } from 'rxjs/operators'

let allCachesResetSeq = 0

/**
 * Clears all memoized data for memoizeObservable calls. All calls made to those functions after
 * clearing will result in the fetch func being called again.
 *
 * You must call this function after you've modified a memoized resource, or else some components of
 * the UI may have a stale view of the resource.
 */
export function resetAllMemoizationCaches(): void {
    allCachesResetSeq++
}

/**
 * Creates a function that memoizes the observable result of func.
 * If the Observable errors, the value will not be cached.
 *
 * @param resolver If resolver provided, it determines the cache key for storing the result based on
 * the first argument provided to the memoized function.
 */
export function memoizeObservable<P, T>(
    func: (params: P) => Observable<T>,
    resolver?: (params: P) => string
): (params: P, force?: boolean) => Observable<T> {
    const cache = new Map<string, Observable<T>>()
    let cacheResetSeq = allCachesResetSeq
    return (params: P, force = false) => {
        // Reset cache if resetAllMemoizationCaches was called.
        if (cacheResetSeq < allCachesResetSeq) {
            cache.clear()
            cacheResetSeq = allCachesResetSeq
        }

        const key = resolver ? resolver(params) : params.toString()
        const hit = cache.get(key)
        if (!force && hit) {
            return hit
        }
        const obs = func(params).pipe(
            publishReplay(),
            refCount(),
            tap(undefined as any, e => {
                cache.delete(key)
            })
        )
        cache.set(key, obs)
        return obs
    }
}
