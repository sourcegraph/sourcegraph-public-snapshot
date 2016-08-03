import type {Route} from "react-router";
import {rel} from "sourcegraph/app/routePatterns";
import urlTo from "sourcegraph/util/urlTo";

export const routes: Array<Route> = [
	{
		onEnter: (nextState, replace) => {
			if (nextState.location.search === "") {
				replace(`${nextState.location.pathname}?filter=all`);
			}
		},
		getComponents: (location, callback) => {
			require.ensure([], (require) => {
				callback(null, {
					main: require("sourcegraph/build/BuildsList").default,
				});
			});
		},
		path: rel.builds,
	},
	{
		getComponents: (location, callback) => {
			require.ensure([], (require) => {
				callback(null, {
					main: require("sourcegraph/build/BuildContainer").default,
				});
			});
		},
		path: rel.build,
	},
];

export function urlToBuild(repo: string, buildId: number): string {
	return urlTo("build", {splat: repo, id: buildId.toString()});
}

export function urlToBuilds(repo: string): string {
	return urlTo("builds", {splat: repo});
}
