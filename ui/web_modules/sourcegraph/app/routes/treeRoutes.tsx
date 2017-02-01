import { PlainRoute } from "react-router";

import { rel } from "sourcegraph/app/routePatterns";
import { pathFromRouteParams, repoRevFromRouteParams } from "sourcegraph/app/router";
import { urlToRepo } from "sourcegraph/repo/routes";
import { urlToTree } from "sourcegraph/tree/routes";
import { Workbench } from "sourcegraph/workbench/workbench";

// canonicalizeTreeURL redirects "/myrepo@myrev/-/tree/" to
// "/myrepo@myrev" and removes the slashes from
// "/myrepo@myrev/-/tree/subdir/".
function canonicalizeTreeURL(nextState: ReactRouter.RouterState, replace: Function): void {
	const repoRev = repoRevFromRouteParams(nextState.params as any);
	const path = pathFromRouteParams(nextState.params as any);
	if (path === "") {
		replace(urlToRepo(repoRev));
	} else if (path.endsWith("/")) {
		replace(nextState, urlToTree(repoRev, "", path.replace(/\/+$/, "")));
	}
}

export const treeRoutes: PlainRoute[] = [
	{
		path: rel.tree,
		onEnter: (nextState, replace) => {
			canonicalizeTreeURL(nextState, replace);
		},
		getComponents: (location, callback) => {
			callback(null, { main: Workbench });
		},
		getIndexRoute: (location, callback) => callback(null, {
			getComponents: (loc, cb) => {
				cb(null, { main: Workbench });
			}
		})
	},

	// Redirect "/myrepo@myrev/-/tree" to "/myrepo@mrev".
	{
		path: rel.tree.replace("/*", ""),
		onEnter: (nextState, replace) => {
			replace(urlToRepo(nextState.params["splat"]));
		},
	},

];
