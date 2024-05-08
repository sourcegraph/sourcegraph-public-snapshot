import { type RepoFile, encodeRepoRevision, parseBrowserRepoURL } from '@sourcegraph/shared/src/util/url'

export { parseBrowserRepoURL }

export function toTreeURL(target: RepoFile): string {
    return `/${encodeRepoRevision(target)}/-/tree/${target.filePath}`
}

/**
 * Replaces the revision in the given URL, or adds one if there is not already
 * one.
 *
 * @param href The URL whose revision should be replaced.
 */
export function replaceRevisionInURL(href: string, newRevision: string): string {
    const parsed = parseBrowserRepoURL(href)
    const repoRevision = `/${encodeRepoRevision(parsed)}`

    const url = new URL(href, window.location.href)
    url.pathname = `/${encodeRepoRevision({ ...parsed, revision: newRevision })}${url.pathname.slice(
        repoRevision.length
    )}`
    return `${url.pathname}${url.search}${url.hash}`
}

/**
 * Returns a URL to a file at a specific commit.
 */
export function getURLToFileCommit(href: string, filename: string, revision: string): string {
    const parsed = parseBrowserRepoURL(href)
    parsed.revision = revision
    parsed.filePath = '/-/blob/' + filename

    const url = new URL(href, window.location.href)
    return `/${parsed.repoName}@${parsed.revision}${parsed.filePath}${url.search}${url.hash}`
}
