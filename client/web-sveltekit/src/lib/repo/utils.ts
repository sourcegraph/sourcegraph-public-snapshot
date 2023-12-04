import { resolvePath } from '@sveltejs/kit'

import type { ResolvedRevision } from '$lib/repo/api/repo'

const TREE_ROUTE_ID = '/[...repo=reporev]/(validrev)/(code)/-/tree/[...path]'

/**
 * Returns a [segment, url] mapping for every segement in `path`.
 * The URL for the last segment is empty.
 *
 * Example:
 *   'foo/bar/baz' converts to
 *   [
 *     ['foo', '/<repo>/-/tree/foo'],
 *     ['bar', '/<repo>/-/tree/foo/bar'],
 *     ['baz', '/<repo>/-/tree/foo/bar/baz'],
 *   ]
 *
 */
export function navFromPath(path: string, repo: string): [string, string][] {
    const parts = path.split('/')
    return parts
        .slice(0, -1)
        .map((part, index, all): [string, string] => [
            part,
            resolvePath(TREE_ROUTE_ID, { repo, path: all.slice(0, index + 1).join('/') }),
        ])
        .concat([[parts.at(-1), '']])
}

export function getRevisionLabel(
    urlRevision: string | undefined,
    resolvedRevision: ResolvedRevision | null
): string | undefined {
    return (
        (urlRevision && urlRevision === resolvedRevision?.commitID
            ? resolvedRevision?.commitID.slice(0, 7)
            : urlRevision?.slice(0, 7)) || resolvedRevision?.defaultBranch
    )
}

export function getFileURL(repoURL: string, file: { canonicalURL: string }): string {
    // TODO: Find out whether there is a safer way to do this
    return repoURL + file.canonicalURL.slice(file.canonicalURL.indexOf('/-/'))
}
