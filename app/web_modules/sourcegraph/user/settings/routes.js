// @flow

import {rel} from "sourcegraph/app/routePatterns";
import type {Route} from "react-router";

export const routes: Array<Route> = [
	{
		path: rel.settingsRepos,
		getComponent: (location, callback) => {
			require.ensure([], (require) => {
				callback(null, {
					main: require("sourcegraph/user/settings/UserSettingsReposMain").default,
				});
			});
		},
	},
];
