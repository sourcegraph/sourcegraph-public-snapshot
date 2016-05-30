// @flow

import type {RouterState} from "react-router";

// shouldUpdateScroll makes it so that the window does NOT scroll
// if the previous and next routes have an identical, non-empty
// field named keepScrollPositionOnRouteChangeKey. This lets us
// achieve the desirable behavior of the window not scrolling when
// the user clicks on a def (and goes to a def route) on the blob page,
// or performs an incremental search that updates the URL.
export default function shouldUpdateScroll(prevRouterProps: ?RouterState, nextRouterProps: ?RouterState): bool {
	if (!prevRouterProps) return true;
	if (!nextRouterProps) return true;

	const prevRoute = prevRouterProps.routes[prevRouterProps.routes.length - 1];
	const nextRoute = nextRouterProps.routes[nextRouterProps.routes.length - 1];
	const changedScrollKey = !prevRoute.keepScrollPositionOnRouteChangeKey || prevRoute.keepScrollPositionOnRouteChangeKey !== nextRoute.keepScrollPositionOnRouteChangeKey;
	if (!changedScrollKey) return false;
	return true;
}
