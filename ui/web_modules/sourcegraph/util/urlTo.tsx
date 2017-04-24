import { browserHistory as history } from "react-router";
import { formatPattern } from "react-router/lib/PatternUtils";

import { RouteName, abs } from "sourcegraph/app/routePatterns";
import { RouteParams } from "sourcegraph/app/router";
import { RouterLocation } from "sourcegraph/app/router";

// urlTo produces the full URL, given a route and route parameters. The
// route names are defined in sourcegraph/app/routePatterns.
export function urlTo(name: RouteName, params: RouteParams): string {
	return formatPattern(`/${abs[name]}`, params);
}

export type oauthProvider = "github" | "google";

function createHrefWithHash(loc: RouterLocation | string): string {
	if (typeof loc === "string") {
		if (loc.indexOf("#") !== -1) {
			throw new Error(`pathname ${JSON.stringify(loc)} must not contain '#' (use {pathname: '/foo/bar', hash: 'baz'})`);
		}
		loc = { pathname: loc } as RouterLocation;
	}
	return history.createHref(loc);
}

// urlToOAuth returns an OAuth initiate URL for given provider, scopes, returnTo.
export function urlToOAuth(provider: oauthProvider, scopes: string, returnTo: string | RouterLocation, newUserReturnTo: string | RouterLocation, webSessionId?: string): string {
	const q = [
		`scopes=${encodeURIComponent(scopes)}`,
		`return-to=${encodeURIComponent(createHrefWithHash(returnTo))}`,
		`new-user-return-to=${encodeURIComponent(createHrefWithHash(newUserReturnTo))}`
	];
	if (webSessionId) {
		q.push(`web-session-id=${encodeURIComponent(webSessionId)}`);
	}
	return `/-/${provider}-oauth/initiate?${q.join("&")}`;
}
