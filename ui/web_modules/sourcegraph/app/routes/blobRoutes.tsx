import { PlainRoute } from "react-router";

import { rel } from "sourcegraph/app/routePatterns";
import { Workbench } from "sourcegraph/workbench/workbench";

export const blobRoutes: PlainRoute[] = [
	{
		path: rel.blob,
		getComponents: (location, callback) => {
			callback(null, { main: Workbench });
		},
	}
];
