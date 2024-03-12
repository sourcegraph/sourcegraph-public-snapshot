// SvelteKit is rolled out in two stages:
// - Routes listed here are enabled by default for everyone on S2 (via the `web-next-enabled` feature flag)
// - Other routes are only enabled for users with the `web-next` feature flag
const enabledRouteIDs = new RegExp(
    [
        // Add route IDs here that should be enabled
        // Keep in sync with 'cmd/frontend/internal/app/ui/sveltekit.go'
        '^/search',
    ].join('|')
)

/**
 * Returns whether the given route is enabled.
 */
export function isRouteEnabled(routeID: string): boolean {
    console.log(routeID, enabledRouteIDs.test(routeID))
    return enabledRouteIDs.test(routeID)
}

/**
 * Helper function to determine whether a route is a repository route.
 * Callers can get the current route ID from the `page` store.
 */
export function isRepoRoute(routeID: string | null): boolean {
    if (!routeID) {
        return false
    }
    return routeID.startsWith('/[...repo=reporev]/')
}
