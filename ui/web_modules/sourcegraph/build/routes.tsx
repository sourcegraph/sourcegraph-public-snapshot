// tslint:disable

import {Route} from "react-router";
import {rel} from "sourcegraph/app/routePatterns";
import {urlTo} from "sourcegraph/util/urlTo";
import {BuildsList} from "sourcegraph/build/BuildsList";
import {BuildContainer} from "sourcegraph/build/BuildContainer";

export const routes: ReactRouter.PlainRoute[] = [
	{
		onEnter: (nextState, replace) => {
			if (nextState.location.search === "") {
				replace(`${nextState.location.pathname}?filter=all`);
			}
		},
		getComponents: (location, callback) => {
			callback(null, {
				main: BuildsList,
			});
		},
		path: rel.builds,
	},
	{
		getComponents: (location, callback) => {
			callback(null, {
				main: BuildContainer,
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
