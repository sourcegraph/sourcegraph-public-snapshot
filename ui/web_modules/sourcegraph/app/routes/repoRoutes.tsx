import { rel } from "sourcegraph/app/routePatterns";

import { blobRoutes } from "sourcegraph/app/routes/blobRoutes";
import { treeRoutes } from "sourcegraph/app/routes/treeRoutes";

// routes are the 2 routes needed for repos: the first is the one for repo
// subroutes, which must take precedence because the repo route matches
// greedily.
export const repoRoutes: any[] = [
	{
		path: `${rel.repo}/-/`,
		getChildRoutes: (location, callback) => {
			callback(null, [
				...blobRoutes,
				...treeRoutes,
			]);
		},
	},
	{
		path: rel.repo,
		indexRoute: {
			keepScrollPositionOnRouteChangeKey: "tree",
			getComponents: (location, callback) => {
				(treeRoutes[0] as any).getComponents(location, callback);
			},
		},
	},
];
