// tslint:disable: typedef ordered-imports

import {urlTo} from "sourcegraph/util/urlTo";
import {makeRepoRev} from "sourcegraph/repo";
import {rel} from "sourcegraph/app/routePatterns";
import {urlToRepo, urlToRepoRev} from "sourcegraph/repo/routes";
import {RouterState} from "react-router";
import {withResolvedRepoRev} from "sourcegraph/repo/withResolvedRepoRev";
import {withTree} from "sourcegraph/tree/withTree";
import {TreeMain} from "sourcegraph/tree/TreeMain";

let _components;

// canonicalizeTreeRoute redirects "/myrepo@myrev/-/tree/" to
// "/myrepo@myrev" and removes the slashes from
// "/myrepo@myrev/-/tree/subdir/".
function canonicalizeTreeRoute(nextRouterState, replace) {
	let {params} = nextRouterState;
	let path = params.splat[1];
	if (path === "") {
		replace(urlToRepo(params.splat[0]));
	} else if (path.endsWith("/")) {
		replace(nextRouterState, urlToTree(params.splat[0], "", path.replace(/\/+$/, "")));
	}
}

export const routes = [
	{
		path: rel.tree,
		keepScrollPositionOnRouteChangeKey: "tree",
		onEnter: (nextRouterState: RouterState, replace: Function) => {
			canonicalizeTreeRoute(nextRouterState, replace);
		},
		onChange: (prevRouterState: RouterState, nextRouterState: RouterState, replace: Function) => {
			canonicalizeTreeRoute(nextRouterState, replace);
		},
		getComponents: (location: Location, callback: Function) => {
			if (!_components) {
				_components = {
					main: withResolvedRepoRev(withTree(TreeMain)),
				};
			}
			callback(null, _components);
		},
	},

	// Redirect "/myrepo@myrev/-/tree" to "/myrepo@mrev".
	{
		path: rel.tree.replace("/*", ""),
		onEnter: (nextRouterState: any, replace: Function) => {
			replace(urlToRepo(nextRouterState.params.splat));
		},
	},
];

// urlToTree generates the URL to a dir. To get a file's URL, use urlToBlob.
export function urlToTree(repo: string, rev: string | null, path: string | string[]): string {
	rev = rev || "";

	// Fast-path: we redirect the tree root to the repo route anyway, so just construct
	// the repo route URL directly.
	if (!path || path === "/" || path.length === 0) {
		return urlToRepoRev(repo, rev);
	}

	const pathStr = typeof path === "string" ? path : path.join("/");
	return urlTo("tree", {splat: [makeRepoRev(repo, rev), pathStr]} as any);
}
