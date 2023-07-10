import { catchError } from 'rxjs/operators'

import { asError, isErrorLike, type ErrorLike } from '$lib/common'
import { fetchTreeEntries } from '$lib/repo/api/tree'

import type { PageLoad } from './$types'

export const load: PageLoad = ({ params, parent }) => ({
    treeEntries: {
        deferred: parent().then(({ resolvedRevision, revision, repoName }) =>
            !isErrorLike(resolvedRevision)
                ? fetchTreeEntries({
                      repoName,
                      commitID: resolvedRevision.commitID,
                      revision: revision ?? '',
                      filePath: params.path,
                      first: 1000,
                  })
                      .pipe(catchError((error): [ErrorLike] => [asError(error)]))
                      .toPromise()
                      .then(commit => (isErrorLike(commit) ? null : commit.tree))
                : null
        ),
    },
})
