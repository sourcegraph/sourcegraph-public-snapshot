import { PlainRoute } from "react-router";
import { matchPattern } from "react-router/lib/PatternUtils";
import { Location } from "sourcegraph/Location";

export type RouteName = "styleguide" |
	"home" |
	"integrations" |
	"tool" |
	"settings" |
	"commit" |
	"repo" |
	"tree" |
	"blob" |
	"login" |
	"signup" |
	"forgot" |
	"reset" |
	"about" |
	"plan" |
	"beta" |
	"docs" |
	"contact" |
	"security" |
	"pricing" |
	"terms" |
	"privacy";

export const rel = {
	// NOTE: If you add a top-level route (e.g., "/about"), add it to the
	// topLevel list in app/internal/ui/router.go.
	about: "about",
	plan: "plan",
	beta: "beta",
	docs: "docs",
	contact: "contact",
	security: "security",
	pricing: "pricing",
	terms: "terms",
	privacy: "privacy",
	styleguide: "styleguide",
	integrations: "integrations",
	settings: "settings",
	login: "login",
	signup: "join",
	authext: "authext",

	home: "",

	repo: "*", // matches both "repo" and "repo@rev"
	commit: "commit",
	tree: "tree/*",
	blob: "blob/*",
};

export const abs = {
	about: rel.about,
	plan: rel.plan,
	contact: rel.contact,
	docs: rel.docs,
	security: rel.security,
	pricing: rel.pricing,
	terms: rel.terms,
	privacy: rel.privacy,
	styleguide: rel.styleguide,
	home: rel.home,
	integrations: rel.integrations,
	settings: rel.settings,
	login: rel.login,
	signup: rel.signup,
	authext: "authext",

	repo: rel.repo,
	commit: `${rel.repo}/-/${rel.commit}`,
	tree: `${rel.repo}/-/${rel.tree}`,
	blob: `${rel.repo}/-/${rel.blob}`,
};

const routeNamesByPattern: { [key: string]: RouteName } = {};
for (let name of Object.keys(abs)) {
	routeNamesByPattern[abs[name]] = name as RouteName;
}

export function getRoutePattern(routes: PlainRoute[]): string {
	return routes.map((route) => route.path).join("").slice(1); // remove leading '/''
}

export function getRouteName(routes: PlainRoute[]): string | null {
	return routeNamesByPattern[getRoutePattern(routes)] || null;
}

export function getViewName(routes: PlainRoute[]): string | null {
	let name = getRouteName(routes);
	if (name) {
		return `View${name.charAt(0).toUpperCase()}${name.slice(1)}`;
	}
	return null;
}

export function getRouteParams(pattern: string, pathname: string): any {
	if (pathname.charAt(0) !== "/") { pathname = `/${pathname}`; }
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

export function isRootRoute(location: Location): boolean {
	return rel.home === location.pathname.slice(1);
};
