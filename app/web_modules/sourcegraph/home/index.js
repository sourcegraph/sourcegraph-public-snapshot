import {rel} from "sourcegraph/app/routePatterns";
import type {Route} from "react-router";

export const integrations = {
	getComponent: (location, callback) => {
		require.ensure([], (require) => {
			callback(null, {
				globalNav: null,
				navContext: null,
				main: require("sourcegraph/home/ToolsContainer").default,
			});
		});
	},
};

export const routes: Array<Route> = [
	{
		...integrations,
		path: rel.integrations,
	},
];
