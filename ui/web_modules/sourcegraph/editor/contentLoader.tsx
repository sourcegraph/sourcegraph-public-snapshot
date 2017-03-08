import URI from "vs/base/common/uri";
import { TPromise } from "vs/base/common/winjs.base";

import { URIUtils } from "sourcegraph/core/uri";
import { fetchGraphQLQuery } from "sourcegraph/util/GraphQLFetchUtil";

export function fetchContent(resource: URI): TPromise<string> {
	const resourceKey = resource.toString();
	if (contentCache.has(resourceKey)) {
		return TPromise.wrap(contentCache.get(resource.toString()));
	}
	const { repo, rev, path } = URIUtils.repoParams(resource);
	return TPromise.wrap(fetchGraphQLQuery(`query Content($repo: String, $rev: String, $path: String) {
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
		}`, { repo, rev, path })
		.then((resp: GQL.IQuery) => {
			if (!resp.root || !resp.root.repository || !resp.root.repository.commit.commit || !resp.root.repository.commit.commit.file) {
				throw new Error("File content not available.");
			}
			return resp.root.repository.commit.commit.file.content;
		}));
}

// The content cache can be imported and used to dynamically modify
// available files.
export const contentCache = new Map<string, string>();

/**
 * fetchContentAndResolveRev fetches the absolute revision SHA for a commit string,
 * like "myBranch" or "" (default), as well as file contents at that revision for
 * the specified resource. If you just need to resolve a revision string, use
 * `resolveRev(resource)` instead.
 */
export async function fetchContentAndResolveRev(resource: URI, isViewingZapRef?: boolean): Promise<{ content: string, commit: string }> {
	const { repo, rev, path } = URIUtils.repoParams(resource);
	const resp = await fetchGraphQLQuery(`query Content($repo: String, $rev: String, $path: String) {
			root {
				repository(uri: $repo) {
					commit(rev: $rev) {
						commit {
							file(path: $path) {
								content
							}
							sha1
						}
					}
				}
			}
		}`, { repo, rev, path });
	if (!resp.root || !resp.root.repository || !resp.root.repository.commit.commit) {
		throw new Error("File content not available.");
	}
	const commit = resp.root.repository.commit.commit.sha1;
	const resourceKey = resource.toString();
	let content: string;
	if (contentCache.has(resourceKey) && isViewingZapRef) {
		// only serve cached resources for zap sessions
		content = contentCache.get(resourceKey)!;
	} else {
		if (!resp.root.repository.commit.commit.file) {
			throw new Error("File content not available.");
		}
		content = resp.root.repository.commit.commit.file.content;
		contentCache.set(resourceKey, content);
	}
	return {
		content,
		commit,
	};
}

/**
 * resolveRev fetches the absolute revision SHA matching a commit string,
 * like "myBranch" or "" (default).
 */
export async function resolveRev(resource: URI): Promise<{ commit: string }> {
	const { repo, rev } = URIUtils.repoParams(resource);
	const resp = await fetchGraphQLQuery(`query ResolveRev($repo: String, $rev: String) {
			root {
				repository(uri: $repo) {
					commit(rev: $rev) {
						commit {
							sha1
						}
					}
				}
			}
		}`, { repo, rev });
	if (!resp.root || !resp.root.repository || !resp.root.repository.commit.commit) {
		throw new Error("Revision not available.");
	}
	const commit = resp.root.repository.commit.commit.sha1;
	return {
		commit,
	};
}
