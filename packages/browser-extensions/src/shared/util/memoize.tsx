import { Observable } from 'rxjs'
import { publishReplay, refCount, tap } from 'rxjs/operators'

/**
 * Creates a function that memoizes the async result of func.
 * If the promise rejects, the value will not be cached.
 *
 * @param resolver If resolver provided, it determines the cache key for storing the result based on
 * the first argument provided to the memoized function.
 */
export function memoizeAsync<P, T>(
    func: (params: P) => Promise<T>,
    resolver?: (params: P) => string
): (params: P, force?: boolean) => Promise<T> {
    const cache = new Map<string, Promise<T>>()
    return (params: P, force = false) => {
        const key = resolver ? resolver(params) : params.toString()
        const hit = cache.get(key)
        if (!force && hit) {
            return hit
        }
        const p = func(params).catch(e => {
            cache.delete(key)
            throw e
        })
        cache.set(key, p)
        return p
    }
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
    return (params: P, force = false) => {
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
