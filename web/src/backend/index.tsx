import 'rxjs/add/operator/toPromise';
import { SearchParams } from 'sourcegraph/search';
import * as util from 'sourcegraph/util';
import { queryGraphQL } from './graphql';

export const cacheKey = (repo: string, rev?: string) => `${repo}@${rev || null}`;

export interface ResolvedRevResp {
    notFound?: boolean;
    cloneInProgress?: boolean;
    commitID?: string;
}

const promiseCache = new Map<string, Promise<ResolvedRevResp>>();

export function resolveRev(repo: string, rev?: string): Promise<ResolvedRevResp> {
    const key = cacheKey(repo, rev);
    const promiseHit = promiseCache.get(key);
    if (promiseHit) {
        return promiseHit;
    }
    const p = queryGraphQL(`
        query ResolveRev($repo: String, $rev: String) {
            root {
                repository(uri: $repo) {
                    commit(rev: $rev) {
                        cloneInProgress,
                        commit {
                            sha1
                        }
                    }
                }
            }
        }
    `, { repo, rev: rev || 'master' }).toPromise().then(result => {
        // Note: only cache the promise if it is not found or found. If it is cloning, we want to recheck.
        promiseCache.delete(key);
        if (!result.data) {
            const error = new Error('invalid response received from graphql endpoint');
            promiseCache.set(key, Promise.reject(error));
            throw error;
        }
        if (!result.data.root.repository) {
            const notFound = { notFound: true };
            promiseCache.set(key, Promise.resolve(notFound));
            return notFound;
        }
        if (result.data.root.repository.commit.cloneInProgress) {
            // don't store this promise, we want to make a new query, holmes.
            return { cloneInProgress: true };
        } else if (!result.data.root.repository.commit.commit) {
            const error = new Error('not able to resolve sha1');
            promiseCache.set(key, Promise.reject(error));
            throw error;
        }
        const found = { commitID: result.data.root.repository.commit.commit.sha1 };
        promiseCache.set(key, Promise.resolve(found));
        return found;
    });
    promiseCache.set(key, p);
    return p;
}

export const blobKey = (repoURI: string, rev: string, path: string) => `${repoURI}@${rev}#${path}`;
const blobCache = new Map<string, Promise<string>>();

export function fetchBlobContent(repoURI: string, rev: string, path: string): Promise<string> {
    const key = blobKey(repoURI, rev, path);
    const promiseHit = blobCache.get(key);
    if (promiseHit) {
        return promiseHit;
    }
    const p = queryGraphQL(`
        query BlobContent($repo: String, $rev: String, $path: String) {
            root {
                repository(uri: $repo) {
                    commit(rev: $rev) {
                        commit {
                            file(path: $path) {
                                content
                            }
                        }
                    }
                }
            }
        }
    `, { repo: repoURI, rev, path }).toPromise().then(result => {
        if (
            !result.data ||
            !result.data.root ||
            !result.data.root.repository ||
            !result.data.root.repository.commit ||
            !result.data.root.repository.commit.commit ||
            !result.data.root.repository.commit.commit.file
        ) {
            throw new Error(`cannot locate blob content: ${key}`);
        }
        return result.data.root.repository.commit.commit.file.content;
    });
    blobCache.set(key, p);
    return p;
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
            console.error('unexpected BlameFile response:', result);
            return null;
        }
        return result.data.root.repository.commit.commit.file.blame;
    });
    return p;
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
            return [];
        }

        return result.data.root.repositories;
    });
    return p;
}

export function fetchDependencyReferences(repo: string, rev: string, path: string, line: number, character: number): Promise<GQL.IDependencyReferences | null> {
    const mode = util.getModeFromExtension(util.getPathExtension(path));
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
            return null;
        }

        return result.data.root.repository.commit.commit.file.dependencyReferences;
    });
    return p;
}

