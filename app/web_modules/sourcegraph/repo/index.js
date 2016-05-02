// @flow

// repoPath returns the path portion of a repo route var match.
export function repoPath(repoRevRouteVar: string): string {
	const at = repoRevRouteVar.indexOf("@");
	if (at === -1) return repoRevRouteVar;
	return repoRevRouteVar.slice(0, at);
}

// repoRev returns the rev portion of a repo route var match, or
// null if there is none.
export function repoRev(repoRevRouteVar: string): ?string {
	const at = repoRevRouteVar.indexOf("@");
	if (at === -1 || at === repoRevRouteVar.length - 1) return null;
	return repoRevRouteVar.slice(at + 1);
}

// makeRepoRev returns "<repo>@<rev>" if rev is a non-empty string, otherwise
// it returns just "<repo>".
export function makeRepoRev(repo: string, rev: string): string {
	if (rev) return `${repo}@${rev}`;
	return repo;
}

export function repoParam(splat: string[] | string): string {
	return splat instanceof Array ? splat[0] : splat;
}

export function trimRepo(repo: string): string {
	let res = repo;
	if (res.indexOf("github.com/") !== -1) {
		res = res.substring("github.com/".length);
	}
	if (res.indexOf("sourcegraph.com/") !== -1) {
		res = res.substring("sourcegraph.com/".length);
	}

	return res;
}
