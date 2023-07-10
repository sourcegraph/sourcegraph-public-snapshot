import { resolvePath } from '@sveltejs/kit'

import type { ResolvedRevision } from '$lib/web'

const TREE_ROUTE_ID = '/[...repo]/(code)/-/tree/[...path]'

/**
 * Returns a [segment, url] mapping for every segement in `path`.
 * The URL for the last segment is empty.
 */
export function navFromPath(path: string, repo: string): [string, string][] {
    const parts = path.split('/')
    return parts
        .slice(0, -1)
        .map((part, index, all): [string, string] => [
            part,
            resolvePath(TREE_ROUTE_ID, { repo, path: all.slice(0, index + 1).join('/') }),
        ])
        .concat([[parts[parts.length - 1], '']])
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
