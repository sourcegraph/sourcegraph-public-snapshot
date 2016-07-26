import {rel} from "sourcegraph/app/routePatterns";
import type {Route} from "react-router";
import {inBeta} from "sourcegraph/user";
import * as betautil from "sourcegraph/util/betautil";


export const desktopHome = {
	getComponent: (location, callback) => {
		require.ensure([], (require) => {
			callback(null, {
				main: require("sourcegraph/desktop/DesktopHome").default,
			});
		});
	},
};

export const routes: Array<Route> = [
	{
		...desktopHome,
		path: rel.desktopHome,
	},
];

export function inDesktopBeta(user) {
	return user && user.Betas && inBeta(user, betautil.DESKTOP);
}

export function redirectDesktopClient(router) {
	if (navigator.userAgent.includes("Electron")) {
		router.replace("/desktop/home");
	}
}
