/**
 * Extracts the components of a text document URI.
 *
 * @param url The text document URL.
 */
export function parseGitURI(url: string): { repo: string; commit: string; path: string } {
    // We are replacing the scheme because hostnames of URIs with custom schemes (e.g. git)
    // are not parsed out in Chrome and Firefox. We have a polyfill for the main web app
    // (see client/shared/src/polyfills/configure-core-js.ts) but that might not be used
    // in all apps.
    const { hostname, pathname, search, hash } = new URL(url.replace(/^git:/, 'http:'))
    return {
        repo: hostname + decodeURIComponent(pathname),
        commit: decodeURIComponent(search.slice(1)),
        path: decodeURIComponent(hash.slice(1)),
    }
}
