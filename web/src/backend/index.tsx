import 'rxjs/add/operator/toPromise'
import { Observable } from 'rxjs/Observable'
import { FileFilter, FileGlobFilter, FilterType, RepoFilter, SearchOptions } from 'sourcegraph/search'
import * as util from 'sourcegraph/util'
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

export function fetchDependencyReferences(repo: string, rev: string, path: string, line: number, character: number): Promise<GQL.IDependencyReferences | null> {
    const mode = util.getModeFromExtension(util.getPathExtension(path))
    const p = queryGraphQL(`
        query DependencyReferences($repo: String, $rev: String, $mode: String, $line: Int, $character: Int) {
            root {
                repository(uri: $repo) {
                    commit(rev: $rev) {
                        commit {
                            file(path: $path) {
                                dependencyReferences(Language: $mode, Line: $line, Character: $character) {
                                    dependencyReferenceData {
                                        references {
                                            dependencyData
                                            repoId
                                            hints
                                        }
                                        location {
                                            location
                                            symbol
                                        }
                                    }
                                    repoData {
                                        repos {
                                            id
                                            uri
                                            lastIndexedRevOrLatest {
                                                commit {
                                                    sha1
                                                }
                                            }
                                        }
                                        repoIds
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
    `, { repo, rev, mode, path, line, character }).toPromise().then(result => {
        // Note: only cache the promise if it is not found or found. If it is cloning, we want to recheck.
        if (!result.data ||
            !result.data.root ||
            !result.data.root.repository ||
            !result.data.root.repository.commit ||
            !result.data.root.repository.commit.commit ||
            !result.data.root.repository.commit.commit.file ||
            !result.data.root.repository.commit.commit.file.dependencyReferences ||
            !result.data.root.repository.commit.commit.file.dependencyReferences.repoData ||
            !result.data.root.repository.commit.commit.file.dependencyReferences.dependencyReferenceData ||
            !result.data.root.repository.commit.commit.file.dependencyReferences.dependencyReferenceData.references.length) {
            return null
        }

        return result.data.root.repository.commit.commit.file.dependencyReferences
    })
    return p
}

export interface SearchResult {
    limitHit: boolean
    lineMatches: LineMatch[]
    resource: string // a URI like git://github.com/gorilla/mux
}

export interface LineMatch {
    lineNumber: number
    offsetAndLengths: number[][] // e.g. [[4, 3]]
    preview: string
}

export function searchText(params: SearchOptions): Observable<GQL.ISearchResults> {
    const variables = {
        pattern: params.query,
        fileMatchLimit: 500,
        isRegExp: params.matchRegex,
        isWordMatch: params.matchWord,
        repositories: params.filters.filter(f => f.type === FilterType.Repo).map((f: RepoFilter) => ({ repo: f.repoPath })),
        isCaseSensitive: params.matchCase,
        includePattern: [
            ...params.filters.filter(f => f.type === FilterType.File).map((f: FileFilter) => f.filePath),
            ...params.filters.filter(f => f.type === FilterType.FileGlob).map((f: FileGlobFilter) => f.glob)
        ].join(','),
        excludePattern: '{.git,**/.git,.svn,**/.svn,.hg,**/.hg,CVS,**/CVS,.DS_Store,**/.DS_Store,node_modules,bower_components,vendor,dist,out,Godeps,third_party}'
    }

    return queryGraphQL(`
        query SearchText(
            $pattern: String!,
            $fileMatchLimit: Int!,
            $isRegExp: Boolean!,
            $isWordMatch: Boolean!,
            $repositories: [RepositoryRevision!]!,
            $isCaseSensitive: Boolean!,
            $includePattern: String!,
            $excludePattern: String!,
        ) {
            root {
                searchRepos(
                    repositories: $repositories,
                    query: {
                        pattern: $pattern,
                        isRegExp: $isRegExp,
                        fileMatchLimit: $fileMatchLimit,
                        isWordMatch: $isWordMatch,
                        isCaseSensitive: $isCaseSensitive,
                        includePattern: $includePattern,
                        excludePattern: $excludePattern,
                }) {
                    limitHit
                    results {
                        resource
                        limitHit
                        lineMatches {
                            preview
                            lineNumber
                            offsetAndLengths
                        }
                    }
                }
            }
        }
    `, variables)
        .do(result => console.error(...result.errors || []))
        .map(result => result.data!.root.searchRepos)
}

export function fetchActiveRepos(): Promise<GQL.IActiveRepoResults | null> {
    return queryGraphQL(`
        query ActiveRepos() {
            root {
                activeRepos() {
                    active
                    inactive
                }
            }
        }
    `).toPromise().then(result => {
        if (!result.data ||
            !result.data.root ||
            !result.data.root.activeRepos) {
            return null
        }
        return result.data.root.activeRepos
    })
}
