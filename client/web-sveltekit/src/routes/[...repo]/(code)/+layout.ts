import { dirname } from 'path'

import { catchError } from 'rxjs/operators'

import { browser } from '$app/environment'
import { asError, isErrorLike, type ErrorLike } from '$lib/common'
import { fetchBlobPlaintext } from '$lib/loader/blob'
import { fetchTreeEntries } from '$lib/repo/api/tree'

import type { LayoutLoad } from './$types'

let topTreePath: string | null = null

export const load: LayoutLoad = ({ parent, params, route }) => {
    const fileTreePath = params.path ?? '.'
    let parentPath = params.path ? dirname(params.path) : fileTreePath

    // Always show the "highest" folder in the path that we have already loaded
    if (topTreePath && (topTreePath === '.' || parentPath.startsWith(topTreePath))) {
        parentPath = topTreePath
    }

    if (browser) {
        topTreePath = parentPath
    }

    const repoInfo = parent()
    const commitWithParentTree = repoInfo.then(({ resolvedRevision, repoName, revision }) =>
        resolvedRevision
            ? fetchTreeEntries({
                  repoName,
                  commitID: resolvedRevision.commitID,
                  revision: revision ?? '',
                  filePath: parentPath,
                  first: 2500,
              })
                  .pipe(catchError((error): [ErrorLike] => [asError(error)]))
                  .toPromise()
            : null
    )
    const commitWithTree = params.path
        ? repoInfo.then(({ resolvedRevision, repoName, revision }) =>
              resolvedRevision && route.id?.includes('/tree/')
                  ? fetchTreeEntries({
                        repoName,
                        commitID: resolvedRevision.commitID,
                        revision: revision ?? '',
                        filePath: fileTreePath,
                        first: 2500,
                    })
                        .pipe(catchError((error): [ErrorLike] => [asError(error)]))
                        .toPromise()
                  : null
          )
        : commitWithParentTree

    const readmeBlob = Promise.all([repoInfo, commitWithTree]).then(([{ repoName, revision }, treeEntries]) => {
        if (treeEntries && !isErrorLike(treeEntries)) {
            const entry = treeEntries?.tree?.entries.find(
                entry => !entry.isDirectory && /^readme(\..*)$/i.test(entry.name)
            )
            if (entry) {
                return fetchBlobPlaintext({
                    filePath: entry.path,
                    repoName,
                    revision: revision ?? '',
                })
                    .toPromise()
                    .then(blob => ({ name: entry.name, content: blob?.content, richHTML: blob?.richHTML }))
            }
        }
        return null
    })
    return {
        fileTreePath: parentPath,
        deferred: {
            commitWithTree: commitWithParentTree,
            readmeBlob,
        },
    }
}
