import { Observable } from 'rxjs/Observable'
import { map } from 'rxjs/operators/map'
import { queryGraphQL } from '../backend/graphql'
import { memoizeObservable } from '../util/memoize'
import { makeRepoURI } from './index'

export const ECLONEINPROGESS = 'ECLONEINPROGESS'
class CloneInProgressError extends Error {
    public readonly code = ECLONEINPROGESS
    constructor(repoPath: string) {
        super(`${repoPath} is clone in progress`)
    }
}

export const EREPONOTFOUND = 'EREPONOTFOUND'
class RepoNotFoundError extends Error {
    public readonly code = EREPONOTFOUND
    constructor(repoPath: string) {
        super(`repo ${repoPath} not found`)
    }
}

export const EREVNOTFOUND = 'EREVNOTFOUND'
class RevNotFoundError extends Error {
    public readonly code = EREVNOTFOUND
    constructor(rev?: string) {
        super(`rev ${rev} not found`)
    }
}

export const ERREPOSEEOTHER = 'ERREPOSEEOTHER'
export class RepoSeeOtherError extends Error {
    public readonly code = ERREPOSEEOTHER
    constructor(public redirectURL: string) {
        super(`repo not found at this location, but might exist at ${redirectURL}`)
    }
}

export interface ResolvedRev {
    commitID: string
    defaultBranch: string
}

/**
 * When `rev` is undefined, the default branch is resolved.
 * @return Observable that emits the commit ID
 *         Errors with a `CloneInProgressError` if the repo is still being cloned.
 */
export const resolveRev = memoizeObservable(
    (ctx: { repoPath: string; rev?: string }): Observable<ResolvedRev> =>
        queryGraphQL(
            `
            query ResolveRev($repoPath: String, $rev: String) {
                root {
                    repository(uri: $repoPath) {
                        commit(rev: $rev) {
                            cloneInProgress,
                            commit {
                                sha1
                            }
                        }
                        defaultBranch
                        redirectURL
                    }
                }
            }
            `,
            { ...ctx, rev: ctx.rev || '' }
        ).pipe(
            map(({ data, errors }) => {
                if (!data) {
                    throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
                }
                if (data.root.repository && data.root.repository.redirectURL) {
                    throw new RepoSeeOtherError(data.root.repository.redirectURL)
                }
                if (!data.root.repository || !data.root.repository.commit) {
                    throw new RepoNotFoundError(ctx.repoPath)
                }
                if (data.root.repository.commit.cloneInProgress) {
                    throw new CloneInProgressError(ctx.repoPath)
                }
                if (!data.root.repository.commit.commit) {
                    throw new RevNotFoundError(ctx.rev)
                }
                if (!data.root.repository.defaultBranch) {
                    throw new RevNotFoundError('HEAD')
                }
                return {
                    commitID: data.root.repository.commit.commit.sha1,
                    defaultBranch: data.root.repository.defaultBranch,
                }
            })
        ),
    makeRepoURI
)

interface FetchFileCtx {
    repoPath: string
    commitID: string
    filePath: string
    disableTimeout?: boolean
    isLightTheme: boolean
}

interface HighlightedFileResult {
    isDirectory: boolean
    highlightedFile: GQL.IHighlightedFile
}

export const fetchHighlightedFile = memoizeObservable(
    (ctx: FetchFileCtx): Observable<HighlightedFileResult> =>
        queryGraphQL(
            `query HighlightedFile($repoPath: String, $commitID: String, $filePath: String, $disableTimeout: Boolean, $isLightTheme: Boolean) {
                root {
                    repository(uri: $repoPath) {
                        commit(rev: $commitID) {
                            commit {
                                file(path: $filePath) {
                                    isDirectory
                                    highlight(disableTimeout: $disableTimeout, isLightTheme: $isLightTheme) {
                                        aborted
                                        html
                                    }
                                }
                            }
                        }
                    }
                }
            }`,
            ctx
        ).pipe(
            map(({ data, errors }) => {
                if (
                    !data ||
                    !data.root ||
                    !data.root.repository ||
                    !data.root.repository.commit ||
                    !data.root.repository.commit.commit ||
                    !data.root.repository.commit.commit.file
                ) {
                    throw Object.assign(
                        new Error('Could not fetch highlighted file: ' + (errors || []).map(e => e.message).join('\n')),
                        { errors }
                    )
                }
                const file = data.root.repository.commit.commit.file
                return { isDirectory: file.isDirectory, highlightedFile: file.highlight }
            })
        ),
    ctx => makeRepoURI(ctx) + `?disableTimeout=${ctx.disableTimeout} ` + `?isLightTheme=${ctx.isLightTheme}`
)

