import { PlainRoute } from "react-router";
import { formatPattern } from "react-router/lib/PatternUtils";
import { makeRepoRev, repoParam, repoPath } from "sourcegraph/repo";
import { urlTo } from "sourcegraph/util/urlTo";

export function urlToRepo(repo: string): string {
	return urlTo("repo", { splat: repo });
}

export function urlToRepoRev(repo: string, rev: string | null): string {
	return urlTo("repo", { splat: makeRepoRev(repo, rev) });
}

// urlWithRev constructs a URL that is equivalent to the current URL (whose
// current routes and routeParams are passed in), but pointing to a new rev. Only the
// rev is overwritten in the returned URL.
export function urlWithRev(currentRoutes: PlainRoute[], currentRouteParams: any, newRev: string): string {
	const path = currentRoutes.map(r => r.path).join("");
	const repoRev = makeRepoRev(repoPath(repoParam(currentRouteParams.splat)), newRev);
	const newParams = Object.assign({}, currentRouteParams, {
		splat: currentRouteParams.splat instanceof Array ? [repoRev, ...currentRouteParams.splat.slice(1)] : repoRev,
	});
	return formatPattern(`/${path}`, newParams);
}
