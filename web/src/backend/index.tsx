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

export function fetchBlameFile(repo: string, rev: string, path: string, startLine: number, endLine: number): Promise<GQL.IHunk[] | null> {
    const p = queryGraphQL(`
        query BlameFile($repo: String, $rev: String, $path: String, $startLine: Int, $endLine: Int) {
            root {
                repository(uri: $repo) {
                    commit(rev: $rev) {
                        commit {
                            file(path: $path) {
                                blame(startLine: $startLine, endLine: $endLine) {
                                    startLine
                                    endLine
                                    startByte
                                    endByte
                                    rev
                                    author {
                                        person {
                                            name
                                            email
                                            gravatarHash
                                        }
                                        date
                                    }
                                    message
                                }
                            }
                        }
                    }
                }
            }
        }
    `, { repo, rev, path, startLine, endLine }).toPromise().then(result => {
        if (!result.data ||
            !result.data.root ||
            !result.data.root.repository ||
            !result.data.root.repository.commit ||
            !result.data.root.repository.commit.commit ||
            !result.data.root.repository.commit.commit.file ||
            !result.data.root.repository.commit.commit.file.blame) {
            console.error('unexpected BlameFile response:', result)
            return null
        }
        return result.data.root.repository.commit.commit.file.blame
    })
    return p
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
