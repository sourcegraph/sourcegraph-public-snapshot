import type { ResolvedRevision } from '$lib/web'

export function navFromPath(path: string, repo: string, blobPage: boolean): [string, string][] {
    const parts = path.split('/')
    return parts
        .slice(0, -1)
        .map((part, index, all): [string, string] => [part, `/${repo}/-/tree/${all.slice(0, index + 1).join('/')}`])
        .concat([[parts[parts.length - 1], `/${repo}/-/${blobPage ? 'blob' : 'tree'}/${path}`]])
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
