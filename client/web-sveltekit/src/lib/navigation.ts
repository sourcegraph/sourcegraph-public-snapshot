import { routeMeta } from '$lib/routeMeta'

/**
 * Returns whether the SvelteKit app is enabled for the given route ID.
 * If not the caller should trigger a page reload to load the React app.
 * The enabled routes are provided by the server via `window.context`.
 *
 * Callers should pass an actual route ID retrived from SvelteKit not an
 * arbitrary path.
 */
export function isRouteEnabled(routeID: string): boolean {
    if (!routeID) {
        return false
    }

    const serverRouteName = routeMeta[routeID]?.serverRouteName
    if (!serverRouteName) {
        return false
    }

    return !!window.context?.svelteKit?.enabledRoutes.includes(serverRouteName)
}

/**
 * Helper function to determine whether a route is a repository route.
 * Callers can get the current route ID from the `page` store.
 */
export function isRepoRoute(routeID: string | null): boolean {
    if (!routeID) {
        return false
    }
    return routeMeta[routeID]?.isRepoRoute ?? false
}
