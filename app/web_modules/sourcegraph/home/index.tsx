// tslint:disable

import {rel} from "sourcegraph/app/routePatterns";
import {Route} from "react-router";

export const routes: any[] = [
	{
		getComponent: (location, callback) => {
			callback(null, {
				navContext: null,
				main: require("sourcegraph/home/IntegrationsContainer").default,
			});
		},
		path: rel.integrations,
	},
];
