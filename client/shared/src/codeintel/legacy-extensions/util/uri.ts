/**
 * Extracts the components of a text document URI.
 *
 * @param url The text document URL.
 */
export function parseGitURI({ hostname, pathname, search, hash }: URL): { repo: string; commit: string; path: string } {
    return {
        repo: hostname + decodeURIComponent(pathname),
        commit: decodeURIComponent(search.slice(1)),
        path: decodeURIComponent(hash.slice(1)),
    }
}
