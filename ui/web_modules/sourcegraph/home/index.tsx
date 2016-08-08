// tslint:disable

import {rel} from "sourcegraph/app/routePatterns";
import {Route} from "react-router";
import IntegrationsContainer from "sourcegraph/home/IntegrationsContainer";

export const routes: any[] = [
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
