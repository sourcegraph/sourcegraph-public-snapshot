import {doFetch as fetch} from "./actions/xhr";

export const cacheKey = (repo: string, rev?: string) => `${repo}@${rev || null}`;

export interface ResolvedRevResp {
	authRequired?: boolean;
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

	return fetch(`https://sourcegraph.com/.api/repos/${repo}${rev ? `@${rev}` : ""}/-/rev`)
		.then((resp) => {
			if (resp.status === 404) {
				return {authRequired: true};
			}
			if (resp.status === 202) {
				return {cloneInProgress: true};
			}

			const p = resp.json().then((json) => ({commitID: json.CommitID}));
			resolvedRevCache.set(key, p);
			return p;
		});
}

const createdRepoOnce = new Set<string>();

export function ensureRepoExists(repo: string): void {
	if (createdRepoOnce.has(repo)) {
		return;
	}

	const body = {
		Op: {
			New: {
				URI: repo,
				CloneURL: `https://${repo}`,
				DefaultBranch: "master",
				Mirror: true,
			},
		},
	};

	fetch("https://sourcegraph.com/.api/repos?AcceptAlreadyExists=true", {method: "POST", body: JSON.stringify(body)});
	createdRepoOnce.add(repo);
}
