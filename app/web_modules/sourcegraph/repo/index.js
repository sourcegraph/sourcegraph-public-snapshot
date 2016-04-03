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

export function makeRepoRev(repo: string, rev: string): string {
	if (rev) return `${repo}@${rev}`;
	return repo;
}

export function repoParam(splat: string[] | string): string {
	return splat instanceof Array ? splat[0] : splat;
}
