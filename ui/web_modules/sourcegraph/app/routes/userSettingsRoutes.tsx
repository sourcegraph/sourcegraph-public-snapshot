import { PlainRoute } from "react-router";
import { rel } from "sourcegraph/app/routePatterns";
import { SettingsMain } from "sourcegraph/user/settings/SettingsMain";
import { Workbench } from "sourcegraph/workbench/workbench";

export const userSettingsRoutes: PlainRoute[] = [
	{
		path: rel.settings,
		getComponents: (location, callback) => {
			callback(null, { main: Workbench, injectedComponent: SettingsMain });
		},
	},
	{
		path: rel.orgSettings,
		getComponents: (location, callback) => {
			callback(null, { main: Workbench, injectedComponent: SettingsMain });
		},
	},
];
