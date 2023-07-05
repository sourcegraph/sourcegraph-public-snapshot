import { catchError } from 'rxjs/operators'

import { asError, isErrorLike, type ErrorLike } from '$lib/common'
import { fetchLastCommit } from '$lib/repo/api/history'
import { fetchTreeEntries } from '$lib/repo/api/tree'
import { asStore } from '$lib/utils'

import type { PageLoad } from './$types'

export const load: PageLoad = ({ params, parent }) => ({
    treeEntries: asStore(
        parent().then(({ resolvedRevision, revision, repoName }) =>
            resolvedRevision
                ? fetchTreeEntries({
                      repoName,
                      commitID: resolvedRevision.commitID,
                      revision: revision ?? '',
                      filePath: params.path,
                      first: 2500,
                  })
                      .pipe(catchError((error): [ErrorLike] => [asError(error)]))
                      .toPromise()
                      .then(result => (isErrorLike(result) ? null : result.tree))
                : null
        )
    ),
    deferred: {
        history: parent().then(({ resolvedRevision }) =>
            resolvedRevision ? fetchLastCommit(resolvedRevision.repo.id, resolvedRevision.commitID, params.path) : null
        ),
        readmeBlob: parent().then(({ deferred: { readmeBlob } }) => readmeBlob),
    },
})
