import {rel} from "sourcegraph/app/routePatterns";
import type {Route} from "react-router";

export const dashboard = {
	getComponent: (location, callback) => {
		require.ensure([], (require) => {
			callback(null, {
				main: require("sourcegraph/dashboard/DashboardContainer").default,
				navContext: null,
			});
		});
	},
};


export const routes: Array<Route> = [
	{
		...dashboard,
		path: rel.home,
	},
];

