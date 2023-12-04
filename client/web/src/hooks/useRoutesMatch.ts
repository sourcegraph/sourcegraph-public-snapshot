import { type RouteObject, matchPath, useLocation } from 'react-router-dom'

/**
 * Match the current pathname against a list of route patterns
 *
 * @param routes List of routes to match against
 *
 * @returns A matching route pattern
 */
export const useRoutesMatch = (routes: RouteObject[]): string | undefined => {
    const location = useLocation()

    // TODO: Replace with useMatches once top-level <Router/> is V6
    return routes.find(
        route =>
            (route.path && matchPath(route.path, location.pathname)) ||
            (route.path && matchPath(route.path.replace(/\/\*$/, ''), location.pathname))
    )?.path
}
