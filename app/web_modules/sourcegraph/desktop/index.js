import {rel} from "sourcegraph/app/routePatterns";
import type {Route} from "react-router";


export const desktopHome = {
	getComponent: (location, callback) => {
		require.ensure([], (require) => {
			callback(null, {
				main: require("sourcegraph/desktop/DesktopHome").default,
			});
		});
	},
};

export const routes: Array<Route> = [
	{
		...desktopHome,
		path: rel.desktopHome,
	},
];
