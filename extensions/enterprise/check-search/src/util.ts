import * as sourcegraph from 'sourcegraph'
import { Observable, from } from 'rxjs'
import localforage from 'localforage'
import { publishReplay, refCount, tap, first } from 'rxjs/operators'

const USE_PERSISTENT_MEMOIZATION_CACHE = true

interface Cache<T> {
    get(key: string): Promise<T>
    set(key: string, value: T | Promise<T>): Promise<void>
    delete(key: string): Promise<void>
}

const createMemoizationCache = <T>(): Cache<T> => {
    const map = new Map<string, Promise<T>>()
    const cache: Cache<T> = {
        get: async key => {
            const localValue = map.get(key)
            if (localValue !== undefined) {
                return localValue
            }
            return localforage.getItem(key)
        },
        set: async (key, value) => {
            map.set(key, Promise.resolve(value))
            Promise.resolve(value)
                .then(value => localforage.setItem(key, value))
                .catch(err => console.error(err))
        },
        delete: async key => {
            map.delete(key)
            await localforage.removeItem(key)
        },
    }
    return cache
}

const createVolatileCache = <T>(): Cache<T> => {
    const map = new Map<string, Promise<T>>()
    const cache: Cache<T> = {
        get: async key => map.get(key),
        set: async (key, value) => {
            map.set(key, Promise.resolve(value))
        },
        delete: async key => {
            map.delete(key)
        },
    }
    return cache
}

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
    // TODO!(sqs): memoization cache is not keyed to prevent collisions across instances if params
    // key collides. need to add a `keyPrefix` or similar arg to memoizeAsync.
    const cache: Cache<T> = USE_PERSISTENT_MEMOIZATION_CACHE ? createMemoizationCache<T>() : createVolatileCache<T>()
    return async (params: P, force = false) => {
        const key = resolver ? resolver(params) : params.toString()
        const hit = await cache.get(key)
        if (!force && hit) {
            return hit
        }
        const p = func(params).catch(async e => {
            await cache.delete(key)
            throw e
        })
        await cache.set(key, p)
        return p
    }
}

export const queryGraphQL = memoizeAsync(
    async ({ query, vars }: { query: string; vars: { [name: string]: any } }): Promise<any> => {
        return sourcegraph.commands.executeCommand<any>('queryGraphQL', query, vars)
    },
    arg => JSON.stringify({ query: arg.query, vars: arg.vars })
)

const _findTextInFiles = memoizeAsync(
    (args: Parameters<typeof sourcegraph.search.findTextInFiles>) =>
        from(sourcegraph.search.findTextInFiles(...args))
            .pipe(first())
            .toPromise(),
    args => JSON.stringify(args)
)

export const memoizedFindTextInFiles = (...args: Parameters<typeof sourcegraph.search.findTextInFiles>) =>
    _findTextInFiles(args)

export const settingsObservable = <T extends object>() =>
    new Observable<T>(subscriber =>
        sourcegraph.configuration.subscribe(() => {
            subscriber.next(sourcegraph.configuration.get<T>().value)
        })
    )
