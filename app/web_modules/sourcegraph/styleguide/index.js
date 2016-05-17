import {rel} from "sourcegraph/app/routePatterns";
import type {Route} from "react-router";

export const styleguide = {
	getComponent: (location, callback) => {
		require.ensure([], (require) => {
			callback(null, {
				main: require("sourcegraph/styleguide/StyleguideContainer").default,
			});
		});
	},
};


export const routes: Array<Route> = [
	{
		...styleguide,
		path: rel.styleguide,
	},
];

