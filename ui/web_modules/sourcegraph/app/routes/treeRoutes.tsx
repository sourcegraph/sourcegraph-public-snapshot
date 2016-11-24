import { RouterState } from "react-router";
import { rel } from "sourcegraph/app/routePatterns";
import { urlToRepo } from "sourcegraph/repo/routes";
import { urlToTree } from "sourcegraph/tree/routes";
import { TreeMain } from "sourcegraph/tree/TreeMain";

// canonicalizeTreeRoute redirects "/myrepo@myrev/-/tree/" to
// "/myrepo@myrev" and removes the slashes from
// "/myrepo@myrev/-/tree/subdir/".
function canonicalizeTreeRoute(nextRouterState: ReactRouter.RouterState, replace: Function): void {
	let {params} = nextRouterState as any;
	let path = params.splat[1];
	if (path === "") {
		replace(urlToRepo(params.splat[0]));
	} else if (path.endsWith("/")) {
		replace(nextRouterState, urlToTree(params.splat[0], "", path.replace(/\/+$/, "")));
	}
}

export const treeRoutes = [
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
			callback(null, { main: TreeMain });
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
