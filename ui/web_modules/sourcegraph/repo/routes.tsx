// tslint:disable

import {rel} from "sourcegraph/app/routePatterns";
import {urlTo} from "sourcegraph/util/urlTo";
import {makeRepoRev, repoPath, repoParam} from "sourcegraph/repo/index";
import {Route} from "react-router";
import {formatPattern} from "react-router/lib/PatternUtils";
import {withResolvedRepoRev} from "sourcegraph/repo/withResolvedRepoRev";
import {withRepoBuild} from "sourcegraph/build/withRepoBuild";
import {NavContext} from "sourcegraph/repo/NavContext";
import {RepoMain} from "sourcegraph/repo/RepoMain";

import {routes as blobRoutes} from "sourcegraph/blob/routes";
import {routes as buildRoutes} from "sourcegraph/build/routes";
import {routes as defRoutes} from "sourcegraph/def/routes";
import {routes as treeRoutes} from "sourcegraph/tree/routes";

let _components;

const getComponents = (location, callback) => {
	if (!_components) {
		_components = {
			navContext: withResolvedRepoRev(NavContext, false),
			main: withResolvedRepoRev(withRepoBuild(RepoMain), true),
		};
	}
	callback(null, {
		main: _components.main,

		// Allow disabling the nav context on a per-route basis.
		navContext: location.routes[location.routes.length - 1].repoNavContext === false ? null : _components.navContext,
	});
};

// routes are the 2 routes needed for repos: the first is the one for repo
// subroutes, which must take precedence because the repo route matches
// greedily.
export const routes: any[] = [
	{
		getComponents: getComponents,
		path: `${rel.repo}/-/`,
		getChildRoutes: (location, callback) => {
			callback(null, [
				...blobRoutes,
				...buildRoutes,
				...defRoutes,
				...treeRoutes,
			]);
		},
	},
	{
		getComponents: getComponents,
		path: rel.repo,
		indexRoute: {
			keepScrollPositionOnRouteChangeKey: "tree",
			getComponents: (location, callback) => {
				(treeRoutes[0] as any).getComponents(location, callback);
			},
		},
	},
];

export function urlToRepo(repo: string): string {
	return urlTo("repo", {splat: repo});
}

export function urlToRepoRev(repo: string, rev: string | null): string {
	return urlTo("repo", {splat: makeRepoRev(repo, rev)});
}

// urlWithRev constructs a URL that is equivalent to the current URL (whose
// current routes and routeParams are passed in), but pointing to a new rev. Only the
// rev is overwritten in the returned URL.
export function urlWithRev(currentRoutes: ReactRouter.PlainRoute[], currentRouteParams: any, newRev: string): string {
	// Ensure this is a repo subroute. If not, it's meaningless to change the rev.
	// The 0th route is the rootRoute; the next should be one of the repo base routes.
	if (!currentRoutes[1] || !routes.includes(currentRoutes[1])) {
		throw new Error("can't overwrite rev for non-repo routes");
	}

	const path = currentRoutes.map(r => r.path).join("");
	const repoRev = makeRepoRev(repoPath(repoParam(currentRouteParams.splat)), newRev);
	const newParams = Object.assign({}, currentRouteParams, {
		splat: currentRouteParams.splat instanceof Array ? [repoRev, ...currentRouteParams.splat.slice(1)] : repoRev,
	});
	return formatPattern(`/${path}`, newParams);
}
