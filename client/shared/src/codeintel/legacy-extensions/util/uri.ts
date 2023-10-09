/**
 * Extracts the components of a text document URI.
 *
 * @param url The text document URL.
 */
export function parseGitURI(url: string): { repo: string; commit: string; path: string } {
    const { hostname, pathname, search, hash } = new URL(url.replace(/^git:/, 'http:'))
    return {
        repo: hostname + decodeURIComponent(pathname),
        commit: decodeURIComponent(search.slice(1)),
        path: decodeURIComponent(hash.slice(1)),
    }
}
