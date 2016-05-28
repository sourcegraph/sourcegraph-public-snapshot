// @flow

import type {Route} from "react-router";
import {rel} from "sourcegraph/app/routePatterns";

export const routes: Array<Route> = [
	{
		path: rel.about,
		getComponents: (location, callback) => {
			require.ensure([], (require) => {
				callback(null, {
					main: require("sourcegraph/page/AboutPage").default,
				});
			});
		},
	},
	{
		path: rel.contact,
		getComponents: (location, callback) => {
			require.ensure([], (require) => {
				callback(null, {
					main: require("sourcegraph/page/ContactPage").default,
				});
			});
		},
	},
	{
		path: rel.security,
		getComponents: (location, callback) => {
			require.ensure([], (require) => {
				callback(null, {
					main: require("sourcegraph/page/SecurityPage").default,
				});
			});
		},
	},
];
