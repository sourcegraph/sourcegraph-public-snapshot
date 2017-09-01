import * as _ from 'lodash';
import { queryGraphQL } from 'sourcegraph/backend/graphql';
import { singleflightCachedFetch } from 'sourcegraph/repo/cache';

export const resolveRev = singleflightCachedFetch((ctx: { repoPath: string, rev?: string }) =>
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
                throw new Error('invalid response received from graphql endpoint');
            }
            if (result.data.root.repository.commit.cloneInProgress) {
                return { cloneInProgress: true };
            }
            if (!result.data.root.repository.commit.commit) {
                throw new Error('not able to resolve sha1');
            }
            return { commitID: result.data.root.repository.commit.commit.sha1 };
        }), ctx => _.pick(ctx, ['repoPath', 'rev'])
);

export const fetchBlobHighlightContentTable = singleflightCachedFetch((ctx: { repoPath: string, commitID: string, filePath: string }) =>
    queryGraphQL(`query HighlightedBlobContent($repoPath: String, $commitID: String, $filePath: String) {
        root {
            repository(uri: $repoPath) {
                commit(rev: $commitID) {
                    commit {
                        file(path: $filePath) {
                            highlightedContentHTML
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
                throw new Error(`cannot locate blob content: ${ctx.repoPath} ${ctx.commitID} ${ctx.filePath}`);
            }
            return result.data.root.repository.commit.commit.file.highlightedContentHTML;
        }), ctx => _.pick(ctx, ['repoPath', 'commitID', 'filePath'])
);

export const listAllFiles = singleflightCachedFetch((ctx: { repoPath: string, commitID: string }) =>
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
                throw new Error('invalid response received from graphql endpoint');
            }
            return result.data.root.repository.commit.commit.tree.files;
        }), ctx => _.pick(ctx, ['repoPath', 'commitID'])
);

export const fetchBlobContent = singleflightCachedFetch((ctx: { repoPath: string, commitID: string, filePath: string }) =>
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
                throw new Error(`cannot locate blob content: ${ctx}`);
            }
            return result.data.root.repository.commit.commit.file.content;
        }), ctx => _.pick(ctx, ['repoPath', 'commitID', 'filePath'])
);
