// tslint:disable

import {rel} from "sourcegraph/app/routePatterns";
import {Route} from "react-router";

export const routes: any[] = [
	{
		getComponent: (location, callback) => {
			require.ensure([], (require) => {
				callback(null, {
					main: require("sourcegraph/styleguide/StyleguideContainer").default,
				});
			});
		},
		path: rel.styleguide,
	},
];
