// tslint:disable

import {rel} from "sourcegraph/app/routePatterns";
import {Route} from "react-router";

export const routes: any[] = [
	{
		getComponent: (location, callback) => {
			callback(null, {
				main: require("sourcegraph/dashboard/DashboardContainer").default,
				navContext: null,
			});
		},
		path: rel.home,
	},
];

