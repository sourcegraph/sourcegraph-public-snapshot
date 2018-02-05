import { Observable } from 'rxjs/Observable'
import { publishReplay } from 'rxjs/operators/publishReplay'
import { refCount } from 'rxjs/operators/refCount'
import { tap } from 'rxjs/operators/tap'

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
