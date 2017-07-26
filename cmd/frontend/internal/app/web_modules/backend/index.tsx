import { doFetch as fetch } from "app/backend/xhr";
import * as util from "app/util";

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
	const body = {
		query: `query ResolveRev($repo: String, $rev: String) {
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
				}`,
		variables: { repo, rev },
	};
	const p = fetch(`/.api/graphql?ResolveRev`, {
		method: "POST",
		body: JSON.stringify(body),
	}).then((resp) => resp.json()).then((json: any) => {
		// Note: only cache the promise if it is not found or found. If it is cloning, we want to recheck.
		promiseCache.delete(key);
		if (!json.data) {
			const error = new Error("invalid response received from graphql endpoint");
			promiseCache.set(key, Promise.reject(error));
			throw error;
		}
		if (!json.data.root.repository) {
			const notFound = { notFound: true };
			promiseCache.set(key, Promise.resolve(notFound));
			return notFound;
		}
		if (json.data.root.repository.commit.cloneInProgress) {
			// don't store this promise, we want to make a new query, holmes.
			return { cloneInProgress: true };
		} else if (!json.data.root.repository.commit.commit) {
			const error = new Error("not able to resolve sha1");
			promiseCache.set(key, Promise.reject(error));
			throw error;
		}
		const found = { commitID: json.data.root.repository.commit.commit.sha1 };
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
	const body = {
		query: `query BlobContent($repo: String, $rev: String, $path: String) {
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
				}`,
		variables: { repo: repoURI, rev, path },
	};
	const p = fetch(`/.api/graphql?BlobContent`, {
		method: "POST",
		body: JSON.stringify(body),
	}).then((resp) => resp.json()).then((json: any) => {
		if (!json.data || !json.data.root || !json.data.root.repository || !json.data.root.repository.commit || !json.data.root.repository.commit.commit || !json.data.root.repository.commit.commit.file) {
			throw new Error(`cannot locate blob content: ${key}`);
		}
		return json.data.root.repository.commit.commit.file.content;
	});
	blobCache.set(key, p);
	return p;
}

export const blameFileCacheKey = (repo: string, rev?: string, path: string, startLine: number, endLine: number) => `${repo}@${rev || null}/${path}#${startLine}-${endLine}`;
const blameFilePromiseCache = new Map<string, Promise<Hunk[]>>();

export interface Signature {
	person: Person;
	date: string;
}

export interface Person {
	name: string;
	email: string;
	gravatarHash: string;
}

export interface Hunk {
	startLine: number;
	endLine: number;
	startByte: number;
	endByte: number;
	rev: string;
	author: Signature;
	message: string;
}

export function blameFile(repo: string, rev?: string, path: string, startLine: number, endLine: number): Promise<Hunk[]> {
	const key = blameFileCacheKey(repo, rev, path, startLine, endLine);
	const promiseHit = blameFilePromiseCache.get(key);
	if (promiseHit) {
		return promiseHit;
	}
	const body = {
		query: `query BlameFile($repo: String, $rev: String, $path: String, $startLine: Int, $endLine: Int) {
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
}`,
		variables: { repo, rev, path, startLine, endLine },
	};
	const p = fetch(`/.api/graphql?BlameFile`, {
		method: "POST",
		body: JSON.stringify(body),
	}).then((resp) => resp.json()).then((json: any) => {
		blameFilePromiseCache.delete(key);
		if (!json.data ||
			!json.data.root ||
			!json.data.root.repository ||
			!json.data.root.repository.commit ||
			!json.data.root.repository.commit.commit ||
			!json.data.root.repository.commit.commit.file ||
			!json.data.root.repository.commit.commit.file.blame) {
			console.error("unexpected BlameFile response:", json);
			return null;
		}
		const found = json.data.root.repository.commit.commit.file.blame
		blameFilePromiseCache.set(key, Promise.resolve(found));
		return found;
	});
	blameFilePromiseCache.set(key, p);
	return p;
}

export function fetchDependencyReferences(repo: string, rev: string, path: string, line: number, character: number): Promise<any> {
	const mode = util.getModeFromExtension(util.getPathExtension(path));
	const body = {
		query: `
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
}`,
		variables: { repo, rev, mode, path, line, character },
	};
	const p = fetch(`/.api/graphql?DependencyReferences`, {
		method: "POST",
		body: JSON.stringify(body),
	}).then((resp) => resp.json()).then((json: any) => {
		// Note: only cache the promise if it is not found or found. If it is cloning, we want to recheck.
		const root = json.data.root;
		if (!root.repository ||
			!root.repository.commit ||
			!root.repository.commit.commit ||
			!root.repository.commit.commit.file ||
			!root.repository.commit.commit.file.dependencyReferences ||
			!root.repository.commit.commit.file.dependencyReferences.repoData ||
			!root.repository.commit.commit.file.dependencyReferences.dependencyReferenceData ||
			!root.repository.commit.commit.file.dependencyReferences.dependencyReferenceData.references.length) {
			return null;
		}

		return root.repository.commit.commit.file.dependencyReferences;
	});
	return p;
}

export interface ResolvedSearchTextResp {
	results?: [{ [key: string]: any }];
	notFound?: boolean;
}

const searchPromiseCache = new Map<string, Promise<ResolvedSearchTextResp>>();

export function searchText(uri: string, query: string): Promise<ResolvedSearchTextResp> {
	const key = cacheKey(uri, query);
	const promiseHit = searchPromiseCache.get(key);
	if (promiseHit) {
		return promiseHit;
	}
	const variables = {
		pattern: query,
		fileMatchLimit: 10000,
		isRegExp: false,
		isWordMatch: false,
		repositories: [{ repo: uri, rev: "" }],
		isCaseSensitive: false,
		includePattern: "",
		excludePattern: "{.git,**/.git,.svn,**/.svn,.hg,**/.hg,CVS,**/CVS,.DS_Store,**/.DS_Store,node_modules,bower_components,vendor,dist,out,Godeps,third_party}",
	};

	const body = {
		query: `query SearchText(
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
		}`,
		variables: variables,
	};

	const p = fetch(`/.api/graphql`, {
		method: "POST",
		body: JSON.stringify(body),
	}).then((resp) => resp.json()).then((json: any) => {
		searchPromiseCache.delete(key);
		const results = json.data && json.data.root!.searchRepos;
		if (!results) {
			const notFound = { notFound: true };
			promiseCache.set(key, Promise.resolve(notFound));
			return notFound;
		}

		searchPromiseCache.set(key, Promise.resolve(results));
		return results;
	});

	searchPromiseCache.set(key, p);
	return p;
}
