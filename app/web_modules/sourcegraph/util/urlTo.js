// @flow

import {formatPattern} from "react-router/lib/PatternUtils";
import type {RouteParams} from "react-router";
import {abs} from "sourcegraph/app/routePatterns";
import type {RouteName} from "sourcegraph/app/routePatterns";

// urlTo produces the full URL, given a route and route parameters. The
// route names are defined in sourcegraph/app/routePatterns.
export default function urlTo(name: RouteName, params: RouteParams): string {
	return formatPattern(`/${abs[name]}`, params);
}

export const urlToGitHubOAuth = "/-/github-oauth/initiate";
export const urlToPrivateGitHubOAuth = `${urlToGitHubOAuth}?scopes=read:org,repo,user:email`;
