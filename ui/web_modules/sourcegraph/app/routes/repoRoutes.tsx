import { rel } from "sourcegraph/app/routePatterns";
import { NavContext } from "sourcegraph/repo/NavContext";
import { RepoMain } from "sourcegraph/repo/RepoMain";
import { withResolvedRepoRev } from "sourcegraph/repo/withResolvedRepoRev";

import { blobRoutes } from "sourcegraph/app/routes/blobRoutes";
import { treeRoutes } from "sourcegraph/app/routes/treeRoutes";

let _components;

const getComponents = (location, callback) => {
	if (!_components) {
		_components = {
			navContext: withResolvedRepoRev(NavContext, false),
			main: withResolvedRepoRev(RepoMain, true),
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
export const repoRoutes: any[] = [
	{
		getComponents: getComponents,
		path: `${rel.repo}/-/`,
		getChildRoutes: (location, callback) => {
			callback(null, [
				...blobRoutes,
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
