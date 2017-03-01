import { sourcegraphUrl } from "../utils/context";
import { doFetch as fetch } from "./xhr";

export const cacheKey = (repo: string, rev?: string) => `${repo}@${rev || null}`;

export interface ResolvedRevResp {
	notFound?: boolean;
	cloneInProgress?: boolean;
	commitID?: string;
}

const promiseCache = new Map<string, Promise<ResolvedRevResp>>();
const inflightPromises = new Map<string, Promise<ResolvedRevResp>>();

export function resolveRev(repo: string, rev?: string): Promise<ResolvedRevResp> {
	const key = cacheKey(repo, rev);
	const promiseHit = promiseCache.get(key);
	if (promiseHit) {
		return promiseHit;
	}
	const inflightPromise = inflightPromises.get(key);
	if (inflightPromise) {
		return inflightPromise;
	}
	const p = fetch(`${sourcegraphUrl}/.api/graphql`, {
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
		// Note: only cache the promise if it is not found or found. If it is cloning, we want to recheck.
		inflightPromises.delete(key);
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
	inflightPromises.set(key, p);
	return p;
}
