// @flow

import {matchPattern} from "react-router/lib/PatternUtils";

export type RouteName = "dashboard" | "def" | "defRefs" | "repo" | "tree" | "blob" | "build" | "builds" | "login" | "signup" | "forgot" | "reset" | "admin";

export const rel: {[key: RouteName]: string} = {
	dashboard: "",
	login: "login",
	signup: "join",
	forgot: "forgot",
	reset: "reset",
	admin: "-/",
	def: "def/*",
	defRefs: "refs",
	repo: "*", // matches both "repo" and "repo@rev"
	tree: "tree/*",
	blob: "blob/*",
	build: "builds/:id",
	builds: "builds",
};

export const abs: {[key: RouteName]: string} = {
	dashboard: rel.dashboard,
	login: rel.login,
	signup: rel.signup,
	forgot: rel.forgot,
	reset: rel.reset,
	admin: rel.admin,
	def: `${rel.repo}/-/${rel.def}`,
	defRefs: `${rel.repo}/-/${rel.def}/-/refs`,
	repo: rel.repo,
	tree: `${rel.repo}/-/${rel.tree}`,
	blob: `${rel.repo}/-/${rel.blob}`,
	build: `${rel.repo}/-/${rel.build}`,
	builds: `${rel.repo}/-/${rel.builds}`,
};

export function getRouteParams(pattern: string, pathname: string): ?{[key: string]: string | string[]} {
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
