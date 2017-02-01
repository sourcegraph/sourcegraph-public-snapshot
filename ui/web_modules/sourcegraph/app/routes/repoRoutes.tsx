import { PlainRoute } from "react-router";

import { rel } from "sourcegraph/app/routePatterns";
import { blobRoutes } from "sourcegraph/app/routes/blobRoutes";
import { treeRoutes } from "sourcegraph/app/routes/treeRoutes";
import { Workbench } from "sourcegraph/workbench/workbench";

// routes are the 2 routes needed for repos: the first is the one for repo
// subroutes, which must take precedence because the repo route matches
// greedily.
export const repoRoutes: PlainRoute[] = [
	{
		path: `${rel.repo}/-/`,
		getChildRoutes: (location, callback) => {
			callback(null, [...blobRoutes, ...treeRoutes]);
		},
	},
	{
		path: rel.repo,
		getIndexRoute: (location, callback) => callback(null, {
			getComponents: (loc, cb) => {
				cb(null, { main: Workbench });
			}
		}),
	},
];
