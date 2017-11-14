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
                }
            }
        }
    `,
            { ...ctx, rev: ctx.rev || '' }
        ).pipe(
            map(result => {
                if (!result.data) {
                    throw new Error('invalid response received from graphql endpoint')
                }
                if (!result.data.root.repository || !result.data.root.repository.commit) {
                    throw new RepoNotFoundError(ctx.repoPath)
                }
                if (result.data.root.repository.commit.cloneInProgress) {
                    throw new CloneInProgressError(ctx.repoPath)
                }
                if (!result.data.root.repository.commit.commit) {
                    throw new RevNotFoundError(ctx.rev)
                }
                return {
                    commitID: result.data.root.repository.commit.commit.sha1,
                    defaultBranch: result.data.root.repository.defaultBranch,
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
}

interface HighlightedFileResult {
    isDirectory: boolean
    highlightedFile: GQL.IHighlightedFile
}

export const fetchHighlightedFile = memoizeObservable(
    (ctx: FetchFileCtx): Observable<HighlightedFileResult> =>
        queryGraphQL(
            `query HighlightedFile($repoPath: String, $commitID: String, $filePath: String, $disableTimeout: Boolean) {
        root {
            repository(uri: $repoPath) {
                commit(rev: $commitID) {
                    commit {
                        file(path: $filePath) {
                            isDirectory
                            highlight(disableTimeout: $disableTimeout) {
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
            map(result => {
                if (result.errors) {
                    const errors = result.errors.map(e => e.message).join(', ')
                    throw new Error(errors)
                }
                if (
                    !result.data ||
                    !result.data.root ||
                    !result.data.root.repository ||
                    !result.data.root.repository.commit ||
                    !result.data.root.repository.commit.commit ||
                    !result.data.root.repository.commit.commit.file
                ) {
                    throw new Error(`cannot locate blob content: ${ctx.repoPath} ${ctx.commitID} ${ctx.filePath}`)
                }
                const file = result.data.root.repository.commit.commit.file
                return { isDirectory: file.isDirectory, highlightedFile: file.highlight }
            })
        ),
    ctx => makeRepoURI(ctx) + `?disableTimeout=${ctx.disableTimeout}`
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
    makeRepoURI
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
            map(result => {
                if (
                    !result.data ||
                    !result.data.root.repository ||
                    !result.data.root.repository.commit ||
                    !result.data.root.repository.commit.commit ||
                    !result.data.root.repository.commit.commit.tree ||
                    !result.data.root.repository.commit.commit.tree.files
                ) {
                    throw new Error('invalid response received from graphql endpoint')
                }
                return result.data.root.repository.commit.commit.tree.files.map(file => file.name)
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
            map(result => {
                if (
                    !result.data ||
                    !result.data.root ||
                    !result.data.root.repository ||
                    !result.data.root.repository.commit ||
                    !result.data.root.repository.commit.commit ||
                    !result.data.root.repository.commit.commit.file
                ) {
                    throw new Error(`cannot locate blob content: ${ctx}`)
                }
                const file = result.data.root.repository.commit.commit.file
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
            map(result => {
                if (result.errors) {
                    const errors = result.errors.map(e => e.message).join(', ')
                    throw new Error(errors)
                }
                if (
                    !result.data ||
                    !result.data.root ||
                    !result.data.root.repository ||
                    !result.data.root.repository.branches ||
                    !result.data.root.repository.tags
                ) {
                    throw new Error(`cannot locate repo revisions: ${ctx}`)
                }
                return result.data.root.repository
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
            map(result => {
                if (result.errors) {
                    const errors = result.errors.map(e => e.message).join(', ')
                    throw new Error(errors)
                }
                if (
                    !result.data ||
                    !result.data.root ||
                    !result.data.root.repository ||
                    !result.data.root.repository.commit.commit ||
                    !result.data.root.repository.commit.commit.tree
                ) {
                    throw new Error(`cannot locate directory tree.`)
                }
                return result.data.root.repository.commit.commit.tree
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
            map(result => {
                if (result.errors) {
                    const errors = result.errors.map(e => e.message).join(', ')
                    throw new Error(errors)
                }
                if (
                    !result.data ||
                    !result.data.root ||
                    !result.data.root.repository ||
                    !result.data.root.repository.commit.commit ||
                    !result.data.root.repository.commit.commit.file ||
                    !result.data.root.repository.commit.commit.file.lastCommit
                ) {
                    throw new Error(`cannot locate file commit info.`)
                }
                return result.data.root.repository.commit.commit.file.lastCommit
            })
        ),
    makeRepoURI
)
