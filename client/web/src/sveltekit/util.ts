import { memoize } from 'lodash'
import { matchPath } from 'react-router-dom'

import { routes } from '../routes'
import { PageRoutes } from '../routes.constants'

// All the routes that are supported by SvelteKit.
// Also see `SVELTEKIT_SUPPORTED_REPO_PATHS` for the list of supported repo sub-pages.
const supportedRoutes = new Set([PageRoutes.Search, PageRoutes.RepoContainer])

// All the routes that are enabled for SvelteKit. This is used for a gradual rollout of SvelteKit.
// Should be a subset of `supportedRoutes`.
// Keep in sync with 'cmd/frontend/internal/app/ui/sveltekit.go' and 'client/web-sveltekit/src/lib/navigation.ts'
const rolledoutRoutes = new Set([PageRoutes.Search])

const SVELTEKIT_SUPPORTED_REPO_PATHS = /^\/.*?\/-\/(blob\/|tree\/|branches$|commit\/|commits$|stats$|tags$)/

function isRepoSubPage(href: string): boolean {
    return href.includes('/-/')
}

const getSvelteKitSupportedRoute = memoize((pathname: string): PageRoutes | null => {
    if (!pathname) {
        return null
    }
    const route = routes.find(
        route => route.path && (matchPath(route.path, pathname) || matchPath(route.path.replace(/\/\*$/, ''), pathname))
    )
    if (route?.path && supportedRoutes.has(route.path as PageRoutes)) {
        if (
            route.path !== PageRoutes.RepoContainer ||
            !isRepoSubPage(pathname) ||
            SVELTEKIT_SUPPORTED_REPO_PATHS.test(pathname)
        ) {
            return route.path as PageRoutes
        }
    }
    return null
})

/**
 * Returns true if the current route is supported (i.e. implemented) by SvelteKit.
 */
export function isSupportedRoute(pathname: string): boolean {
    return getSvelteKitSupportedRoute(pathname) !== null
}

/**
 * Returns true if the current route is enabled for SvelteKit.
 * This is used for a gradual rollout of SvelteKit.
 * This should be used togehther with the `web-next-enabled` feature flag.
 */
export function isRolledOutRoute(pathname: string): boolean {
    const route = getSvelteKitSupportedRoute(pathname)
    if (route) {
        return rolledoutRoutes.has(route)
    }
    return false
}

export function reload(): void {
    const url = new URL(window.location.href)
    url.searchParams.append('feat', 'web-next')
    window.location.href = url.toString()
}
