import isAbsoluteUrl from 'is-absolute-url'
import { matchPath } from 'react-router-dom'

import { routes } from '../routes'
import { PageRoutes } from '../routes.constants'

const SVELTEKIT_SUPPORTED_REPO_PATHS = /^\/.*?\/-\/(blob|tree|branches|commit|commits|stats|tags)\//

export function isSvelteKitSupportedURL(href: string): boolean {
    if (!href || isAbsoluteUrl(href)) {
        return false
    }
    const route = routes.find(
        route =>
            (route.path && matchPath(route.path, href)) ||
            (route.path && matchPath(route.path.replace(/\/\*$/, ''), href))
    )
    return (
        route?.path === PageRoutes.Search ||
        (route?.path === PageRoutes.RepoContainer && SVELTEKIT_SUPPORTED_REPO_PATHS.test(href))
    )
}

export function reload(): void {
    const url = new URL(window.location.href)
    url.searchParams.append('feat', 'enable-sveltekit')
    window.location.href = url.toString()
}
