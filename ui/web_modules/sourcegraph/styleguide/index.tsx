// tslint:disable: typedef ordered-imports

import {rel} from "sourcegraph/app/routePatterns";
import {StyleguideContainer} from "sourcegraph/styleguide/StyleguideContainer";

export const routes: any[] = [
	{
		getComponent: (location, callback) => {
			callback(null, {
				main: StyleguideContainer,
			});
		},
		path: rel.styleguide,
	},
];
