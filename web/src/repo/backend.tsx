import 'rxjs/add/operator/toPromise'
import { memoizedFetch } from 'sourcegraph/backend'
import { queryGraphQL } from 'sourcegraph/backend/graphql'
import { makeRepoURI } from 'sourcegraph/repo'

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

/**
 * @return Promise that resolves to the commit ID
 *         Will reject with a `CloneInProgressError` if the repo is still being cloned.
 */
export const resolveRev = memoizedFetch((ctx: { repoPath: string, rev?: string }): Promise<string> =>
    queryGraphQL(`
        query ResolveRev($repoPath: String, $rev: String) {
            root {
                repository(uri: $repoPath) {
                    commit(rev: $rev) {
                        cloneInProgress,
                        commit {
                            sha1
                        }
                    }
                }
            }
        }
    `, { ...ctx, rev: ctx.rev || 'master' }).toPromise().then(result => {
        if (!result.data) {
            throw new Error('invalid response received from graphql endpoint')
        }
        if (!result.data.root.repository) {
            throw new RepoNotFoundError(ctx.repoPath)
        }
        if (result.data.root.repository.commit.cloneInProgress) {
            throw new CloneInProgressError(ctx.repoPath)
        }
        if (!result.data.root.repository.commit.commit) {
            throw new RevNotFoundError(ctx.rev)
        }
        return result.data.root.repository.commit.commit.sha1
    }), makeRepoURI
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

export const fetchHighlightedFile = memoizedFetch((ctx: FetchFileCtx): Promise<HighlightedFileResult> =>
    queryGraphQL(`query HighlightedFile($repoPath: String, $commitID: String, $filePath: String, $disableTimeout: Boolean) {
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
    }`, ctx).toPromise().then(result => {
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
    }), ctx => makeRepoURI(ctx) + `?disableTimeout=${ctx.disableTimeout}`
)

/**
 * Produces a list like ['<tr>...</tr>', ...]
 */
export const fetchHighlightedFileLines = memoizedFetch((ctx: FetchFileCtx, force?: boolean): Promise<string[]> =>
        fetchHighlightedFile(ctx, force).then(result => {
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
    , makeRepoURI)

export const listAllFiles = memoizedFetch((ctx: { repoPath: string, commitID: string }): Promise<string[]> =>
    queryGraphQL(`query FileTree($repoPath: String!, $commitID: String!) {
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
    }`, ctx).toPromise().then(result => {
        if (!result.data ||
            !result.data.root.repository ||
            !result.data.root.repository.commit ||
            !result.data.root.repository.commit.commit ||
            !result.data.root.repository.commit.commit.tree ||
            !result.data.root.repository.commit.commit.tree.files
        ) {
            throw new Error('invalid response received from graphql endpoint')
        }
        return result.data.root.repository.commit.commit.tree.files.map(file => file.name)
    }), makeRepoURI
)

interface BlobContent {
    isDirectory: boolean
    content: string
}

export const fetchBlobContent = memoizedFetch((ctx: FetchFileCtx): Promise<BlobContent> =>
    queryGraphQL(`query BlobContent($repoPath: String, $commitID: String, $filePath: String) {
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
    }`, ctx).toPromise().then(result => {
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
    }), makeRepoURI
)

export interface RepoRevisions {
    branches: string[]
    tags: string[]
}

export const fetchRepoRevisions = memoizedFetch((ctx: { repoPath: string }): Promise<RepoRevisions> =>
    queryGraphQL(`query RepoRevisions($repoPath: String) {
        root {
            repository(uri: $repoPath) {
                branches
                tags
            }
        }
    }`, ctx).toPromise().then(result => {
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
    }), makeRepoURI
)
