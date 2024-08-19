import { browser, dev } from '$app/environment'

import { svelteKitRoutes, type SvelteKitRoute } from './routes'

let knownRoutesRegex: RegExp | undefined

// Additional known routes that are not in the list of known routes provided by the server.
// Having a separate list is error-prone and should be avoided if possible.
const additionalKnownRoutes: string[] = [
    // Cody's marketing page
    '^/cody/?$',
]

function getKnownRoutesRegex(): RegExp {
    if (!knownRoutesRegex) {
        const knownRoutes = additionalKnownRoutes.concat((browser && window.context?.svelteKit?.knownRoutes) || [])
        knownRoutesRegex = knownRoutes.length === 0 ? /$^/ : new RegExp(`${knownRoutes.join('|')}`)
    }
    return knownRoutesRegex
}

/**
 * Returns true if the given pathname is a known sub page.
 * This depends on the list of known routes provided by the server.
 */
export function isKnownSubPage(pathname: string): boolean {
    return getKnownRoutesRegex().test(pathname)
}

/**
 * Returns whether the SvelteKit app is enabled for the given route ID.
 * If not the caller should trigger a page reload to load the React app.
 * The enabled routes are provided by the server via `window.context`.
 *
 * Callers should pass an actual route ID retrived from SvelteKit not an
 * arbitrary path.
 *
 * NOTE: When in dev or preview mode, all routes are always enabled.
 *
 * @param pathname The pathname of the route to check.
 * @returns Whether the SvelteKit app is enabled for the given route.
 */
export function isRouteEnabled(pathname: string): boolean {
    // Preview mode is like production mode but with all routes enabled.
    // This necessary for running `pnpm run preview` or playwright tests.
    if (dev || import.meta.env.MODE === 'preview') {
        return true
    }

    if (!pathname) {
        return false
    }
    const enabledRoutes = window.context?.svelteKit?.enabledRoutes ?? []

    let foundRoute: SvelteKitRoute | undefined

    for (const routeIndex of enabledRoutes) {
        const route = svelteKitRoutes.at(routeIndex)
        if (route && route.pattern.test(pathname)) {
            foundRoute = route
            if (!route.isRepoRoot) {
                break
            }
            // If the found route is the repo root we have to keep going
            // to find a more specific route.
        }
    }

    if (foundRoute) {
        if (foundRoute.isRepoRoot) {
            // Check known routes to see if there is a more specific route than the repo root.
            // If yes then we should load the React app (if the more specific route was enabled
            // it would have been found above).
            return !isKnownSubPage(pathname)
        }
        return true
    }

    return false
}

/**
 * Helper function to determine whether a route is a repository route.
 * Callers can get the current route ID from the `page` store.
 */
export function isRepoRoute(routeID: string | null): boolean {
    if (!routeID) {
        return false
    }
    return routeID.startsWith('/[...repo=reporev]')
}
