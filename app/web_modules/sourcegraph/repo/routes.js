// @flow

import {rel} from "sourcegraph/app/routePatterns";
import urlTo from "sourcegraph/util/urlTo";
import {makeRepoRev} from "sourcegraph/repo";
import type {Route} from "react-router";

let _components;

const common = {
	getComponents: (location, callback) => {
		require.ensure([], (require) => {
			if (!_components) {
				const withResolvedRepoRev = require("sourcegraph/repo/withResolvedRepoRev").default;
				_components = {
					navContext: withResolvedRepoRev(require("sourcegraph/repo/NavContext").default),
					main: withResolvedRepoRev(require("sourcegraph/repo/RepoMain").default),
				};
			}
			callback(null, _components);
		});
	},
};

// routes are the 2 routes needed for repos: the first is the one for repo
// subroutes, which must take precedence because the repo route matches
// greedily.
export const routes: Array<Route> = [
	{
		...common,
		path: `${rel.repo}/-/`,
		getChildRoutes: (location, callback) => {
			require.ensure([], (require) => {
				callback(null, [
					...require("sourcegraph/blob/routes").routes,
					...require("sourcegraph/build/routes").routes,
					...require("sourcegraph/def/routes").routes,
					...require("sourcegraph/tree/routes").routes,
				]);
			});
		},
	},
	{
		...common,
		path: rel.repo,
		disableTreeSearchOverlay: true,
		indexRoute: {
			getComponents: (location, callback) => {
				require.ensure([], (require) => {
					require("sourcegraph/tree/routes").routes[0].getComponents(location, callback);
				});
			},
		},
	},
];

export function urlToRepo(repo: string): string {
	return urlTo("repo", {splat: repo});
}

export function urlToRepoRev(repo: string, rev: string): string {
	return urlTo("repo", {splat: makeRepoRev(repo, rev)});
}
