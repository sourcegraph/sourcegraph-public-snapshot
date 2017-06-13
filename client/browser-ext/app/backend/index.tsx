import { sourcegraphUrl } from "../utils/context";
import { doFetch as fetch } from "./xhr";

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
		query: `query Content($repo: String, $rev: String) {
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
	const p = fetch(`${sourcegraphUrl}/.api/graphql`, {
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

export interface ResolvedSearchTextResp {
	results: [{ [key: string]: any }];
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
		isRegExp: false,
		isCaseSensitive: false,
		isWordMatch: false,
		wordSeparators: "`~!@#$%^&*()-=+[{]}\\|;:'\",.<>/?",
		revision: "",
		uri: uri,
		fileMatchLimit: 10000,
		includePattern: "",
		excludePattern: "{.git,**/.git,.svn,**/.svn,.hg,**/.hg,CVS,**/CVS,.DS_Store,**/.DS_Store,node_modules,bower_components,vendor,dist,out,Godeps,third_party}",
	};

	const body = {
		query: `query SearchText($uri: String!, $pattern: String!, $revision: String!, $isRegExp: Boolean!, $isWordMatch: Boolean!, $isCaseSensitive: Boolean!, $fileMatchLimit: Int!, $includePattern: String!, $excludePattern: String!) {
			root {
				repository(uri: $uri) {
					commit(rev: $revision) {
						commit {
							textSearch(query: { pattern: $pattern, isRegExp: $isRegExp, isWordMatch: $isWordMatch, isCaseSensitive: $isCaseSensitive, fileMatchLimit: $fileMatchLimit, includePattern: $includePattern, excludePattern: $excludePattern }) {
								results {
									resource
									lineMatches {
										preview
										lineNumber
										offsetAndLengths
									}
								}
							}
						}
					}
				}
			}
			}`,
		variables: variables,
	};

	const p = fetch(`${sourcegraphUrl}/.api/graphql`, {
		method: "POST",
		body: JSON.stringify(body),
	}).then((resp) => resp.json()).then((json: GQL.IGraphQLResponseRoot) => {
		searchPromiseCache.delete(key);
		const repo = json.data && json.data.root!.repository;
		if (!repo || !repo.commit.commit || !repo.commit.commit.textSearch) {
			const error = new Error("invalid response received from search graphql endpoint");
			searchPromiseCache.set(key, Promise.reject(error));
			throw error;
		}
		const found = repo.commit.commit.textSearch;
		searchPromiseCache.set(key, Promise.resolve(found));
		return found;
	});

	searchPromiseCache.set(key, p);
	return p;
}
