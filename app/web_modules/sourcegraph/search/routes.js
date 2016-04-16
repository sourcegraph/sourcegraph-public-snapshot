// @flow weak

import type {Route} from "react-router";
import {rel} from "sourcegraph/app/routePatterns";

export const routes: Array<Route> = [
	{
		path: rel.globalSearch,
		getComponents: (location, callback) => {
			require.ensure([], (require) => {
				callback(null, {
					main: require("sourcegraph/search/GlobalSearch").default,
				});
			});
		},
	},
];
