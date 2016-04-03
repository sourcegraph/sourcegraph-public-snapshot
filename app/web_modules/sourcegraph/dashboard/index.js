import {rel} from "sourcegraph/app/routePatterns";

export const route = {
	path: rel.dashboard,
	getComponent: (location, callback) => {
		require.ensure([], (require) => {
			callback(null, {
				main: require("sourcegraph/dashboard/DashboardContainer").default,
			});
		});
	},
};
