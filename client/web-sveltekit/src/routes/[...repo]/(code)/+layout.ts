import { dirname } from 'path'

import { browser } from '$app/environment'
import { isErrorLike } from '$lib/common'
import { fetchSidebarFileTree } from '$lib/repo/api/tree'

import type { LayoutLoad } from './$types'

let getRootPath = (path: string) => path

// We keep state in the browser to load the tree entries of the "highest"
// path that we visisted.
if (browser) {
    let topTreePath: string | undefined

    getRootPath = (path: string) => {
        if (topTreePath && (topTreePath === '.' || path.startsWith(topTreePath))) {
            return topTreePath
        }
        return (topTreePath = path)
    }
}

export const load: LayoutLoad = ({ parent, params }) => {
    const parentPath = getRootPath(params.path ? dirname(params.path) : '.')

    return {
        parentPath,
        fileTree: {
            deferred: parent().then(({ resolvedRevision, repoName, revision = '' }) => {
                if (isErrorLike(resolvedRevision)) {
                    throw resolvedRevision
                }
                return fetchSidebarFileTree({
                    repoName,
                    commitID: resolvedRevision.commitID,
                    revision,
                    filePath: parentPath,
                })
            }),
        },
    }
}
