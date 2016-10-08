// tslint:disable: typedef ordered-imports

import {PlainRoute} from "react-router";
import {rel} from "sourcegraph/app/routePatterns";
import {SettingsMain} from "sourcegraph/user/settings/SettingsMain";

export const routes: PlainRoute[] = [
	{
		path: rel.settings,
		getComponent: (location, callback) => {
			callback(null, {navContext: null, main: SettingsMain} as any);
		},
	},
];
