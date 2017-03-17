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
export function urlToOAuth(provider: oauthProvider, scopes: string | null, returnTo: string | RouterLocation | null, newUserReturnTo: string | RouterLocation | null): string {
	scopes = scopes ? `scopes=${encodeURIComponent(scopes)}` : null;
	const returnToStr = returnTo ? `return-to=${encodeURIComponent(createHrefWithHash(returnTo))}` : null;
	const newUserReturnToStr = newUserReturnTo ? `new-user-return-to=${encodeURIComponent(createHrefWithHash(newUserReturnTo))}` : null;

	let q;
	if (scopes && returnTo && newUserReturnTo) {
		q = `${scopes}&${returnToStr}&${newUserReturnToStr}`;
	} else if (scopes && returnTo) {
		q = `${scopes}&${returnToStr}`;
	} else if (scopes) {
		q = scopes;
	} else if (returnTo) {
		q = returnTo;
	}
	return `/-/${provider}-oauth/initiate${q ? `?${q}` : ""}`;
}
