import { Location } from "history";
import { InjectedRouter, RouterState } from "react-router";
import { abs, getRoutePattern } from "sourcegraph/app/routePatterns";
import { RangeOrPosition } from "sourcegraph/core/rangeOrPosition";
import { IRange } from "vs/editor/common/editorCommon";

export type Router = InjectedRouter & RouterState;

// Global singleton instance of React Router; should be initialized via setRouter()
// during the application bootstrap.
export let router: Router;
export function setRouter(r: Router): void {
	router = r;
}

// Listeners may subscribe to Router (i.e. URL) changes.
// This should *only* be used for components that must
// exist outside of the main application component hierarchy.
// The only such components are those embedded beneath the
// VS Code Workbench layer (<WorkbenchShell>).
type Listener = (router: Router) => void;

let listenerId = 0;
const listenerMap = new Map<number, Listener>();

export function addRouterListener(listener: Listener): number {
	listenerId += 1;
	listenerMap.set(listenerId, listener);
	return listenerId;
}

export function removeRouterListener(id: number): boolean {
	return listenerMap.delete(id);
}

export function routerUpdated(): void {
	listenerMap.forEach((listener) => listener(router));
}

// routerSplat extract the "splat" argument from React's router, if defined.
// It is an array with an entry matching each of the "*" globs of the
// matched router pattern.
function routerSplat(): string[] | undefined {
	if (router.params) {
		return router.params["splat"] as any; // TODO(john): the declaration files lie, sigh
	}
}

// repoFromSplat returns the repo URI from the "splat" argument, if defined.
// Assumes repoRev is the first arugment of the splat according to route definitions.
function repoFromSplat(): string | undefined {
	const splat = routerSplat();
	if (splat && splat.length > 0) {
		const repoRev = splat[0];
		return repoRev.split("@")[0];
	}
}

// revFromSplat returns the repo URI from the "splat" argument, if defined.
// Assumes repoRev is the first arugment of the splat according to route definitions.
function revFromSplat(): string | undefined {
	const splat = routerSplat();
	if (splat && splat.length > 0) {
		const repoRev = splat[0];
		return repoRev.split("@")[1];
	}
}

// pathFromSplat returns the blob path from the "splat" argument.
// Assumes path is the second arugment of the splat according to route definitions, after repoRev.
function pathFromSplat(): string | undefined {
	const splat = routerSplat();
	if (splat && splat.length > 1) {
		return splat[1];
	}
}

// getRepoFromRouter returns repo URI from the URL, if defined.
export function getRepoFromRouter(): string | undefined {
	switch (getRoutePattern(router.routes)) {
		case abs.repo:
		case abs.tree:
		case abs.blob:
			return repoFromSplat();
	}
}

// getRevFromRouter returns revision from the URL, if defined.
export function getRevFromRouter(): string | undefined {
	switch (getRoutePattern(router.routes)) {
		case abs.repo:
		case abs.tree:
		case abs.blob:
			return revFromSplat();
	}
}

// getPathFromRouter returns blob path from the URL
export function getPathFromRouter(): string | undefined {
	switch (getRoutePattern(router.routes)) {
		case abs.tree:
		case abs.blob:
			return pathFromSplat();
	}
}

export interface RouterBlobProps {
	repo: string;
	rev: string | null;
	path: string;
}

// getBlobPropsFromRouter returns repo, (rev), and path from the URL;
// throws an exception if any required props are missing.
export function getBlobPropsFromRouter(): RouterBlobProps {
	if (getRoutePattern(router.routes) !== abs.blob) {
		throw new Error("not a blob route");
	}

	const repo = getRepoFromRouter();
	if (!repo) {
		throw new Error("missing repo");
	}
	const rev = getRevFromRouter() || null;
	const path = getPathFromRouter();
	if (!path) {
		throw new Error("missing path");
	}

	return { repo, rev, path };
}

export function getSelectionFromRouter(): IRange {
	let rop = RangeOrPosition.parse(window.location.hash.substr(2));
	if (!rop) {
		rop = RangeOrPosition.fromOneIndexed(1);
	}
	return rop.toMonacoRangeAllowEmpty();
}
