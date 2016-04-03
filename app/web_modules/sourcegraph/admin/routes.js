// @flow weak

import type {Route} from "react-router";
import {rel} from "sourcegraph/app/routePatterns";

const globalBuilds: Route = {
	path: rel.builds,
	getComponents: (location, callback) => {
		require.ensure([], (require) => {
			callback(null, {
				main: require("sourcegraph/build/BuildsList").default,
			});
		});
	},
};

export const routes: Array<Route> = [
	{
		path: rel.admin,
		getChildRoutes: (location, callback) => {
			require.ensure([], (require) => {
				callback(null, [
					globalBuilds,
				]);
			});
		},
	},
];
