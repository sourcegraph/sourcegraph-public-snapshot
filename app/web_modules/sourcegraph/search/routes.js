// @flow

import {rel} from "sourcegraph/app/routePatterns";
import type {Route, RouterLocation} from "react-router";

export const routes: Array<Route> = [
	{
		path: rel.search,
		getComponent: (location, callback) => {
			require.ensure([], (require) => {
				callback(null, {
					main: require("sourcegraph/search/GlobalSearchMain").default,
				});
			});
		},
	},
];

// urlToSearch returns the URL to a search results page, or
// (if query is falsey) the search form page.
export function urlToSearch(query?: ?string): string {
	if (!query) return `/${rel.search}`;
	return `/${rel.search}?q=${encodeURIComponent(query)}`;
}

// locationForSearch returns the Location for the given query. It is different
// from the urlToSearch because this function's behavior is dependent on the current
// location and the updatePath/updateQuery parameters.
//
// This function modifies loc.
export function locationForSearch(loc: RouterLocation, query: ?string, updatePath: bool, updateQuery: bool): RouterLocation {
	if (updatePath) {
		loc.pathname = `/${rel.search}`;
	}

	if (!updateQuery) {
		if (query && !loc.state) loc.state = {};
		if (query) loc.state.q = query;
		else delete loc.state.q;
		return loc;
	}

	if (query && !loc.query) loc.query = {};
	if (query) loc.query.q = query;
	else delete loc.query.q;
	if (loc.state) delete loc.state.q;

	return loc;
}

function firstQueryValue(v: string | string[]): string {
	return typeof v === "string" ? v : v[0];
}

export function queryFromStateOrURL(loc: RouterLocation): ?string {
	if (loc.state && loc.state.hasOwnProperty("q")) return loc.state.q;
	else if (loc.query && loc.query.hasOwnProperty("q")) return firstQueryValue(loc.query.q);
	return null;
}
