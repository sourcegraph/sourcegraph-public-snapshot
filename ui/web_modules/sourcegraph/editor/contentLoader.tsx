import URI from "vs/base/common/uri";

import { fetchGQL } from "sourcegraph/util/gqlClient";
import { getURIContext } from "sourcegraph/workbench/utils";

// The content cache can be imported and used to dynamically modify
// available files.
export const contentCache = new Map<string, string>();

/**
 * fetchContentAndResolveRev fetches the absolute revision SHA for a commit string,
 * like "myBranch" or "" (default), as well as file contents at that revision for
 * the specified resource. If you just need to resolve a revision string, use
 * `resolveRev(resource)` instead.
 */
export async function fetchContentAndResolveRev(resource: URI): Promise<{ content: string, commit: string }> {
	const { repo, rev, path } = getURIContext(resource);
	const resp = await fetchGQL(`query FileContentAndRev($repo: String, $rev: String, $path: String) {
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
	const root = resp.data.root;
	if (!root || !root.repository || !root.repository.commit.commit) {
		throw new Error("File content not available.");
	}
	const commit = root.repository.commit.commit.sha1;
	const resourceKey = resource.toString();
	let content: string;
	if (!root.repository.commit.commit.file) {
		throw new Error("File content not available.");
	}
	content = root.repository.commit.commit.file.content;
	contentCache.set(resourceKey, content);
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
	const { repo, rev } = getURIContext(resource);
	const resp = await fetchGQL(`query ResolveRev($repo: String, $rev: String) {
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
	const root = resp.data.root;
	if (!root || !root.repository || !root.repository.commit.commit) {
		throw new Error("Revision not available.");
	}
	const commit = root.repository.commit.commit.sha1;
	return {
		commit,
	};
}
