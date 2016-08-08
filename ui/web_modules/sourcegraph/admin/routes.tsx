// tslint:disable

import {Route} from "react-router";
import {rel} from "sourcegraph/app/routePatterns";
import BuildsList from "sourcegraph/build/BuildsList";
import CoverageDashboard from "sourcegraph/admin/CoverageDashboard";

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

const coverage: ReactRouter.PlainRoute = {
	path: rel.coverage,
	getComponents: (location, callback) => {
		callback(null, {main: CoverageDashboard});
	},
};

export const routes: ReactRouter.PlainRoute[] = [
	{
		path: rel.admin,
		getChildRoutes: (location, callback) => {
			callback(null, [
				globalBuilds,
				coverage,
			]);
		},
	},
];
