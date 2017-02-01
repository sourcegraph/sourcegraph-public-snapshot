import { PlainRoute } from "react-router";

import { abs } from "sourcegraph/app/routePatterns";
import { Workbench } from "sourcegraph/workbench/workbench";

export const symbolRoutes: PlainRoute[] = [
	{
		path: abs.goSymbol,
		getIndexRoute: (location, callback) => callback(null, {
			getComponents: (loc, cb) => {
				cb(null, { main: Workbench });
			}
		}),
	},
];
