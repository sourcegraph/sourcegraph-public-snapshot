import { PlainRoute } from "react-router";

import { rel } from "sourcegraph/app/routePatterns";
import { urlToRepo } from "sourcegraph/repo/routes";
import { Workbench } from "sourcegraph/workbench/workbench";

export const blobRoutes: PlainRoute[] = [
	{
		path: rel.blob,
		getComponents: (location, callback) => {
			callback(null, { main: Workbench });
		},
	},

	// Redirect "/myrepo@myrev/-/blob" to "/myrepo@mrev".
	{
		path: rel.blob.replace("/*", ""),
		onEnter: (nextState, replace) => {
			replace(urlToRepo(nextState.params["splat"]));
		},
	},
];
