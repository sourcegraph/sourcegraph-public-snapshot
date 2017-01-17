import { formatPattern } from "react-router/lib/PatternUtils";

import { RouteParams, repoFromRouteParams } from "sourcegraph/app/router";
import { makeRepoRev, repoPath } from "sourcegraph/repo";
import { urlTo } from "sourcegraph/util/urlTo";

export function urlToRepo(repo: string): string {
	return urlTo("repo", { splat: repo });
}

export function urlToRepoRev(repo: string, rev: string | null): string {
	return urlTo("repo", { splat: makeRepoRev(repo, rev) });
}

/**
 * urlWithRev constructs a URL that is equivalent to the current URL (whose
 * current routes and routeParams are passed in), but pointing to a new rev. Only the
 * rev is overwritten in the returned URL.
 */
export function urlWithRev(routePattern: string, currentRouteParams: RouteParams, newRev: string | null): string {
	const repoRev = makeRepoRev(repoPath(repoFromRouteParams(currentRouteParams)), newRev);
	const newParams = Object.assign({}, currentRouteParams, {
		splat: currentRouteParams.splat instanceof Array ? [repoRev, ...currentRouteParams.splat.slice(1)] : repoRev,
	});
	return formatPattern(`/${routePattern}`, newParams);
}