export interface SearchResult {
    limitHit: boolean;
    lineMatches: LineMatch[];
    resource: string; // a URI like git://github.com/gorilla/mux
}

export interface LineMatch {
    lineNumber: number;
    offsetAndLengths: number[][]; // e.g. [[4, 3]]
    preview: string;
}

export interface ResolvedSearchTextResp {
    results?: SearchResult[];
    notFound?: boolean;
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
    };

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
        const results = result.data && result.data.root!.searchRepos;
        if (!results) {
            const notFound = { notFound: true };
            return notFound;
        }

        return results;
    });

    return p;
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
            return null;
        }
        return result.data.root.activeRepos;
    });
}

const fileTreeCache = new Map<string, Promise<any>>();

export function listAllFiles(repo: string, revision: string): Promise<GQL.IFile[]> {
    const key = cacheKey(repo, revision);
    const promiseHit = fileTreeCache.get(key);
    if (promiseHit) {
        return promiseHit;
    }
    const p = queryGraphQL(`
        query FileTree($repo: String!, $revision: String!) {
            root {
                repository(uri: $repo) {
                    commit(rev: $revision) {
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
        }
    `, { repo, revision }).toPromise().then(result => {
        if (!result.data ||
            !result.data.root.repository ||
            !result.data.root.repository.commit ||
            !result.data.root.repository.commit.commit ||
            !result.data.root.repository.commit.commit.tree ||
            !result.data.root.repository.commit.commit.tree.files
        ) {
            throw new Error('invalid response received from graphql endpoint');
        }
        const results = result.data.root.repository.commit.commit.tree.files;
        return results;
    });
    fileTreeCache.set(key, p);
    return p;
}

const localStorageKeyListAllFiles = 'listAllFiles';

interface LocalStorageListAllFiles {
    timestamp: number;
    data: any;
}

/**
 * Like listAllFiles, except the last repo@revision result is cached in
 * localstore. A call for a new repo@revision will wipe out the cache, i.e.
 * only a single repo@revision is stored in localstore at a time.
 */
export function localStoreListAllFiles(repo: string, revision: string): Promise<GQL.IFile[]> {
    // Uncomment to debug the non-cached path more easily:
    window.localStorage.setItem(localStorageKeyListAllFiles, '');

    const data = window.localStorage.getItem(localStorageKeyListAllFiles);
    if (data) {
        const d: LocalStorageListAllFiles = JSON.parse(data);
        const fiveMin = 5 * 60 * 1000; // 5m * 60s * 1000ms == 5m in milliseconds
        if (d.timestamp && (Date.now() - d.timestamp) < fiveMin) {
            // data exists and isn't stale.
            return Promise.resolve(d.data);
        }
    }

    // Fetch fresh data and store it.
    return listAllFiles(repo, revision).then(res => {
        const d: LocalStorageListAllFiles = {
            timestamp: Date.now(),
            data: res
        };
        window.localStorage.setItem(localStorageKeyListAllFiles, JSON.stringify(d));
        return res;
    });
}

export function fetchBlobHighlightContentTable(repoURI: string, rev: string, path: string): Promise<string> {
    return queryGraphQL(`
        query HighlightedBlobContent($repo: String, $rev: String, $path: String) {
            root {
                repository(uri: $repo) {
                    commit(rev: $rev) {
                        commit {
                            file(path: $path) {
                                highlightedContentHTML
                            }
                        }
                    }
                }
            }
        }
    `, { repo: repoURI, rev, path }).toPromise().then(result => {
        if (
            !result.data ||
            !result.data.root ||
            !result.data.root.repository ||
            !result.data.root.repository.commit ||
            !result.data.root.repository.commit.commit ||
            !result.data.root.repository.commit.commit.file
        ) {
            throw new Error(`cannot locate blob content: ${repoURI} ${rev} ${path}`);
        }
        return result.data.root.repository.commit.commit.file.highlightedContentHTML;
    });
}
