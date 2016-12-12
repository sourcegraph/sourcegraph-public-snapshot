import URI from "vs/base/common/uri";
import { TPromise } from "vs/base/common/winjs.base";

import { URIUtils } from "sourcegraph/core/uri";
import { singleflightFetch } from "sourcegraph/util/singleflightFetch";
import { checkStatus, defaultFetch } from "sourcegraph/util/xhr";

const fetch = singleflightFetch(defaultFetch);

export function getContent(resource: URI): TPromise<string> {
	const {repo, rev, path} = URIUtils.repoParams(resource);
	return fetch(`/.api/graphql`, {
		method: "POST",
		headers: {
			"Content-Type": "application/json",
		},
		body: JSON.stringify({
			query: `query Content($repo: String, $rev: String, $path: String) {
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
			variables: { repo, rev, path },
		}),
	})
		.then(checkStatus)
		.then(resp => resp.json())
		.then((resp: GQL.IGraphQLResponseRoot) => {
			if (!resp.data || !resp.data.root.repository || !resp.data.root.repository.commit.commit || !resp.data.root.repository.commit.commit.file) {
				throw new Error("file content not available");
			}
			return resp.data.root.repository.commit.commit.file.content;
		})
		.catch(err => err);
}
