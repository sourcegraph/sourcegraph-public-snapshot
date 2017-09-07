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
        if (!result.data || !result.data.root.repository) {
            throw new Error('invalid response received from graphql endpoint')
        }
        if (result.data.root.repository.commit.cloneInProgress) {
            throw new CloneInProgressError(ctx.repoPath)
        }
        if (!result.data.root.repository.commit.commit) {
            throw new Error('not able to resolve sha1')
        }
        return result.data.root.repository.commit.commit.sha1
    }), makeRepoURI
)

export const fetchHighlightedFile = memoizedFetch((ctx: { repoPath: string, commitID: string, filePath: string, disableTimeout: boolean }): Promise<GQL.IHighlightedFile> =>
    queryGraphQL(`query HighlightedFile($repoPath: String, $commitID: String, $filePath: String, $disableTimeout: Boolean) {
        root {
            repository(uri: $repoPath) {
                commit(rev: $commitID) {
                    commit {
                        file(path: $filePath) {
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
            throw new Error(`error fetching highlighted file: ${errors}`)
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
        return result.data.root.repository.commit.commit.file.highlight
    }), ctx => makeRepoURI(ctx) + `?disableTimeout=${ctx.disableTimeout}`
)

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

export const fetchBlobContent = memoizedFetch((ctx: { repoPath: string, commitID: string, filePath: string }): Promise<string> =>
    queryGraphQL(`query BlobContent($repoPath: String, $commitID: String, $filePath: String) {
        root {
            repository(uri: $repoPath) {
                commit(rev: $commitID) {
                    commit {
                        file(path: $filePath) {
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
        return result.data.root.repository.commit.commit.file.content
    }), makeRepoURI
)
