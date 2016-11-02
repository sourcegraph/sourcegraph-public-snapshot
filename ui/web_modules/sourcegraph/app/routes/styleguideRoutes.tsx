import {rel} from "sourcegraph/app/routePatterns";
import {StyleguideContainer} from "sourcegraph/styleguide/StyleguideContainer";

export const styleguideRoutes: any[] = [
	{
		getComponent: (location, callback) => {
			callback(null, {
				main: StyleguideContainer,
			});
		},
		path: rel.styleguide,
	},
];
