import * as sourcegraph from 'sourcegraph'

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

export const queryGraphQL = memoizeAsync(
    async ({ query, vars }: { query: string; vars: { [name: string]: any } }): Promise<any> => {
        return sourcegraph.commands.executeCommand<any>('queryGraphQL', query, vars)
    },
    arg => JSON.stringify({ query: arg.query, vars: arg.vars })
)
