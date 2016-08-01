import {formatPattern} from "react-router/lib/PatternUtils";
import type {RouteParams} from "react-router";
import {abs} from "sourcegraph/app/routePatterns";
import type {RouteName} from "sourcegraph/app/routePatterns";

// urlTo produces the full URL, given a route and route parameters. The
// route names are defined in sourcegraph/app/routePatterns.
export default function urlTo(name: RouteName, params: RouteParams): string {
	return formatPattern(`/${abs[name]}`, params);
}

// urlToGitHubOAuth
export function urlToGitHubOAuth(scopes: ?string, returnTo: ?(string | Location)): string {
	scopes = scopes ? `scopes=${encodeURIComponent(scopes)}` : null;
	if (returnTo && typeof returnTo !== "string") {
		returnTo = `${returnTo.pathname}${returnTo.search}${returnTo.hash}`;
	}
	returnTo = returnTo ? `return-to=${encodeURIComponent(returnTo)}` : null;

	let q;
	if (scopes && returnTo) {
		q = `${scopes}&${returnTo}`;
	} else if (scopes) {
		q = scopes;
	} else if (returnTo) {
		q = returnTo;
	}
	return `/-/github-oauth/initiate${q ? `?${q}` : ""}`;
}
export const privateGitHubOAuthScopes = "read:org,repo,user:email";
