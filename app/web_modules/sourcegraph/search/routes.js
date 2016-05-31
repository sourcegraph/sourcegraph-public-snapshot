// @flow

import {rel} from "sourcegraph/app/routePatterns";
import type {Route} from "react-router";

export const routes: Array<Route> = [
	{
		path: rel.search,
		getComponent: (location, callback) => {
			require.ensure([], (require) => {
				callback(null, {
					main: require("sourcegraph/search/GlobalSearchContainer").default,
				});
			});
		},
	},
];

// urlToSearch returns the URL to a search results page, or
// (if query is falsey) the search form page.
export function urlToSearch(query?: string): string {
	if (!query) return `/${rel.search}`;
	return `/${rel.search}?q=${encodeURIComponent(query)}`;
}
