import { getSourcegraphUrl } from "../utils/context";
import { doFetch as fetch } from "./xhr";

export const cacheKey = (repo: string, rev?: string) => `${repo}@${rev || null}`;

export interface ResolvedRevResp {
	notFound?: boolean;
	cloneInProgress?: boolean;
	commitID?: string;
}

const resolvedRevCache = new Map<string, Promise<ResolvedRevResp>>();

export function resolveRev(repo: string, rev?: string): Promise<ResolvedRevResp> {
	const key = cacheKey(repo, rev);
	const cacheHit = resolvedRevCache.get(key);
	if (cacheHit) {
		return cacheHit;
	}
	const p = fetch(`${getSourcegraphUrl()}/.api/graphql`, {
		method: "POST",
		headers: {
			"Content-Type": "application/json",
		},
		body: JSON.stringify({
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
		}),
	}).then((resp) => resp.json()).then((json: GQL.IGraphQLResponseRoot) => {
		if (!json.data) {
			throw new Error("invalid response received from graphql endpoint");
		}
		if (!json.data.root.repository) {
			return { notFound: true } as ResolvedRevResp;
		}
		if (json.data.root.repository.commit.cloneInProgress) {
			return { cloneInProgress: true } as ResolvedRevResp;
		} else if (!json.data.root.repository.commit.commit) {
			throw new Error("not able to resolve sha1");
		}
		return { commitID: json.data.root.repository.commit.commit.sha1 } as ResolvedRevResp;
	});
	resolvedRevCache.set(key, p);
	return p;
}
