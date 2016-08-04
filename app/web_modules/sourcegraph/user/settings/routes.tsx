// tslint:disable

import {rel} from "sourcegraph/app/routePatterns";
import {Route} from "react-router";

export const routes: ReactRouter.PlainRoute[] = [
	{
		path: rel.settings,
		getComponent: (location, callback) =>
			require("sourcegraph/user/settings/SettingsMain")
			.then(m => callback(null, {navContext: null, main: m.default} as any)),
		childRoutes: [
			{
				path: rel.settingsRepos,
				getComponent: (location, callback) =>
					require("sourcegraph/user/settings/UserSettingsReposMain")
					.then(m => callback(null, {navContext: null, main: m.default} as any)),
			},
		],
	},
];
