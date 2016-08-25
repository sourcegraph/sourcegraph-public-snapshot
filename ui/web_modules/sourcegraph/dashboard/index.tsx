// tslint:disable: typedef ordered-imports

import {rel} from "sourcegraph/app/routePatterns";
import {DashboardContainer} from "sourcegraph/dashboard/DashboardContainer";
import { withRouter } from "react-router";

export const routes: any[] = [
	{
		getComponent: (location, callback) => {
			callback(null, {
				main: withRouter(DashboardContainer),
				navContext: null,
			});
		},
		path: rel.dashboard,
	},
];
