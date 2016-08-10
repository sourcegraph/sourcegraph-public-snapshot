// tslint:disable: typedef ordered-imports curly

import {rel} from "sourcegraph/app/routePatterns";
import {BuildsList} from "sourcegraph/build/BuildsList";

const globalBuilds: ReactRouter.PlainRoute = {
	path: rel.builds,
	onEnter: (nextState, replace) => {
		if (nextState.location.search === "") {
			replace(`${nextState.location.pathname}?filter=all`);
		}
	},
	getComponents: (location, callback) => {
		callback(null, {
			main: BuildsList,
		});
	},
};

export const routes: ReactRouter.PlainRoute[] = [
	{
		path: rel.admin,
		getChildRoutes: (location, callback) => {
			callback(null, [
				globalBuilds,
			]);
		},
	},
];
