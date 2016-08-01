import type {Route} from "react-router";
import {rel} from "sourcegraph/app/routePatterns";
import urlTo from "sourcegraph/util/urlTo";

const builds = {
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
};
const build = {
	getComponents: (location, callback) => {
		require.ensure([], (require) => {
			callback(null, {
				main: require("sourcegraph/build/BuildContainer").default,
			});
		});
	},
};

export const routes: Array<Route> = [
	{
		...builds,
		path: rel.builds,
	},
	{
		...build,
		path: rel.build,
	},
];

export function urlToBuild(repo: string, buildId: number): string {
	return urlTo("build", {splat: repo, id: buildId.toString()});
}

export function urlToBuilds(repo: string): string {
	return urlTo("builds", {splat: repo});
}
