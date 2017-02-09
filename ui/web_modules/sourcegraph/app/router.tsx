import { Location as HistoryLocation } from "history";
import { InjectedRouter, RouterState } from "react-router";

import { IRange } from "vs/editor/common/editorCommon";

import { abs, getRoutePattern } from "sourcegraph/app/routePatterns";
import { RangeOrPosition } from "sourcegraph/core/rangeOrPosition";
import { repoPath, repoRev } from "sourcegraph/repo";

export type RouterLocation = Location & HistoryLocation;
export type Router = InjectedRouter & RouterState & { params: RouteParams, location: RouterLocation };

export interface LocationProps {
	location: RouterLocation;
};

/**
 * React components may use RouterContext as the type of context received w/
 * the following declaration:
 *
 *     static contextTypes: React.ValidationMap<any> = {
 *         router: React.PropTypes.object.isRequired,
 *     };
 *
 *     context: RouterContext;
 */
export interface RouterContext {
	router: Router;
}

type Splat = string | string[];

/**
 * RouteParams includes the matched wildcard (*) arguments from a route pattern.
 */
export interface RouteParams {
	// splat is the matched wildcard character(s) of the route's pattern
	// e.g. "*/-/*" matched on "hello/-/world" will produce ["hello", "world"]
	splat: Splat;
}

let r: Router; // global singleton, use with care

/**
 * Global singleton instance of React Router, set by the application root.
 *
 * To receive an injected router in the rest of the React application, use the explicit
 * React contextTypes syntax.
 */
export function setRouter(router: Router): void {
	if (r) {
		throw new Error("illegal invocation to setRouter when router has already been defined");
	}
	r = router;
}

/**
 * Returns the application router for components that have been wrapped in
 * workbench/RouterContext. These components are split from the React application hierarchy
 * by the VS Code workbench, but we want them to operate with the same router.
 *
 * THIS SHOULD ONLY BE USED WITHIN WORKBENCH CONTEXT.
 */
export function __getRouterForWorkbenchOnly(): Router {
	if (!r) {
		throw new Error("application router hasn't been set");
	}
	return r;
}

/**
 * repoRevFromSplat returns the "repo[@rev]" string from the "splat" parameter.
 */
export function repoRevFromRouteParams(params: RouteParams): string {
	const { splat } = params;
	return splat instanceof Array ? splat[0] : splat;
}

/**
 * repoFromSplat returns the repo URI from the "splat" parameter.
 */
export function repoFromRouteParams(params: RouteParams): string {
	return repoPath(repoRevFromRouteParams(params));
}

/**
 * revFromSplat returns the repo URI from the "splat" parameter, if defined.
 */
export function revFromRouteParams(params: RouteParams): string | null {
	return repoRev(repoRevFromRouteParams(params));
}

/**
 * pathFromSplat returns the blob path from the "splat" parameter, if defined.
 */
export function pathFromRouteParams(params: RouteParams): string {
	const { splat } = params;
	const path = splat instanceof Array ? splat[1] : "";
	return path === "" ? "/" : path;
}

/**
 * getRepoFromRouter returns repo URI from the URL, if defined.
 */
export function getRepoFromRouter(router: Router): string | null {
	const routePattern = getRoutePattern(router.routes);
	switch (routePattern) {
		case abs.repo:
		case abs.tree:
		case abs.blob:
			return repoFromRouteParams(router.params);
	}
	return null;
}

/**
 * getRevFromRouter returns revision from the URL, if defined, or null for HEAD.
 */
export function getRevFromRouter(router: Router): string | null {
	const routePattern = getRoutePattern(router.routes);
	switch (routePattern) {
		case abs.repo:
		case abs.tree:
		case abs.blob:
			return revFromRouteParams(router.params);
	}
	return null;
}

/**
 * getPathFromRouter returns blob path from the URL, if defined.
 */
export function getPathFromRouter(router: Router): string | null {
	const routePattern = getRoutePattern(router.routes);
	switch (routePattern) {
		case abs.tree:
		case abs.blob:
			return pathFromRouteParams(router.params);
	}
	return null;
}
/**
 * BlobRouteProps returns the matched route arguments for blob URLs.
 */
export interface BlobRouteProps {
	repo: string;
	rev: string | null;
	path: string;
	selection: IRange | null;
}

/**
 * getBlobPropsFromRouter returns repo, (rev), path, and selection from the URL;
 * throws an exception if any required props are missing.
 */
export function getBlobPropsFromRouter(router: Router): BlobRouteProps {
	if (getRoutePattern(router.routes) !== abs.blob) {
		throw new Error("not a blob route");
	}

	const repo = getRepoFromRouter(router);
	if (!repo) {
		throw new Error("missing repo");
	}
	const rev = getRevFromRouter(router) || null;
	const path = getPathFromRouter(router);
	if (!path) {
		throw new Error("missing path");
	}

	return { repo, rev, path, selection: getSelectionFromRouter(router) };
}

/**
 * getSelectionFromRouter returns selection of the blob view from the URL hash part.
 */
export function getSelectionFromRouter(router: Router): IRange | null {
	const location = router.location;
	if (location.hash && location.hash.startsWith("#L")) {
		const rop = RangeOrPosition.parse(location.hash.replace(/^#L/, ""));
		if (rop) {
			return rop.toMonacoRangeAllowEmpty();
		}
	}
	return null;
}
