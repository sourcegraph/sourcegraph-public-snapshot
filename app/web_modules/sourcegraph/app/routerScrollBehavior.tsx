// tslint:disable

import {RouterState} from "react-router";

// shouldUpdateScroll makes it so that the window does NOT scroll
// if the previous and next routes have an identical, non-empty
// field named keepScrollPositionOnRouteChangeKey. This lets us
// achieve the desirable behavior of the window not scrolling when
// the user clicks on a def (and goes to a def route) on the blob page,
// or performs an incremental search that updates the URL.
export function shouldUpdateScroll(prevRouterProps: RouterState | null, nextRouterProps: RouterState | null): boolean {
	if (!prevRouterProps) return true;
	if (!nextRouterProps) return true;

	const prevRoute = prevRouterProps.routes[prevRouterProps.routes.length - 1];
	const nextRoute = nextRouterProps.routes[nextRouterProps.routes.length - 1];
	const changedScrollKey = !(prevRoute as any).keepScrollPositionOnRouteChangeKey || (prevRoute as any).keepScrollPositionOnRouteChangeKey !== (nextRoute as any).keepScrollPositionOnRouteChangeKey;
	if (!changedScrollKey) return false;
	return true;
}

// Work around for react-router hash scroll behavior - https://github.com/reactjs/react-router/issues/394
export function hashLinkScroll() {
	const {hash} = window.location;
	if (hash !== "") {
		// Push onto callback queue so it runs after the DOM is updated,
		// this is required when navigating from a different page so that
		// the element is rendered on the page before trying to getElementById.
		setTimeout(() => {
			const id = hash.replace("#", "");
			const element = document.getElementById(id);
			if (element) element.scrollIntoView();
		}, 0);
	}
}
