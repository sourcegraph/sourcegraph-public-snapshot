// tslint:disable: typedef ordered-imports

import {rel} from "sourcegraph/app/routePatterns";
import {Home} from "sourcegraph/home/Home";
import {IntegrationsContainer} from "sourcegraph/home/IntegrationsContainer";

export const routes: any[] = [
	{
		getComponent: (location, callback) => {
			callback(null, {
				main: Home,
				navContext: null,
			});
		},
		path: rel.newHome,
	},
	{
		getComponent: (location, callback) => {
			callback(null, {
				navContext: null,
				main: IntegrationsContainer,
			});
		},
		path: rel.integrations,
	},
];