/**
 * Produces a list like ['<tr>...</tr>', ...]
 */
export const fetchHighlightedFileLines = memoizeObservable(
    (ctx: FetchFileCtx, force?: boolean): Observable<string[]> =>
        fetchHighlightedFile(ctx, force).pipe(
            map(result => {
                if (result.isDirectory) {
                    return []
                }
                if (result.highlightedFile.aborted) {
                    throw new Error('aborted fetching highlighted contents')
                }
                let parsed = result.highlightedFile.html.substr('<table>'.length)
                parsed = parsed.substr(0, parsed.length - '</table>'.length)
                const rows = parsed.split('</tr>')
                for (let i = 0; i < rows.length; ++i) {
                    rows[i] += '</tr>'
                }
                return rows
            })
        ),
    ctx => makeRepoURI(ctx) + `?isLightTheme=${ctx.isLightTheme}`
)

export const listAllFiles = memoizeObservable(
    (ctx: { repoPath: string; commitID: string }): Observable<string[]> =>
        queryGraphQL(
            `query FileTree($repoPath: String!, $commitID: String!) {
                root {
                    repository(uri: $repoPath) {
                        commit(rev: $commitID) {
                            commit {
                                tree(recursive: true) {
                                    files {
                                        name
                                    }
                                }
                            }
                        }
                    }
                }
            }`,
            ctx
        ).pipe(
            map(({ data, errors }) => {
                if (
                    !data ||
                    !data.root.repository ||
                    !data.root.repository.commit ||
                    !data.root.repository.commit.commit ||
                    !data.root.repository.commit.commit.tree ||
                    !data.root.repository.commit.commit.tree.files
                ) {
                    throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
                }
                return data.root.repository.commit.commit.tree.files.map(file => file.name)
            })
        ),
    makeRepoURI
)

interface BlobContent {
    isDirectory: boolean
    content: string
}

export const fetchBlobContent = memoizeObservable(
    (ctx: FetchFileCtx): Observable<BlobContent> =>
        queryGraphQL(
            `query BlobContent($repoPath: String, $commitID: String, $filePath: String) {
                root {
                    repository(uri: $repoPath) {
                        commit(rev: $commitID) {
                            commit {
                                file(path: $filePath) {
                                    isDirectory
                                    content
                                }
                            }
                        }
                    }
                }
            }`,
            ctx
        ).pipe(
            map(({ data, errors }) => {
                if (
                    !data ||
                    !data.root ||
                    !data.root.repository ||
                    !data.root.repository.commit ||
                    !data.root.repository.commit.commit ||
                    !data.root.repository.commit.commit.file
                ) {
                    throw Object.assign(
                        'Could not fetch blob content: ' + new Error((errors || []).map(e => e.message).join('\n')),
                        { errors }
                    )
                }
                const file = data.root.repository.commit.commit.file
                return { isDirectory: file.isDirectory, content: file.content }
            })
        ),
    makeRepoURI
)

export interface RepoRevisions {
    branches: string[]
    tags: string[]
}

