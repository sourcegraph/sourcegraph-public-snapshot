import type { Reroute } from '@sveltejs/kit'

import { type SvelteKitRoute, svelteKitRoutes } from '$lib/routes'

export const reroute: Reroute = ({ url }) => {
    return rerouteWithEncodedFilePath(url)
}

/**
 * For SvelteKit to process repo URLs with file paths properly we have to encode the file path portion of the URL pathname.
 *
 * Because SvelteKit uses eager quantifiers in its route regular expressions, and because some repo routes contain two
 * rest parameters to match the repository and the filepath, a path such as
 * ```
 * /repo-name/-/blob/path/-/blob/file
 * ```
 * would be interpreted as
 * ```
 * /repo-name/-/blob/path/-/blob/file
 *  ^^^^^^^^^^^^^^^^^^^^^        ^^^^
 *       repo               path
 * ```
 * instead of the desired
 * ```
 * /repo-name/-/blob/path/-/blob/file
 *  ^^^^^^^^^        ^^^^^^^^^^^^^^^^
 *    repo                path
 * ```
 * By URL-encoding the file path portion of the URL pathname we avoid this behavior.
 * @param url The URL to navigate to
 */
function rerouteWithEncodedFilePath(url: URL): string | void {
    for (const route of ROUTES_WITH_FILEPATH) {
        if (route.pattern.test(url.pathname)) {
            return route.encodeFilePath(url)
        }
    }
}

interface SvelteKitRouteWithPathEncoder extends SvelteKitRoute {
    /**
     * Encodes the part of the URL pathname that represents a file path.
     * @param url
     */
    encodeFilePath(url: URL): string
}

/**
 * A list of routes for which the file path portion of the URL path name should be encoded.
 */
const ROUTES_WITH_FILEPATH: SvelteKitRouteWithPathEncoder[] = (function () {
    const filePathParameter = '[...path]'
    const filePathSegment = `/${filePathParameter}`
    const routesWithFilepath: SvelteKitRouteWithPathEncoder[] = []

    for (const route of svelteKitRoutes) {
        // To keep the logic simple we currently only do this for routes that have the file path parameter at
        // the end.
        if (route.id.endsWith(filePathSegment)) {
            const suffix = getStaticSuffix(route.id.split('/').slice(0, -1))
            const separator = suffix.length > 0 ? `/${suffix.join('/')}/` : '/'
            routesWithFilepath.push({
                ...route,
                encodeFilePath(url) {
                    const start = url.pathname.indexOf(separator)
                    if (start < 0) {
                        return url.pathname
                    }
                    const end = start + separator.length
                    return url.pathname.slice(0, end) + encodeURIComponent(url.pathname.slice(end))
                },
            })
        }
    }

    return routesWithFilepath
})()

/**
 * Returns all "static" route segments following from the last optional parameter till the end. Ignores groups since
 * they are not part of URL.s
 * Example:
 *     Input: ['s1', '[...rest]', 's2', '(group)', 's3', 's4']
 *     Output: ['s2', 's3', 's4']
 * @param routeSegments The route to compute the static prefix for
 */
function getStaticSuffix(routeSegments: string[]): string[] {
    const staticSegments: string[] = []

    for (let i = routeSegments.length - 1; i > -1; i--) {
        const segment = routeSegments[i]
        if (segment[0] === '(') {
            continue
        }
        if (segment[0] === '[') {
            break
        }
        staticSegments.unshift(segment)
    }
    return staticSegments
}

export const getStaticSuffix_TEST_ONLY = getStaticSuffix
