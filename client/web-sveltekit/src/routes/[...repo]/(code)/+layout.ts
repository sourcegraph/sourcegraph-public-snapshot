import { dirname } from 'path'

import { browser } from '$app/environment'
import { isErrorLike } from '$lib/common'
import { fetchSidebarFileTree } from '$lib/repo/api/tree'
import { parseRepoRevision } from '$lib/shared'

import type { LayoutLoad } from './$types'

// Signifies the path of the repository root
const REPO_ROOT = '.'

let getRootPath = (_repo: string, path: string) => path

// We keep state in the browser to load the tree entries of the "highest" directory that was visited.
if (browser) {
    let topTreePath: Record<string, string> = {}

    getRootPath = (repo: string, path: string) => {
        const treePath = topTreePath[repo]
        if (treePath && (treePath === REPO_ROOT || path.startsWith(treePath))) {
            return topTreePath[repo]
        }
        return (topTreePath[repo] = path)
    }
}

export const load: LayoutLoad = ({ parent, params }) => {
    const { repoName } = parseRepoRevision(params.repo)
    const parentPath = getRootPath(repoName, params.path ? dirname(params.path) : REPO_ROOT)

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
