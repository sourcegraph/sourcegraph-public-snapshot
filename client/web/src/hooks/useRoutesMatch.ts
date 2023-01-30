import { matchPath, useLocation } from 'react-router'

import { LayoutRouteProps } from '../routes'

/**
 * Match the current pathname against a list of route patterns
 *
 * @param routes List of routes to match against
 *
 * @returns A matching route pattern
 */
export const useRoutesMatch = (routes: readonly LayoutRouteProps[]): string | undefined => {
    const location = useLocation()

    // TODO: Replace with useMatches once top-level <Router/> is V6
    return routes.find(
        route =>
            matchPath(location.pathname, { path: route.path, exact: true }) ||
            matchPath(location.pathname, { path: route.path.replace(/\/\*$/, ''), exact: true })
    )?.path
}
