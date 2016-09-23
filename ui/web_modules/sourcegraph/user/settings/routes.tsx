// tslint:disable: typedef ordered-imports

import {PlainRoute} from "react-router";
import {rel} from "sourcegraph/app/routePatterns";
import {UserSettingsReposMain} from "sourcegraph/user/settings/UserSettingsReposMain";

export const routes: PlainRoute[] = [
	{
		path: rel.settings,
		getComponent: (location, callback) => {
			callback(null, {navContext: null, main: UserSettingsReposMain} as any);
		},
	},
];
