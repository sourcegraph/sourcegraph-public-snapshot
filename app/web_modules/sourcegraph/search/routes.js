import {rel} from "sourcegraph/app/routePatterns";
import type {Route, RouterLocation} from "react-router";
import {searchScopes} from "sourcegraph/search";

export const routes: Array<Route> = [
	{
		path: rel.search,
		getComponent: (location, callback) => {
			require.ensure([], (require) => {
				callback(null, {
					main: require("sourcegraph/search/GlobalSearchMain").default,
					navContext: null,
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
// @NOTE `loc` cannot have immutable properties. `lang` and `scope` might be immutable.
// Ensure that if you assign new properties to loc that the properties are mutable.
export function locationForSearch(loc: RouterLocation, query: ?string, lang: ?string[], scope: ?Object, updatePath: bool, updateQuery: bool): RouterLocation {
	if (updatePath) {
		loc.pathname = `/${rel.search}`;
	}

	if (!lang) lang = [];

	if (!updateQuery) {
		loc.state = updateScopeAndLanguage(loc.state, scope, lang);
		if (query && !loc.state) loc.state = {};
		if (query) loc.state.q = query;
		else delete loc.state.q;
		return loc;
	}

	loc.query = updateScopeAndLanguage(loc.query, scope, lang);
	if (query && !loc.query) loc.query = {};
	if (query) loc.query.q = query;
	else delete loc.query.q;
	if (loc.state) delete loc.state.q;
	return loc;
}

function updateScopeAndLanguage(oldState: any, scope, lang) {
	let state = Object.assign({}, oldState);
	searchScopes.map((scopeName) => {
		if (scope && scope[scopeName]) state[scopeName] = true;
		else delete state[scopeName];
	});
	// creating a mutable version of lang since loc requires all properties to be mutable
	state.lang = lang.slice();
	return state;
}

function firstQueryValue(v: string | string[]): string {
	return typeof v === "string" ? v : v[0];
}

export function queryFromStateOrURL(loc: RouterLocation): ?string {
	if (loc.state && loc.state.hasOwnProperty("q")) return loc.state.q;
	else if (loc.query && loc.query.hasOwnProperty("q")) return firstQueryValue(loc.query.q);
	return null;
}


export function langsFromStateOrURL(loc: RouterLocation): ?string[] {
	if (loc.state && loc.state.hasOwnProperty("lang")) return loc.state.lang;
	else if (loc.query && loc.query.hasOwnProperty("lang")) return typeof loc.query.lang === "string" ? [loc.query.lang] : loc.query.lang;
	return null;
}

export function scopeFromStateOrURL(loc: RouterLocation): ?Object {
	let scope = {};
	searchScopes.forEach((scopeName) => {
		if (loc.state && loc.state.hasOwnProperty(scopeName)) scope[scopeName] = loc.state.scopeName;
		else if (loc.query.hasOwnProperty(scopeName)) scope[scopeName] = loc.query.scopeName;
	});
	return scope;
}
