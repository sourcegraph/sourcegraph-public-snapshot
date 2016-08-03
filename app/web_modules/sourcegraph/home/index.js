import {rel} from "sourcegraph/app/routePatterns";
import type {Route} from "react-router";

export const routes: Array<Route> = [
	{
		getComponent: (location, callback) => {
			require.ensure([], (require) => {
				callback(null, {
					navContext: null,
					main: require("sourcegraph/home/IntegrationsContainer").default,
				});
			});
		},
		path: rel.integrations,
	},
];