export const fetchRepoRevisions = memoizeObservable(
    (ctx: { repoPath: string }): Observable<RepoRevisions> =>
        queryGraphQL(
            `query RepoRevisions($repoPath: String) {
                root {
                    repository(uri: $repoPath) {
                        branches
                        tags
                    }
                }
            }`,
            ctx
        ).pipe(
            map(({ data, errors }) => {
                if (
                    !data ||
                    !data.root ||
                    !data.root.repository ||
                    !data.root.repository.branches ||
                    !data.root.repository.tags
                ) {
                    throw Object.assign(
                        'Could not fetch repo revisions: ' + new Error((errors || []).map(e => e.message).join('\n')),
                        { errors }
                    )
                }
                return data.root.repository
            })
        ),
    makeRepoURI
)

export const fetchPhabricatorRepo = memoizeObservable(
    (ctx: { repoPath: string }): Observable<GQL.IPhabricatorRepo | null> =>
        queryGraphQL(
            `query PhabricatorRepo($repoPath: String) {
                root {
                    phabricatorRepo(uri: $repoPath) {
                        callsign
                        uri
                    }
                }
            }`,
            ctx
        ).pipe(
            map(result => {
                if (result.errors || !result.data || !result.data.root || !result.data.root.phabricatorRepo) {
                    return null
                }
                return result.data.root.phabricatorRepo
            })
        ),
    makeRepoURI
)

export const fetchDirTree = memoizeObservable(
    (ctx: { repoPath: string; commitID: string; filePath: string }): Observable<GQL.ITree> =>
        queryGraphQL(
            `query fetchDirectoryTree($repoPath: String, $commitID: String, $filePath: String) {
                root {
                    repository(uri: $repoPath) {
                        commit(rev: $commitID) {
                            commit {
                                tree(path: $filePath) {
                                    directories {
                                        name
                                    }
                                    files {
                                        name
                                    }
                                }
                            }
                        }
                    }
                }
            }`,
            ctx
        ).pipe(
            map(({ data, errors }) => {
                if (
                    !data ||
                    !data.root ||
                    !data.root.repository ||
                    !data.root.repository.commit.commit ||
                    !data.root.repository.commit.commit.tree
                ) {
                    throw Object.assign(
                        'Could not fetch directory tree: ' + new Error((errors || []).map(e => e.message).join('\n')),
                        { errors }
                    )
                }
                return data.root.repository.commit.commit.tree
            })
        ),
    makeRepoURI
)

export const fetchFileCommitInfo = memoizeObservable(
    (ctx: { repoPath: string; commitID: string; filePath: string }): Observable<GQL.ICommitInfo> =>
        queryGraphQL(
            `query fetchFileCommitInfo($repoPath: String, $commitID: String, $filePath: String) {
                root {
                    repository(uri: $repoPath) {
                        commit(rev: $commitID) {
                            commit {
                                file(path: $filePath) {
                                    lastCommit {
                                        rev
                                        message
                                        committer {
                                            person {
                                                name
                                                avatarURL
                                            }
                                            date
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }`,
            ctx
        ).pipe(
            map(({ data, errors }) => {
                if (
                    !data ||
                    !data.root ||
                    !data.root.repository ||
                    !data.root.repository.commit.commit ||
                    !data.root.repository.commit.commit.file ||
                    !data.root.repository.commit.commit.file.lastCommit
                ) {
                    throw Object.assign(
                        'Could not fetch commit info: ' + new Error((errors || []).map(e => e.message).join('\n')),
                        { errors }
                    )
                }
                return data.root.repository.commit.commit.file.lastCommit
            })
        ),
    makeRepoURI
)

export const fetchRepositories = memoizeObservable(
    (ctx: { query: string }): Observable<GQL.IRepository[]> =>
        queryGraphQL(
            `query fetchRepositories($query: String) {
                root {
                    repositories(query: $query) {
                        uri
                        description
                        private
                    }
                }
            }`,
            ctx
        ).pipe(
            map(({ data, errors }) => {
                if (!data || !data.root || !data.root.repositories) {
                    throw Object.assign(
                        'Could not fetch repositories: ' + new Error((errors || []).map(e => e.message).join('\n')),
                        { errors }
                    )
                }
                return data.root.repositories
            })
        ),
    ctx => ctx.query
)
