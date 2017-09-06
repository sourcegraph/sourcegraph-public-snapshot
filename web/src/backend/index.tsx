import 'rxjs/add/operator/toPromise'
import { SearchParams } from 'sourcegraph/search'
import * as util from 'sourcegraph/util'
import { queryGraphQL } from './graphql'

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

export interface ResolvedSearchTextResp {
    results?: SearchResult[]
    notFound?: boolean
}

export function searchText(query: string, repositories: { repo: string, rev: string }[], params: SearchParams): Promise<ResolvedSearchTextResp> {
    const variables = {
        pattern: query,
        fileMatchLimit: 500,
        isRegExp: params.matchRegex,
        isWordMatch: params.matchWord,
        repositories,
        isCaseSensitive: params.matchCase,
        // TODO(john)??: currently VS Code converts a string like "*.go" into "{*.go/**,*.go,**/*.go}" -- should we similarly add "**" glob patterns here?
        includePattern: params.files !== '' ? `{${params.files}` : '',
        excludePattern: '{.git,**/.git,.svn,**/.svn,.hg,**/.hg,CVS,**/CVS,.DS_Store,**/.DS_Store,node_modules,bower_components,vendor,dist,out,Godeps,third_party}'
    }

    const p = queryGraphQL(`
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
    `, variables).toPromise().then(result => {
        const results = result.data && result.data.root!.searchRepos
        if (!results) {
            const notFound = { notFound: true }
            return notFound
        }

        return results
    })

    return p
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
