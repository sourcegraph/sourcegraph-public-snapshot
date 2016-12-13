import URI from "vs/base/common/uri";
import { TPromise } from "vs/base/common/winjs.base";

import { URIUtils } from "sourcegraph/core/uri";
import { fetchGraphQLQuery } from "sourcegraph/util/GraphQLFetchUtil";

export function fetchContent(resource: URI): TPromise<string> {
	const {repo, rev, path} = URIUtils.repoParams(resource);
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
				throw new Error("file content not available");
			}
			return resp.root.repository.commit.commit.file.content;
		})
		.catch(err => err));
}
