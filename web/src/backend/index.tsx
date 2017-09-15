import 'rxjs/add/operator/toPromise'
import { queryGraphQL } from './graphql'

export function memoizedFetch<K, T>(fetch: (ctx: K, force?: boolean) => Promise<T>, makeKey?: (ctx: K) => string): (ctx: K, force?: boolean) => Promise<T> {
    const cache = new Map<string, Promise<T>>()
    return (ctx: K, force?: boolean) => {
        const key = makeKey ? makeKey(ctx) : ctx.toString()
        const hit = cache.get(key)
        if (!force && hit) {
            return hit
        }
        const p = fetch(ctx, force)
        cache.set(key, p)
        return p.catch(e => {
            cache.delete(key)
            throw e
        })
    }
}

export function fetchRepos(query: string): Promise<GQL.IRepository[]> {
    const p = queryGraphQL(`
        query SearchRepos($query: String, $fast: Boolean) {
            root {
                repositories(query: $query, fast: $fast) {
                    uri
                    description
                    private
                    fork
                    pushedAt
                }
            }
        }
    `, { query, fast: true }).toPromise().then(result => {
        if (!result.data ||
            !result.data.root ||
            !result.data.root.repositories) {
            return []
        }

        return result.data.root.repositories
    })
    return p
}
