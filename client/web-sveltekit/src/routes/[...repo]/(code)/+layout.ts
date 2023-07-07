import { dirname } from 'path'

import { catchError } from 'rxjs/operators'

import { browser } from '$app/environment'
import { asError, isErrorLike, type ErrorLike } from '$lib/common'
import { fetchTreeEntries } from '$lib/repo/api/tree'

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
        commitWithTree: {
            deferred: parent().then(({ resolvedRevision, repoName, revision }) =>
                !isErrorLike(resolvedRevision)
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
            ),
        },
    }
}
