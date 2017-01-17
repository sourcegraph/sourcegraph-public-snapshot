import { PlainRoute } from "react-router";

import { rel } from "sourcegraph/app/routePatterns";
import { StyleguideContainer } from "sourcegraph/styleguide/StyleguideContainer";

export const styleguideRoutes: PlainRoute[] = [
	{
		getComponents: (location, callback) => {
			callback(null, { main: StyleguideContainer });
		},
		path: rel.styleguide,
	},
];
