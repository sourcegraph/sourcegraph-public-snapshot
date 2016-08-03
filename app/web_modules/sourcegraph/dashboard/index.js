import {rel} from "sourcegraph/app/routePatterns";
import type {Route} from "react-router";

export const routes: Array<Route> = [
	{
		getComponent: (location, callback) => {
			require.ensure([], (require) => {
				callback(null, {
					main: require("sourcegraph/dashboard/DashboardContainer").default,
					navContext: null,
				});
			});
		},
		path: rel.home,
	},
];

