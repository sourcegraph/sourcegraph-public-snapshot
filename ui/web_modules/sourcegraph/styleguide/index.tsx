// tslint:disable

import {rel} from "sourcegraph/app/routePatterns";
import {Route} from "react-router";
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
