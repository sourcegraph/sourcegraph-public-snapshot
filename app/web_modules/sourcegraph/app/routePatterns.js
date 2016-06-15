// @flow

import type {Route} from "react-router";
import {matchPattern} from "react-router/lib/PatternUtils";

export type RouteName = "styleguide" |
	"search" |
	"home" |
	"tools" |
	"tool" |
	"myRepos" |
	"tour" |
	"commit" |
	"def" |
	"defInfo" |
	"repo" |
	"tree" |
	"blob" |
	"build" |
	"builds" |
	"login" |
	"signup" |
	"forgot" |
	"reset" |
	"about" |
	"contact" |
	"security" |
	"pricing" |
	"terms" |
	"privacy" |
	"admin" |
	"adminBuilds" |
	"coverage" |
	"adminCoverage";

// NOTE: If you add a top-level route (e.g., "/tour"), add it to the
// topLevel list in app/internal/ui/router.go.
export const rel: {[key: RouteName]: string} = {
	search: "search",
	about: "about",
	contact: "contact",
	security: "security",
	pricing: "pricing",
	terms: "-/terms",
	privacy: "-/privacy",
	styleguide: "styleguide",
	home: "",
	tools: "tools",
	tool: "tools/*",
	myRepos: "repositories",
	tour: "tour",
	login: "login",
	signup: "join",
	forgot: "forgot",
	reset: "reset",
	admin: "-/",
	commit: "commit",
	def: "def/*",
	defInfo: "info/*",
	repo: "*", // matches both "repo" and "repo@rev"
	tree: "tree/*",
	blob: "blob/*",
	build: "builds/:id",
	builds: "builds",
	coverage: "coverage",
};

export const abs: {[key: RouteName]: string} = {
	search: rel.search,
	about: rel.about,
	contact: rel.contact,
	security: rel.security,
	pricing: rel.pricing,
	terms: rel.terms,
	privacy: rel.privacy,
	styleguide: rel.styleguide,
	home: rel.home,
	tools: rel.tools,
	tool: rel.tool,
	myRepos: rel.myRepos,
	tour: rel.tour,
	login: rel.login,
	signup: rel.signup,
	forgot: rel.forgot,
	reset: rel.reset,
	admin: rel.admin,
	adminBuilds: `${rel.admin}${rel.builds}`,
	adminCoverage: `${rel.admin}${rel.coverage}`,
	commit: `${rel.repo}/-/${rel.commit}`,
	def: `${rel.repo}/-/${rel.def}`,
	defInfo: `${rel.repo}/-/${rel.defInfo}`,
	repo: rel.repo,
	tree: `${rel.repo}/-/${rel.tree}`,
	blob: `${rel.repo}/-/${rel.blob}`,
	build: `${rel.repo}/-/${rel.build}`,
	builds: `${rel.repo}/-/${rel.builds}`,
};

const routeNamesByPattern: {[key: string]: RouteName} = {};
// $FlowHack
for (let name: RouteName of Object.keys(abs)) {
	routeNamesByPattern[abs[name]] = name;
}

export function getRoutePattern(routes: Array<Route>): string {
	return routes.map((route) => route.path).join("").slice(1); // remove leading '/''
}

export function getRouteName(routes: Array<Route>): ?string {
	return routeNamesByPattern[getRoutePattern(routes)];
}

export function getViewName(routes: Array<Route>): ?string {
	let name = getRouteName(routes);
	if (name) {
		return `View${name.charAt(0).toUpperCase()}${name.slice(1)}`;
	}
	return null;
}

export function getRouteParams(pattern: string, pathname: string): ?{[key: string]: string | string[]} {
	if (pathname.charAt(0) !== "/") pathname = `/${pathname}`;
	const {paramNames, paramValues} = matchPattern(pattern, pathname);

	if (paramValues !== null) {
		return paramNames.reduce((memo, paramName, index) => {
			if (typeof memo[paramName] === "undefined") {
				memo[paramName] = paramValues[index];
			} else if (typeof memo[paramName] === "string") {
				memo[paramName] = [memo[paramName], paramValues[index]];
			} else {
				memo[paramName].push(paramValues[index]);
			}
			return memo;
		}, {});
	}

	return null;
}
