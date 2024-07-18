import { resolveRoute } from '$app/paths'
import { encodeURIPathComponent } from '$lib/common'

const TREE_ROUTE_ID = '/[...repo=reporev]/(validrev)/(code)/-/tree/[...path]'
const BLOB_ROUTE_ID = '/[...repo=reporev]/(validrev)/(code)/-/blob/[...path]'

export function pathHrefFactory({
    repoName,
    revision,
    fullPath,
    fullPathType,
}: {
    repoName: string
    revision: string | undefined
    fullPath: string
    fullPathType: 'blob' | 'tree'
}): (targetPath: string) => string {
    return (targetPath: string) =>
        resolveRoute(
            // If we are targeting the last item in the path, respect the passed-in type.
            // Otherwise, we know we are targeting a tree higher up in the path.
            fullPath === targetPath && fullPathType === 'blob' ? BLOB_ROUTE_ID : TREE_ROUTE_ID,
            {
                repo: revision ? `${repoName}@${revision}` : repoName,
                path: encodeURIPathComponent(targetPath),
            }
        )
}
