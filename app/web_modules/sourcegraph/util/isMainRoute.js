// @flow

import type {Route} from "react-router";

// isMainRoute returns whether route is the deepest-matched route
// in routeStack. For example, a def route is [root, repo, def],
// and it's helpful to know if the component's this.props.route
// is the main route or something else.
export default function isMainRoute(route: Route, routeStack: Array<Route>): boolean {
	return !route || !routeStack || route === routeStack[routeStack.length - 1];
}
