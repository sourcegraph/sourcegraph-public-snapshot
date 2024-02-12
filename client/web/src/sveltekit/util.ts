import { matchPath } from 'react-router-dom'

import { routes } from '../routes'
import { PageRoutes } from '../routes.constants'

const SVELTEKIT_SUPPORTED_REPO_PATHS = /^\/.*?\/-\/(blob\/|tree\/|branches$|commit\/|commits$|stats$|tags$)/

function isRepoSubPage(href: string): boolean {
    return href.includes('/-/')
}

export function isSvelteKitSupportedURL(pathname: string): boolean {
    if (!pathname) {
        return false
    }
    const route = routes.find(
        route => route.path && (matchPath(route.path, pathname) || matchPath(route.path.replace(/\/\*$/, ''), pathname))
    )
    return (
        route?.path === PageRoutes.Search ||
        (route?.path === PageRoutes.RepoContainer &&
            (!isRepoSubPage(pathname) || SVELTEKIT_SUPPORTED_REPO_PATHS.test(pathname)))
    )
}

export function reload(): void {
    const url = new URL(window.location.href)
    url.searchParams.append('feat', 'web-next')
    window.location.href = url.toString()
}
