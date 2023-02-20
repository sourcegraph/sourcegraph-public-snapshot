import { matchPath, useLocation } from 'react-router-dom-v5-compat'

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
            matchPath(route.path, location.pathname) || matchPath(route.path.replace(/\/\*$/, ''), location.pathname)
    )?.path
}
