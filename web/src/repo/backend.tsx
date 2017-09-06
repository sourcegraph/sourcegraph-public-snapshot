import memoize = require('lodash/memoize')
import { queryGraphQL } from 'sourcegraph/backend/graphql'
import { makeRepoURI } from 'sourcegraph/repo'

export interface ResolvedRev {
    cloneInProgress: boolean
    commitID?: string
}

export const resolveRev = memoize((ctx: { repoPath: string, rev?: string }): Promise<ResolvedRev> =>
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
        const resolved: ResolvedRev = { cloneInProgress: false }
        if (!result.data || !result.data.root.repository) {
            throw new Error('invalid response received from graphql endpoint')
        }
        if (result.data.root.repository.commit.cloneInProgress) {
            resolved.cloneInProgress = true
            return resolved
        }
        if (!result.data.root.repository.commit.commit) {
            throw new Error('not able to resolve sha1')
        }
        resolved.commitID = result.data.root.repository.commit.commit.sha1
        return resolved
    }), makeRepoURI
)

export const fetchHighlightedFile = memoize((ctx: { repoPath: string, commitID: string, filePath: string, disableTimeout: boolean }): Promise<GQL.IHighlightedFile> =>
    queryGraphQL(`query HighlightedFile($repoPath: String, $commitID: String, $filePath: String, $disableTimeout: Boolean) {
        root {
            repository(uri: $repoPath) {
                commit(rev: $commitID) {
                    commit {
                        file(path: $filePath) {
                            highlight(disableTimeout: $disableTimeout) {
                                isBinary
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

export const listAllFiles = memoize((ctx: { repoPath: string, commitID: string }): Promise<string[]> =>
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

export const fetchBlobContent = memoize((ctx: { repoPath: string, commitID: string, filePath: string }): Promise<string> =>
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
