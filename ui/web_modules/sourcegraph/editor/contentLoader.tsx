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

const contentCache = new Map<string, string>();

export async function fetchContentAndResolveRev(resource: URI): Promise<{ content: string, commit: string }> {
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
	if (!resp.root || !resp.root.repository || !resp.root.repository.commit.commit || !resp.root.repository.commit.commit.file) {
		throw new Error("File content not available.");
	}
	const commit = resp.root.repository.commit.commit.sha1;
	const content = resp.root.repository.commit.commit.file.content;
	const resourceKey = resource.with({ query: commit }).toString();
	contentCache.set(resourceKey, content);
	return {
		content,
		commit,
	};
}
